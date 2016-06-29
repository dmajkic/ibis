package jsonapi

import (
	"errors"
	"fmt"
)

// DB Interface that must be implemente by driver
type Database interface {
	ConnectDB(config map[string]string) (Database, error)
	FindAll(model, parent_id interface{}, query string) (*DocCollection, error)
	FindRecord(model, id interface{}, query string) (*DocItem, error)
	Delete(model, id interface{}) error
	Update(model, id interface{}, doc *DocItem) error
	Create(model interface{}, doc *DocItem) (*DocItem, error)
	ToResource(value interface{}, includes *Includes) *Resource
}

var (
	// ErrNotFound: record not found error, driver should return this for jsonapi specifed return code
	ErrNotFound = errors.New("Record not found")
)

// MetaFiller interface is implemented by models that
// need database access to supply additional data to resource
type MetaFiller interface {
	AddMeta(Database, *DocItem)
}

// All models should be able to return ID as string
type Resourcer interface {
	GetID() string
}

var dbDriverMap = map[string]Database{}

// RegisterDriver should be used from driver implementation
func RegisterDriver(name string, database Database) {
	dbDriverMap[name] = database
}

// Database access constructor
func NewDatabase(driver string) (Database, error) {
	if db, ok := dbDriverMap[driver]; ok {
		return db, nil
	}

	return nil, fmt.Errorf("Unknown database driver '%v'", driver)
}
