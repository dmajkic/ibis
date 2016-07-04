package gorm

import (
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

func (g *noneDriver) ConnectDB(config map[string]string) (jsonapi.Database, error) {
	return &noneDriver{sync.RWMutex{}}, nil
}

func getSliceValue(model interface{}) reflect.Value {
	switch reflect.TypeOf(model).Kind() {
	case reflect.Slice, reflect.Array:
		return reflect.ValueOf(model)
	case reflect.Ptr:
		if reflect.TypeOf(reflect.ValueOf(model).Elem().Interface()).Kind() == reflect.Slice {
			return getSliceValue(reflect.ValueOf(model).Elem().Interface())
		}
	case reflect.Func:
		result := model.(func() interface{})()
		return getSliceValue(result)
	}

	model_type := reflect.TypeOf(model)
	models := reflect.MakeSlice(reflect.SliceOf(model_type), 1, 1)
	models.Index(0).Set(reflect.ValueOf(model))

	return models
}

// getItemId returns ID of the item by best guess
func getItemId(item reflect.Value) interface{} {

	// If item suports Resources - Use that
	if r, ok := item.Interface().(jsonapi.Resourcer); ok {
		return r.GetID()
	}

	// If item is not a struct, then whole object is used as ID
	if item.Kind() != reflect.Struct {
		return item.Interface()
	}

	// If struct item has field ID - use it
	if value := item.FieldByName("ID"); value.IsValid() {
		return value.Interface()
	}

	// If struct item has field Id - use it
	if value := item.FieldByName("Id"); value.IsValid() {
		return value.Interface()
	}

	// Return whole struct as its ID
	return item.Interface()
}

// findItem finds id in slice using reflection
func findItem(slice reflect.Value, id interface{}) int {

	switch slice.Kind() {
	case reflect.Struct:
		if getItemId(slice) == id {
			return 0
		} else {
			return -1
		}
	case reflect.Slice, reflect.Array:
		break
	default:
		return -1
	}

	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i)
		if getItemId(item) == id {
			return i
		}
	}

	return -1
}

func (g *noneDriver) FindAll(model interface{}, parent_id interface{}, query string) (*jsonapi.DocCollection, error) {
	g.Lock()
	defer g.Unlock()

	models := getSliceValue(model)

	collection := make([]*jsonapi.Resource, models.Len())
	includes := jsonapi.NewIncludes()

	for i, _ := range collection {
		collection[i] = g.ToResource(models.Index(i).Interface(), includes)
	}

	return &jsonapi.DocCollection{
		Data:     collection,
		Included: includes.ToArray(),
		JsonApi:  &jsonapi.JsonApiObject{Version: "1.0"},
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
		JsonApi:  &jsonapi.JsonApiObject{Version: "1.0"},
	}, nil
}

func (g *noneDriver) Delete(model interface{}, id interface{}) error {
	g.Lock()
	defer g.Unlock()

	models := getSliceValue(model)
	if idx := findItem(models, id); idx >= 0 {

	}

	return nil
}

func (g *noneDriver) Update(model interface{}, id interface{}, doc *jsonapi.DocItem) error {
	g.Lock()
	defer g.Unlock()

	models := getSliceValue(model)
	if idx := findItem(models, id); idx >= 0 {

	}

	return nil
}

func (g *noneDriver) Create(model interface{}, doc *jsonapi.DocItem) (*jsonapi.DocItem, error) {
	g.Lock()
	defer g.Unlock()

	// If there is Id, user cen return "204 Ok - No Content", or retreive record by id
	if doc.Data.Id != "" {
		doc.Data.Attributes["id"] = doc.Data.Id
		return nil, nil
	}

	models := getSliceValue(model)
	reflect.AppendSlice(models, reflect.ValueOf(model))

	return doc, nil
}

func (g *noneDriver) ToResource(value interface{}, includes *jsonapi.Includes) *jsonapi.Resource {

	// Use ApiConvertor interface if there is one
	if convertor, implements := value.(jsonapi.ApiConvertor); implements {
		return convertor.JsonApiResource(includes)
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
		Id:            id,
		Type:          typ.Type().Name(),
		Relationships: make(map[string]*jsonapi.Relationship),
		Attributes:    make(map[string]interface{}),
	}

	if typ.Kind() != reflect.Struct {
		resource.Attributes["value"] = value
		resource.Id = "value"
		return resource
	}

	kin := reflect.TypeOf(value)

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		t := kin.Field(i)

		if !t.Anonymous && (strings.ToUpper(t.Name) != "ID") {
			resource.Attributes[t.Name] = f.Interface()
		}
	}

	// Return value
	return resource
}
