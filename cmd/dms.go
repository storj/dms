package main

import (
	"crypto/subtle"
	"io"
	"log"
	"text/template"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/embed"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/kristaxox/dms/pkg/routes"
	"github.com/kristaxox/dms/pkg/storage"
)

var (
	basicAuthUsername = kingpin.Flag("basic-auth-username", "").Required().String()
	basicAuthPassword = kingpin.Flag("basic-auth-password", "").Required().String()
	jwtSecret         = kingpin.Flag("jwt-secret", "").Required().String()

	etcdEndpoints = kingpin.Flag("endpoints", "").Strings()

	heartbeatExpiration = kingpin.Flag("heartbeat-expiration", "").Default("10m").Duration()
)

type Template struct {
	templates *template.Template
}

func main() {
	kingpin.Parse()

	if len(*etcdEndpoints) == 0 {
		logrus.Info("using embeded etcd server")
		e := setup()
		defer e.Close()
		etcdEndpoints = &[]string{"localhost:2379"}
	}

	store, err := storage.NewEtcdStorage(*etcdEndpoints)
	if err != nil {
		panic(err)
	}

	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e := echo.New()
	e.Renderer = t

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// dms router
	dmsRouter := &routes.DMSRouter{
		Store:               store,
		JWTSecret:           *jwtSecret,
		HeartbeatExpiration: *heartbeatExpiration,
	}

	basicAuthMiddlewareFunc := func(username, password string, c echo.Context) (bool, error) {
		// Be careful to use constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(*basicAuthUsername)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(*basicAuthPassword)) == 1 {
			return true, nil
		}
		return false, nil
	}

	// Pingdom route (public)
	p := e.Group("/pingdom")
	p.GET("", dmsRouter.Pingdom)

	// Login route (basic auth)
	g := e.Group("/register")
	g.Use(middleware.BasicAuth(basicAuthMiddlewareFunc))
	g.GET("", dmsRouter.RegisterForm)
	g.POST("", dmsRouter.Register)

	// Status route (basic auth)
	i := e.Group("/status")
	i.Use(middleware.BasicAuth(basicAuthMiddlewareFunc))
	i.GET("", dmsRouter.Status)

	// Ingest route (jwt)
	r := e.Group("/ingest")
	r.Use(middleware.JWT([]byte(*jwtSecret)))
	r.POST("", dmsRouter.Ingest)

	e.Logger.Fatal(e.Start(":1323"))
}

func setup() *embed.Etcd {
	// launch etcd server
	cfg := embed.NewConfig()
	cfg.Dir = "default.etcd"
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		panic(err)
	}
	select {
	case <-e.Server.ReadyNotify():
		log.Printf("Server is ready!")
	case <-time.After(30 * time.Second):
		e.Server.Stop() // trigger a shutdown
	}
	return e
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
