package jsonapi

import (
	"fmt"
	"net/http"
)

// DocError creates JSONAPI DocItem document representing errors from error slice
func DocError(httpErrorCode int, errors ...error) *DocItem {
	errorlist := make([]Err, len(errors))

	for i := range errors {
		errorlist[i].Code = string(httpErrorCode)
		errorlist[i].Status = http.StatusText(httpErrorCode)
		errorlist[i].Detail = errors[i].Error()
	}

	return &DocItem{Data: nil, Errors: errorlist}
}

// NewResource creates JSONAPI Resource initialized with id and type fields
func NewResource(id, typeName string) *Resource {
	return &Resource{
		ID:            id,
		Type:          typeName,
		Attributes:    make(map[string]interface{}),
		Relationships: make(map[string]*Relationship),
		Links:         make(map[string]interface{}),
		Meta:          make(map[string]interface{}),
	}
}

// SetOneRelationship
func (r *Resource) SetOneRelationship(name string, model interface{}, includes *Includes) *Relationship {

	rel := &Relationship{}
	r.Relationships[name] = rel

	rel.Links.Related = fmt.Sprintf("%v", name)

	api, ok := model.(ResourceConvertor)
	if !ok {
		//println(name, ": not ResourceConvertor")
		return rel
	}

	resource := api.ToResource(includes)

	if resource.ID == "" {
		//println(name, ": no ID")
		return rel
	}

	rel.Links.Self = fmt.Sprintf("/%v/%v", name, r.ID)
	rel.Data.IsSingle = true
	rel.Data.ResourceIds = []ResourceIdentifier{
		{ID: resource.ID, Type: resource.Type},
	}

	if includes != nil {
		includes.Set(resource.ID, resource)
	}

	return rel
}

// SetManyRelationship
func (r *Resource) SetManyRelationship(name string, models []ResourceConvertor, includes *Includes) *Relationship {

	rel := &Relationship{}
	r.Relationships[name] = rel
	rel.Links.Related = fmt.Sprintf("%v", name)

	if len(models) == 0 {
		//println(name, ": no models")
		return rel
	}

	resType := name

	rel.Data.ResourceIds = make([]ResourceIdentifier, len(models))
	for i, item := range models {
		resource := item.ToResource(includes)
		rel.Data.ResourceIds[i] = ResourceIdentifier{
			ID:   resource.ID,
			Type: resource.Type,
		}
		if includes != nil {
			includes.Set(resource.ID, resource)
		}
		resType = resource.Type
	}
	rel.Links.Self = fmt.Sprintf("/%v", resType)
	rel.Data.IsSingle = false

	return rel
}
