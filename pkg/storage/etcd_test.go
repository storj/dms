package storage_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/embed"

	"github.com/storj/dms/pkg/storage"
)

func TestEtcdStorage(t *testing.T) {
	// launch etcd server
	e := setup(t)
	defer e.Close()

	etcD, err := storage.NewEtcdStorage([]string{"localhost:2379"})
	assert.Nil(t, err)
	assert.NotNil(t, etcD)

	assert.Nil(t, etcD.Purge("env"))

	checkinTime := time.Now()
	err = etcD.StoreCheckin("foo", checkinTime)
	assert.Nil(t, err)

	all, err := etcD.AllCheckins()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(all))

	incidentTime := time.Now()
	err = etcD.StoreIncident("foo", "2021-07-17", []time.Time{incidentTime})
	assert.Nil(t, err)

	allInc, err := etcD.AllIncidents()
	assert.Nil(t, err)
	log.Printf("%+v", allInc)
	assert.Equal(t, 1, len(allInc["env/foo/incidents/2021-07-17"]))

	assert.Nil(t, etcD.Purge("env"))

	for i := 0; i < 5; i++ {
		err = etcD.StoreCheckin(fmt.Sprintf("bar_%d", i), time.Now())
		assert.Nil(t, err)
	}

	all, err = etcD.AllCheckins()
	assert.Nil(t, err)
	assert.Equal(t, 5, len(all))

	assert.Nil(t, etcD.Purge("env"))
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
