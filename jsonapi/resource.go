package jsonapi

import (
	"fmt"
	"net/http"
)

func DocError(httpErrorCode int, errors ...error) *DocItem {
	errorlist := make([]Err, len(errors))

	for i := range errors {
		errorlist[i].Code = string(httpErrorCode)
		errorlist[i].Status = http.StatusText(httpErrorCode)
		errorlist[i].Detail = errors[i].Error()
	}

	return &DocItem{Data: nil, Errors: errorlist}
}

func NewResource(id, typeName string) *Resource {
	return &Resource{
		Id:            id,
		Type:          typeName,
		Attributes:    make(map[string]interface{}),
		Relationships: make(map[string]*Relationship),
		Links:         make(map[string]interface{}),
		Meta:          make(map[string]interface{}),
	}
}

func (r *Resource) SetOneRelationship(name string, model interface{}, includes *Includes) *Relationship {

	rel := &Relationship{}
	r.Relationships[name] = rel

	rel.Links.Related = fmt.Sprintf("%v", name)

	api, ok := model.(ApiConvertor)
	if !ok {
		//println(name, ": not ApiConvertor")
		return rel
	}

	resource := api.JsonApiResource(includes)

	if resource.Id == "" {
		//println(name, ": no ID")
		return rel
	}

	rel.Links.Self = fmt.Sprintf("/%v/%v", name, r.Id)
	rel.Data.IsSingle = true
	rel.Data.ResourceIds = []ResourceIdentifier{
		{Id: resource.Id, Type: resource.Type},
	}

	if includes != nil {
		includes.Set(resource.Id, resource)
	}

	return rel
}

func (r *Resource) SetManyRelationship(name string, models []ApiConvertor, includes *Includes) *Relationship {

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
		if api, ok := item.(ApiConvertor); ok {
			resource := api.JsonApiResource(includes)
			rel.Data.ResourceIds[i] = ResourceIdentifier{
				Id:   resource.Id,
				Type: resource.Type,
			}
			if includes != nil {
				includes.Set(resource.Id, resource)
			}
			resType = resource.Type
		}
	}
	rel.Links.Self = fmt.Sprintf("/%v", resType)
	rel.Data.IsSingle = false

	return rel
}
