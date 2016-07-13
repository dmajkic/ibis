package none

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/dmajkic/ibis/jsonapi"
)

type noneDriver struct {
	sync.RWMutex
}

func init() {
	jsonapi.RegisterDriver("none", &noneDriver{sync.RWMutex{}})
}

func (g *noneDriver) ConnectDB(config map[string]string) error {
	return nil
}

func getSliceValue(model interface{}) []interface{} {

	switch v := model.(type) {
	case []interface{}:
		return v
	case *[]interface{}:
		return *v
	case func() interface{}:
		return getSliceValue(v())
	default:
		return []interface{}{model}
	}
}

// getItemId returns ID of the item by best guess
func getItemID(item interface{}) interface{} {

	// If item suports Resources - Use that
	if r, ok := item.(jsonapi.Resourcer); ok {
		return r.GetID()
	}

	v := reflect.ValueOf(item)

	// If item is not a struct, then whole object is used as ID
	if v.Kind() != reflect.Struct {
		return item
	}

	// If struct item has field ID - use it
	if value := v.FieldByName("ID"); v.IsValid() {
		return value.Interface()
	}

	// If struct item has field Id - use it
	if value := v.FieldByName("Id"); v.IsValid() {
		return value.Interface()
	}

	// Return whole struct as its ID
	println("No ID:", item)
	return item
}

// findItem finds id in slice using reflection
func findItem(slice []interface{}, id interface{}) int {

	for i, item := range slice {
		if getItemID(item) == id {
			return i
		}
	}

	return -1
}

func (g *noneDriver) FindAll(model interface{}, parentID interface{}, query string) (*jsonapi.DocCollection, error) {
	g.Lock()
	defer g.Unlock()

	models := getSliceValue(model)

	collection := make([]*jsonapi.Resource, len(models))
	includes := jsonapi.NewIncludes()

	for i := range collection {
		collection[i] = g.ToResource(models[i], includes)
	}

	return &jsonapi.DocCollection{
		Data:     collection,
		Included: includes.ToArray(),
		JSONApi:  &jsonapi.VersionMeta{Version: "1.0"},
	}, nil
}

func (g *noneDriver) FindRecord(model, id interface{}, query string) (*jsonapi.DocItem, error) {
	g.Lock()
	defer g.Unlock()

	includes := jsonapi.NewIncludes()
	item := g.ToResource(model, includes)

	return &jsonapi.DocItem{
		Data:     item,
		Included: includes.ToArray(),
		JSONApi:  &jsonapi.VersionMeta{Version: "1.0"},
	}, nil
}

func (g *noneDriver) Delete(model interface{}, id interface{}) error {
	g.Lock()
	defer g.Unlock()

	models := getSliceValue(model)
	if idx := findItem(models, id); idx >= 0 {
		copy(models[idx:], models[idx+1:])
		models[len(models)-1] = nil // or the zero value of T
		models = models[:len(models)-1]
	}

	return nil
}

func (g *noneDriver) Update(model interface{}, id interface{}, doc *jsonapi.DocItem) error {
	g.Lock()
	defer g.Unlock()

	models := getSliceValue(model)
	if idx := findItem(models, id); idx >= 0 {

	}

	return jsonapi.ErrNotFound
}

func (g *noneDriver) Create(model interface{}, doc *jsonapi.DocItem) (*jsonapi.DocItem, error) {
	g.Lock()
	defer g.Unlock()

	// If there is Id, user cen return "204 Ok - No Content", or retreive record by id
	if doc.Data.ID != "" {
		doc.Data.Attributes["id"] = doc.Data.ID
		return nil, nil
	}

	//append(getSliceValue(model), model)

	return doc, nil
}

func (g *noneDriver) ToResource(value interface{}, includes *jsonapi.Includes) *jsonapi.Resource {

	// Use ApiConvertor interface if there is one
	if convertor, implements := value.(jsonapi.ResourceConvertor); implements {
		return convertor.ToResource(includes)
	}

	id := ""
	if v, ok := value.(jsonapi.Resourcer); ok {
		id = v.GetID()
	}

	typ := reflect.ValueOf(value)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	resource := &jsonapi.Resource{
		ID:            id,
		Type:          typ.Type().Name(),
		Relationships: make(map[string]*jsonapi.Relationship),
		Attributes:    make(map[string]interface{}),
	}

	if typ.Kind() != reflect.Struct {
		resource.Attributes["value"] = value
		resource.ID = fmt.Sprintf("%v", value)
		return resource
	}

	kin := typ.Type()

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		t := kin.Field(i)

		if !t.Anonymous && (strings.ToUpper(t.Name) != "ID") {
			resource.Attributes[jsonapi.LowerInitial(t.Name)] = f.Interface()
		} else if resource.ID == "" {
			resource.ID = fmt.Sprintf("%v", f.Interface())
		}
	}

	// Return value
	return resource
}
