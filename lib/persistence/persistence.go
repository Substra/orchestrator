// Package persistence holds everything related to data persistence
package persistence

// Database is the main interface to act on the persistence layer
// This covers all CRUD operations
type Database interface {
	DBWriter
	DBReader
}

// DBWriter handles persisting and updating data
type DBWriter interface {
	PutState(resource string, key string, data []byte) error
}

// DBReader handles data retrieval
type DBReader interface {
	GetState(resource string, key string) ([]byte, error)
	GetAll(resource string) ([][]byte, error)
}
