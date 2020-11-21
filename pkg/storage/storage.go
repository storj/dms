package storage

import "time"

// Data is the data object stored under the environment key
type Data struct {
	Last time.Time `json:"last"`
}

// StorageImpl is an interface for the storage implemntation
type StorageImpl interface {
	Store(string, Data) error
	All() (map[string]Data, error)
}
