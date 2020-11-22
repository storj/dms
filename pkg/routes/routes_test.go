package routes_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/embed"

	"github.com/kristaxox/dms/pkg/routes"
	"github.com/kristaxox/dms/pkg/storage"
)

func TestPingdomRoute(t *testing.T) {
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

	// pingdom with empty data store
	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	assert.Nil(t, dmsRouter.Pingdom(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	// pingdom with valid heartbeat
	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	dmsRouter.Store.Store("test", storage.Data{Last: time.Now()})
	assert.Nil(t, dmsRouter.Pingdom(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.Nil(t, etcD.Purge())

	// pingdom with expired heartbeat
	rec = httptest.NewRecorder()
	c = ech.NewContext(req, rec)
	last := time.Now().AddDate(0, -1, 0)
	dmsRouter.Store.Store("test", storage.Data{Last: last})
	assert.Nil(t, dmsRouter.Pingdom(c))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	assert.Nil(t, etcD.Purge())
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
