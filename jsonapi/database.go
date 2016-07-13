package jsonapi

import (
	"errors"
	"fmt"
	"unicode"
)

// Database is ibis DB interface to model storage. It is implemented by drivers
type Database interface {
	ConnectDB(config map[string]string) error
	FindAll(model, parentID interface{}, query string) (*DocCollection, error)
	FindRecord(model, id interface{}, query string) (*DocItem, error)
	Delete(model, id interface{}) error
	Update(model, id interface{}, doc *DocItem) error
	Create(model interface{}, doc *DocItem) (*DocItem, error)
	ToResource(value interface{}, includes *Includes) *Resource
}

var (
	// ErrNotFound Record Not Found error, driver should return this for jsonapi specifed return code
	ErrNotFound = errors.New("Record not found")
)

// MetaFiller interface is implemented by models that
// need database access to supply additional data to resource
type MetaFiller interface {
	AddMeta(Database, *DocItem)
}

// Resourcer interface should be suported by models to return ID as string
type Resourcer interface {
	GetID() string
}

// Helper to convert Go public names camelCase
func LowerInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

var dbDriverMap = map[string]Database{}

// RegisterDriver should be used from driver implementation
func RegisterDriver(name string, database Database) {
	dbDriverMap[name] = database
}

// NewDatabase creates new DB driver object
func NewDatabase(driver string) (Database, error) {
	if db, ok := dbDriverMap[driver]; ok {
		return db, nil
	}

	return nil, fmt.Errorf("Unknown database driver '%v'", driver)
}
