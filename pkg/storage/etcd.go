package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
)

var (
	dialTimeout    = 30 * time.Second
	requestTimeout = 10 * time.Second
)

// EtcdStorage is the StorageImpl interface implementation for using Etcd as the backing store
type EtcdStorage struct {
	kv clientv3.KV
}

// NewEtcdStorage instantiates the type and creates the etcd client
func NewEtcdStorage(endpoints []string) (*EtcdStorage, error) {
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   endpoints,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create etcd.v3 client")
	}
	kv := clientv3.NewKV(cli)
	return &EtcdStorage{
		kv: kv,
	}, nil
}

// Store stores the key
func (etcds *EtcdStorage) StoreCheckin(key string, value time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	encoded, err := json.Marshal(value)
	if err != nil {
		cancel()
		return errors.Wrap(err, "failed to store key")
	}
	key = strings.TrimSuffix(strings.TrimPrefix(key, "env/"), "/last") // sanitize the strings
	_, err = etcds.kv.Put(ctx, fmt.Sprintf("env/%s/last", key), string(encoded))
	cancel()
	return errors.Wrap(err, "failed to store key")
}

func (etcds *EtcdStorage) StoreIncident(key string, date string, value []time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	encoded, err := json.Marshal(value)
	if err != nil {
		cancel()
		return errors.Wrap(err, "failed to store key")
	}
	key = strings.TrimSuffix(strings.TrimPrefix(key, "env/"), "/last") // sanitize the strings
	_, err = etcds.kv.Put(ctx, fmt.Sprintf("env/%s/incidents/%s", key, date), string(encoded))
	cancel()
	return errors.Wrap(err, "failed to store key")
}

// All returns a map of all keys and values
func (etcds *EtcdStorage) AllCheckins() (map[string]time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	v, err := etcds.kv.Get(ctx, "env", clientv3.WithPrefix())
	cancel()
	if err != nil {
		return map[string]time.Time{}, errors.Wrap(err, "failed to retrieve all checkins")
	}
	all := map[string]time.Time{}
	for _, item := range v.Kvs {
		if strings.Contains(string(item.Key), "last") {
			var data time.Time
			if err := json.Unmarshal(item.Value, &data); err != nil {
				return map[string]time.Time{}, errors.Wrap(err, "failed to retrieve all checkins")
			}
			all[string(item.Key)] = data
		}
	}
	return all, nil
}

func (etcds *EtcdStorage) AllIncidents() (map[string][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	v, err := etcds.kv.Get(ctx, "env", clientv3.WithPrefix())
	cancel()
	if err != nil {
		return map[string][]string{}, errors.Wrap(err, "failed to retrieve all stored environments")
	}
	all := map[string][]string{}
	for _, item := range v.Kvs {
		if strings.Contains(string(item.Key), "incidents") {
			var data []string
			if err := json.Unmarshal(item.Value, &data); err != nil {
				return map[string][]string{}, errors.Wrap(err, "failed to retrieve all stored incidents")
			}
			all[fmt.Sprintf("%s", string(item.Key))] = data
		}
	}
	return all, nil
}

// Purge removes all keys with directory/env prefix
func (etcds *EtcdStorage) Purge(prefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := etcds.kv.Delete(ctx, prefix, clientv3.WithPrefix())
	cancel()
	return err
}
