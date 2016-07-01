package gorm

import (
	"reflect"
	"sync"

	"github.com/dmajkic/ibis/jsonapi"

	"github.com/jinzhu/gorm"
	"github.com/nu7hatch/gouuid"
)

type gormDriver struct {
	sync.RWMutex
	Orm *gorm.DB
}

func init() {
	jsonapi.RegisterDriver("gorm", &gormDriver{})
}

func errConv(err error) error {
	switch err {
	case gorm.ErrRecordNotFound:
		return jsonapi.ErrNotFound
	default:
		return err
	}
}

func (g *gormDriver) ConnectDB(config map[string]string) (jsonapi.Database, error) {
	db, err := gorm.Open(config["adapter"], config["dbUrl"])
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	return &gormDriver{sync.RWMutex{}, db}, nil
}

func (g *gormDriver) GetMany(model, value interface{}, foreign_keys ...string) error {
	g.Lock()
	defer g.Unlock()
	return g.Orm.Model(model).Related(value, foreign_keys...).Error
}

func (g *gormDriver) GetOne(model, value interface{}, foreign_keys ...string) error {
	g.Lock()
	defer g.Unlock()
	return g.Orm.Model(model).Related(value, foreign_keys...).Error
}

func (g *gormDriver) FindAll(model interface{}, parent_id interface{}, query string) (*jsonapi.DocCollection, error) {
	g.Lock()
	defer g.Unlock()

	model_type := reflect.TypeOf(model)
	models_type := reflect.MakeSlice(reflect.SliceOf(model_type), 0, 0).Type()
	models := reflect.New(models_type)

	// page[number] and page[size]

	// TODO: Query parameters - "400 Bad Requset" ako nije moguce
	//var q url.Values
	//if q, err := url.ParseQuery(query); err != nil {
	//	return nil, err
	//}

	//include := q.Get("include")    // include=author,comments.author
	//sort := q.Get("sort")          // sort=-age,name
	//filter := q.Get("filter")      // filter=string
	//parent_id = q.Get("parent_id") // parent_id=1234567890
	//fileds := q.Get("fields[" + gorm.ToDBName() + "]") // fields[articles]=title,body
	//if fields, ok := q["fields"]; ok {}
	//if page, ok := q["page"]; ok {}

	scopes := DefaultScopes(model, parent_id)

	if err := g.Orm.Scopes(scopes...).Find(models.Interface()).Error; err != nil {
		return nil, err
	}

	collection := make([]*jsonapi.Resource, models.Elem().Len())
	includes := jsonapi.NewIncludes()

	for i, _ := range collection {
		collection[i] = g.ToResource(models.Elem().Index(i).Interface(), includes)
	}

	return &jsonapi.DocCollection{
		Data:     collection,
		Included: includes.ToArray(),
		JsonApi:  &jsonapi.JsonApiObject{Version: "1.0"},
	}, nil
}

func (g *gormDriver) FindRecord(model, id interface{}, query string) (*jsonapi.DocItem, error) {
	g.Lock()
	defer g.Unlock()

	model_type := reflect.TypeOf(model)
	model_copy := reflect.New(model_type).Interface()

	if err := g.Orm.Find(model_copy, "id=?", id).Error; err != nil {
		return nil, errConv(err)
	}

	includes := jsonapi.NewIncludes()
	item := g.ToResource(model_copy, includes)

	return &jsonapi.DocItem{
		Data:     item,
		Included: includes.ToArray(),
		JsonApi:  &jsonapi.JsonApiObject{Version: "1.0"},
	}, nil
}

func (g *gormDriver) Delete(model interface{}, id interface{}) error {
	g.Lock()
	defer g.Unlock()

	model_type := reflect.TypeOf(model)
	model_copy := reflect.New(model_type).Interface()
	err := g.Orm.Delete(model_copy, "id=?", id).Error

	return errConv(err)
}

func (g *gormDriver) Update(model interface{}, id interface{}, doc *jsonapi.DocItem) error {
	g.Lock()
	defer g.Unlock()

	err := g.Orm.First(&model, id).Updates(doc.Data.Attributes).Error

	return errConv(err)
}

func (g *gormDriver) Create(model interface{}, doc *jsonapi.DocItem) (*jsonapi.DocItem, error) {
	g.Lock()
	defer g.Unlock()

	// If there is Id, user cen return "204 Ok - No Content", or retreive record by id
	if doc.Data.Id != "" {
		doc.Data.Attributes["id"] = doc.Data.Id
		return nil, g.Orm.Model(model).Create(doc.Data.Attributes).Error
	}

	// We create Id, and retreive new record
	id, _ := uuid.NewV4()
	doc.Data.Attributes["id"] = id

	if err := g.Orm.Model(model).Create(doc.Data.Attributes).Error; err != nil {
		return nil, err
	}

	return g.FindRecord(model, id, "")
}

func (g *gormDriver) ToResource(value interface{}, includes *jsonapi.Includes) *jsonapi.Resource {

	// Use ApiConvertor interface if there is one
	if convertor, implements := value.(jsonapi.ApiConvertor); implements {
		return convertor.JsonApiResource(includes)
	}

	// Else use orm reflection support
	scope := g.Orm.NewScope(value)

	id := ""
	if v, ok := value.(jsonapi.Resourcer); ok && !scope.PrimaryKeyZero() {
		id = v.GetID()
	}

	resource := &jsonapi.Resource{
		Id:            id,
		Type:          scope.TableName(),
		Relationships: make(map[string]*jsonapi.Relationship),
		Attributes:    make(map[string]interface{}),
	}

	// Skip 'id' and '*_id' fields; add relationships
	for _, v := range scope.Fields() {

		if v.IsNormal && !v.IsPrimaryKey && !v.IsForeignKey && !v.IsIgnored {

			resource.Attributes[v.DBName] = v.Field.Interface()

		} else if !v.IsNormal && (v.Relationship != nil) {

			if (v.Relationship.Kind == "belongs_to") || (v.Relationship.Kind == "has_one") {
				value := reflect.New(v.Field.Type())
				value.Elem().Set(v.Field)
				resource.SetOneRelationship(v.DBName, value.Interface(), includes)
			} else if (v.Relationship.Kind == "has_many") || (v.Relationship.Kind == "many2many") {
				value := reflect.New(v.Field.Type())
				value.Elem().Set(v.Field)
				resource.SetManyRelationship(v.DBName, jsonapi.Apis(value.Interface()), includes)
			}
		}
	}

	// Return value
	return resource
}
