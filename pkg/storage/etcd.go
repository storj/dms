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
func (etcds *EtcdStorage) Store(key string, value Data) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	encoded, _ := json.Marshal(value)
	_, err := etcds.kv.Put(ctx, fmt.Sprintf("env=%s", key), string(encoded))
	cancel()
	return errors.Wrap(err, "failed to store key")
}

// All returns a map of all keys and values
func (etcds *EtcdStorage) All() (map[string]Data, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	v, err := etcds.kv.Get(ctx, "env", clientv3.WithPrefix())
	cancel()
	if err != nil {
		return map[string]Data{}, errors.Wrap(err, "failed to retrieve all stored values")
	}

	var valsInvalidFormat []string
	all := map[string]Data{}
	for _, item := range v.Kvs {
		var data Data
		if err := json.Unmarshal(item.Value, &data); err != nil {
			valsInvalidFormat = append(valsInvalidFormat, string(item.Key))
			continue
		}
		all[string(item.Key)] = data
	}

	if len(valsInvalidFormat) > 0 {
		return all, errors.Wrap(err,
			fmt.Sprintf("failed unmarshaling the stored values with key: %s", strings.Join(valsInvalidFormat, ", ")),
		)
	}
	return all, nil
}

// Purge removes all keys with "env" prefix
func (etcds *EtcdStorage) Purge() error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := etcds.kv.Delete(ctx, "env", clientv3.WithPrefix())
	cancel()
	return err
}
