package routes_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ZeFort/chance"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/embed"

	"github.com/storj/dms/pkg/routes"
	"github.com/storj/dms/pkg/storage"
)

var (
	c         *chance.Chance
	usedPorts []int
)

func TestPingdomEmpty(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:               etcD,
		HeartbeatExpiration: 5 * time.Second,
	}

	ech := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	assert.Nil(t, dmsRouter.Pingdom(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.Nil(t, etcD.Purge("env"))
}

func TestPingdomValid(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:               etcD,
		HeartbeatExpiration: 5 * time.Second,
	}

	ech := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	dmsRouter.Store.StoreCheckin("test", time.Now())
	assert.Nil(t, dmsRouter.Pingdom(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.Nil(t, etcD.Purge("env"))
}

func TestPingdomExpired(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:               etcD,
		HeartbeatExpiration: 5 * time.Second,
	}

	ech := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	last := time.Now().AddDate(0, -1, 0)
	dmsRouter.Store.StoreCheckin("test", last)
	assert.Nil(t, dmsRouter.Pingdom(c))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	incidents, err := dmsRouter.Store.AllIncidents()
	assert.Nil(t, err)
	d := time.Now().Format("2006-01-02")
	assert.Equal(t, 1, len(incidents[fmt.Sprintf("env/test/incidents/%s", d)]))

	assert.Nil(t, etcD.Purge("env"))
}

func TestRegister(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:     etcD,
		JWTSecret: "secret",
	}

	ech := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("environment=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	assert.Nil(t, dmsRouter.Register(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp)

	// validate returned token
	token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil
	})
	assert.Nil(t, err)

	// validate claims remained intact
	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, "test", claims["environment"])

	assert.Nil(t, etcD.Purge("env"))
}

func TestStatusEmpty(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:               etcD,
		HeartbeatExpiration: 5 * time.Second,
	}

	ech := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)

	assert.Nil(t, dmsRouter.Status(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]time.Time
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp))

	assert.Nil(t, etcD.Purge("env"))
}

func TestStatusNonEmpty(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:               etcD,
		HeartbeatExpiration: 5 * time.Second,
	}

	last := time.Now()
	etcD.StoreCheckin("test", last)

	ech := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)

	assert.Nil(t, dmsRouter.Status(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]time.Time
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp))
	val, ok := resp["env/test/last"]
	assert.True(t, ok)
	assert.Equal(t, last.UnixNano(), val.UnixNano())

	assert.Nil(t, etcD.Purge("env"))
}

func TestIncidentsEmpty(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:               etcD,
		HeartbeatExpiration: 5 * time.Second,
	}

	ech := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)

	assert.Nil(t, dmsRouter.Incidents(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp))

	assert.Nil(t, etcD.Purge("env"))
}

func TestIncidentsNonEmpty(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:               etcD,
		HeartbeatExpiration: 5 * time.Second,
	}

	incTime := time.Now()
	etcD.StoreIncident("test", "2021-06-21", []time.Time{incTime})

	ech := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)

	assert.Nil(t, dmsRouter.Incidents(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string][]string
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp))
	val, ok := resp["env/test/incidents/2021-06-21"]
	assert.True(t, ok)
	assert.Equal(t, 1, len(val))

	assert.Nil(t, etcD.Purge("env"))
}

func TestIngest(t *testing.T) {
	e, c1 := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{c1})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	dmsRouter := routes.DMSRouter{
		Store:     etcD,
		JWTSecret: "secret",
	}

	ech := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("environment=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	assert.Nil(t, dmsRouter.Register(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp)

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("environment=test"))
	req.Header.Set("Authorization", "Bearer "+resp.Token)

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	claims := jwt.MapClaims{}
	claims["environment"] = "test"
	c.Set("user", &jwt.Token{
		Claims: claims,
	})
	assert.Nil(t, dmsRouter.Ingest(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body)

	all, err := etcD.AllCheckins()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(all))
	assert.NotEmpty(t, all["env/test/last"])

	assert.Nil(t, etcD.Purge("env"))
}

func setup(t *testing.T) (*embed.Etcd, string) {
	// launch etcd server
	cfg := embed.NewConfig()
	cfg.Dir = "default.etcd"
	p1, p2 := getPorts()
	cfg.LPUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", p1)}}
	cfg.LCUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", p2)}}
	cfg.LogOutputs = []string{"stderr"}
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		log.Fatal(err)
		t.Fail()
	}
	select {
	case <-e.Server.ReadyNotify():
		log.Printf("Server is ready!")
	case <-time.After(30 * time.Second):
		e.Server.Stop() // trigger a shutdown
		log.Printf("Server took too long to start!")
		t.Fail()
	}
	connectionString := fmt.Sprintf("127.0.0.1:%d", p2)
	return e, connectionString
}

func getPorts() (int, int) {
	c = chance.New()
portPick:
	p1 := c.IntBtw(3100, 3200)
	p2 := c.IntBtw(3100, 3200)
	for _, b := range usedPorts {
		if b == p1 || b == p2 {
			goto portPick
		}
	}
	return p1, p2
}
