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
	jsonapi.RegisterDriver("gorm", &gormDriver{sync.RWMutex{}, nil})
}

func errConv(err error) error {
	switch err {
	case gorm.ErrRecordNotFound:
		return jsonapi.ErrNotFound
	default:
		return err
	}
}

func (g *gormDriver) ConnectDB(config map[string]string) error {
	db, err := gorm.Open(config["adapter"], config["dbUrl"])
	if err != nil {
		return err
	}

	db.LogMode(true)
	g.Orm = db
	return nil
}

func (g *gormDriver) FindAll(model interface{}, parentID interface{}, query string) (*jsonapi.DocCollection, error) {
	g.Lock()
	defer g.Unlock()

	modelType := reflect.TypeOf(model)
	modelsType := reflect.MakeSlice(reflect.SliceOf(modelType), 0, 0).Type()
	models := reflect.New(modelsType)

	// page[number] and page[size]

	// TODO: Query parameters - "400 Bad Requset" if not possible
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

	scopes := DefaultScopes(model, parentID)

	if err := g.Orm.Scopes(scopes...).Find(models.Interface()).Error; err != nil {
		return nil, err
	}

	collection := make([]*jsonapi.Resource, models.Elem().Len())
	includes := jsonapi.NewIncludes()

	for i := range collection {
		collection[i] = g.ToResource(models.Elem().Index(i).Interface(), includes)
	}

	return &jsonapi.DocCollection{
		Data:     collection,
		Included: includes.ToArray(),
		JSONApi:  &jsonapi.VersionMeta{Version: "1.0"},
	}, nil
}

func (g *gormDriver) FindRecord(model, id interface{}, query string) (*jsonapi.DocItem, error) {
	g.Lock()
	defer g.Unlock()

	modelType := reflect.TypeOf(model)
	modelCopy := reflect.New(modelType).Interface()

	if err := g.Orm.Find(modelCopy, "id=?", id).Error; err != nil {
		return nil, errConv(err)
	}

	includes := jsonapi.NewIncludes()
	item := g.ToResource(modelCopy, includes)

	return &jsonapi.DocItem{
		Data:     item,
		Included: includes.ToArray(),
		JSONApi:  &jsonapi.VersionMeta{Version: "1.0"},
	}, nil
}

func (g *gormDriver) Delete(model interface{}, id interface{}) error {
	g.Lock()
	defer g.Unlock()

	modelType := reflect.TypeOf(model)
	modelCopy := reflect.New(modelType).Interface()
	err := g.Orm.Delete(modelCopy, "id=?", id).Error

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
	if doc.Data.ID != "" {
		doc.Data.Attributes["id"] = doc.Data.ID
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
	if convertor, implements := value.(jsonapi.ResourceConvertor); implements {
		return convertor.ToResource(includes)
	}

	// Else use orm reflection support
	scope := g.Orm.NewScope(value)

	id := ""
	if v, ok := value.(jsonapi.Resourcer); ok && !scope.PrimaryKeyZero() {
		id = v.GetID()
	}

	resource := &jsonapi.Resource{
		ID:            id,
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
