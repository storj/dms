package storage

import "time"

// StorageImpl is an interface for the storage implemntation
type StorageImpl interface {
	StoreCheckin(string, time.Time) error
	StoreIncident(string, string, []time.Time) error
	AllCheckins() (map[string]time.Time, error)
	AllIncidents() (map[string][]string, error)
}
