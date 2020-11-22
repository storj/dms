package routes_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/embed"

	"github.com/kristaxox/dms/pkg/routes"
	"github.com/kristaxox/dms/pkg/storage"
)

func TestPingdomEmpty(t *testing.T) {
	e := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{"localhost:2379"})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge())

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

	assert.Nil(t, etcD.Purge())
}

func TestPingdomValid(t *testing.T) {
	e := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{"localhost:2379"})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge())

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
	dmsRouter.Store.Store("test", storage.Data{Last: time.Now()})
	assert.Nil(t, dmsRouter.Pingdom(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.Nil(t, etcD.Purge())
}

func TestPingdomExpired(t *testing.T) {
	e := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{"localhost:2379"})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge())

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
	dmsRouter.Store.Store("test", storage.Data{Last: last})
	assert.Nil(t, dmsRouter.Pingdom(c))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	assert.Nil(t, etcD.Purge())
}

func TestRegister(t *testing.T) {
	e := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{"localhost:2379"})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge())

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

	assert.Nil(t, etcD.Purge())
}

func TestStatusEmpty(t *testing.T) {
	e := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{"localhost:2379"})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge())

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

	var resp map[string]storage.Data
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp))
}

func TestStatusNonEmpty(t *testing.T) {
	e := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{"localhost:2379"})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge())

	dmsRouter := routes.DMSRouter{
		Store:               etcD,
		HeartbeatExpiration: 5 * time.Second,
	}

	last := time.Now()
	etcD.Store("test", storage.Data{Last: last})

	ech := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var rec *httptest.ResponseRecorder
	var c echo.Context

	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)

	assert.Nil(t, dmsRouter.Status(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]storage.Data
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp))
	val, ok := resp["env=test"]
	assert.True(t, ok)
	assert.Equal(t, last.UnixNano(), val.Last.UnixNano)
}

func TestIngest(t *testing.T) {
	e := setup(t)
	defer e.Close()
	etcD, err := storage.NewEtcdStorage([]string{"localhost:2379"})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge())

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

	all, err := etcD.All()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(all))
	assert.NotEmpty(t, all["env=test"])
}

func setup(t *testing.T) *embed.Etcd {
	// launch etcd server
	cfg := embed.NewConfig()
	cfg.Dir = "default.etcd"
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		log.Fatal(err)
	}
	select {
	case <-e.Server.ReadyNotify():
		log.Printf("Server is ready!")
	case <-time.After(30 * time.Second):
		e.Server.Stop() // trigger a shutdown
		log.Printf("Server took too long to start!")
		t.Fail()
	}
	return e
}
