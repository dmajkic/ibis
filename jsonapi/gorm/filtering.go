package gorm

import (
	"github.com/jinzhu/gorm"
)

type Scoper interface {
	DefaultScope(parent_id interface{}) func(*gorm.DB) *gorm.DB
}

type Orderer interface {
	DefaultOrder() func(*gorm.DB) *gorm.DB
}

func DefaultScopes(model interface{}, parent_id interface{}) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0, 2)

	if scoper, ok := model.(Scoper); ok {
		scopes = append(scopes, scoper.DefaultScope(parent_id))
	}

	if orderer, ok := model.(Orderer); ok {
		scopes = append(scopes, orderer.DefaultOrder())
	}

	return scopes
}
