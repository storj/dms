package main

import (
	"crypto/subtle"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/storj/dms/pkg/routes"
	"github.com/storj/dms/pkg/storage"
)

var (
	basicAuthUsername = kingpin.Flag("basic-auth-username", "").Required().String()
	basicAuthPassword = kingpin.Flag("basic-auth-password", "").Required().String()
	jwtSecret         = kingpin.Flag("jwt-secret", "").Required().String()

	etcdEndpoints = kingpin.Flag("endpoints", "").Required().Strings()

	heartbeatExpiration = kingpin.Flag("heartbeat-expiration", "").Default("10m").Duration()
)

func main() {
	kingpin.Parse()

	store, err := storage.NewEtcdStorage(*etcdEndpoints)
	if err != nil {
		panic(err)
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// dms router
	dmsRouter := &routes.DMSRouter{
		Store:               store,
		JWTSecret:           *jwtSecret,
		HeartbeatExpiration: *heartbeatExpiration,
	}

	// Pingdom route (public)
	p := e.Group("/pingdom")
	p.GET("", dmsRouter.Pingdom)

	basicAuthMiddlewareFunc := func(username, password string, c echo.Context) (bool, error) {
		// Be careful to use constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(*basicAuthUsername)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(*basicAuthPassword)) == 1 {
			return true, nil
		}
		return false, nil
	}

	// Login route (basic auth)
	g := e.Group("/register")
	g.Use(middleware.BasicAuth(basicAuthMiddlewareFunc))
	g.POST("", dmsRouter.Register)

	// Status route (basic auth)
	i := e.Group("/status")
	i.Use(middleware.BasicAuth(basicAuthMiddlewareFunc))
	i.GET("", dmsRouter.Status)

	k := e.Group("/incidents")
	k.Use(middleware.BasicAuth(basicAuthMiddlewareFunc))
	k.GET("", dmsRouter.Incidents)

	// Ingest route (jwt)
	r := e.Group("/ingest")
	r.Use(middleware.JWT([]byte(*jwtSecret)))
	r.POST("", dmsRouter.Ingest)

	e.Logger.Fatal(e.Start(":1323"))
}
