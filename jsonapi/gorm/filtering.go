package gorm

import (
	"github.com/jinzhu/gorm"
)

// Scoper sets default scope on model
type Scoper interface {
	DefaultScope(parentID interface{}) func(*gorm.DB) *gorm.DB
}

// Orderer defauilt order for interface
type Orderer interface {
	DefaultOrder() func(*gorm.DB) *gorm.DB
}

// DefaultScopes adds default scopes to gorm query
func DefaultScopes(model interface{}, parentID interface{}) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0, 2)

	if scoper, ok := model.(Scoper); ok {
		scopes = append(scopes, scoper.DefaultScope(parentID))
	}

	if orderer, ok := model.(Orderer); ok {
		scopes = append(scopes, orderer.DefaultOrder())
	}

	return scopes
}
