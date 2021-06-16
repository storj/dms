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

	assert.Nil(t, etcD.Purge())

	err = etcD.Store("env=foo", storage.Data{})
	assert.Nil(t, err)

	all, err := etcD.All()
	assert.Nil(t, err)
	assert.Equal(t, len(all), 1)

	assert.Nil(t, etcD.Purge())

	for i := 0; i < 5; i++ {
		err = etcD.Store(fmt.Sprintf("env=bar_%d", i), storage.Data{})
		assert.Nil(t, err)
	}

	all, err = etcD.All()
	assert.Nil(t, err)
	assert.Equal(t, len(all), 5)

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
