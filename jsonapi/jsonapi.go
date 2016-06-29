// JSONAPI v1.0 implementatiopn from http://jsonapi.org/format/1.0
package jsonapi

import (
	"encoding/json"
)

//  JsonApiObject optional version info
type JsonApiObject struct {
	Version string                 `json:"version,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// DocItem represents JSONAPI document where
// Data represents single resource
type DocItem struct {
	Data     *Resource              `json:"data,omitempty"`
	Errors   []Err                  `json:"errors,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
	JsonApi  *JsonApiObject         `json:"jsonapi,omitempty"`
	Links    *Links                 `json:"links,omitempty"`
	Included []*Resource            `json:"included,omitempty"`
}

// DocItem represents JSONAPI document where
// Data represents collection of resources
type DocCollection struct {
	Data     []*Resource            `json:"data"`
	Errors   []Err                  `json:"errors,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
	JsonApi  *JsonApiObject         `json:"jsonapi,omitempty"`
	Links    *Links                 `json:"links,omitempty"`
	Included []*Resource            `json:"included,omitempty"`
}

// Reousrce JSONAPI represents object by type and id
type ResourceIdentifier struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

// Reousrce JSONAPI representation of single item
type Resource struct {
	Id            string                   `json:"id,omitempty"`
	Type          string                   `json:"type"`
	Attributes    map[string]interface{}   `json:"attributes,omitempty"`
	Relationships map[string]*Relationship `json:"relationships,omitempty"`
	Links         map[string]interface{}   `json:"links,omitempty"`
	Meta          map[string]interface{}   `json:"meta,omitempty"`
}

// Links represents JSONAPI links related to document
// or resource
type Links struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
	First   string `json:"first,omitempty"`
	Last    string `json:"last,omitempty"`
	Prev    string `json:"prev,omitempty"`
	Next    string `json:"next,omitempty"`
}

// Relationship represents resource ToOne and ToMany
// relationships.
type Relationship struct {
	Links Links                  `json:"links,omitempty"`
	Data  RelationshipData       `json:"data,omitempty"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
}

type RelationshipData struct {
	ResourceIds []ResourceIdentifier
	IsSingle    bool
}

func (r *Resource) String() string {
	data, _ := json.Marshal(r)
	return string(data)
}

func (d *DocItem) String() string {
	data, _ := json.Marshal(d)
	return string(data)
}

func (d *DocCollection) String() string {
	data, _ := json.Marshal(d)
	return string(data)
}

// Relationship.Data can be ResourceIdentifier or []ResourceIdentifier
func (r *RelationshipData) UnmarshalJSON(b []byte) (err error) {

	single := ResourceIdentifier{}
	if err = json.Unmarshal(b, &single); err == nil {
		r.IsSingle = true
		r.ResourceIds = []ResourceIdentifier{single}

		return nil
	}

	if err != nil {
		return err
	}

	array := make([]ResourceIdentifier, 0)
	if err = json.Unmarshal(b, &array); err == nil {
		r.IsSingle = false
		r.ResourceIds = array
		return nil
	}

	return err
}

func (r *RelationshipData) MarshalJSON() ([]byte, error) {

	if r.IsSingle {
		if len(r.ResourceIds) > 0 {
			return json.Marshal(r.ResourceIds[0])
		} else {
			return json.Marshal(nil)
		}
	}

	return json.Marshal(r.ResourceIds)
}

// Err is JSONAPI Error resource
type Err struct {
	Id    string `json:"id,omitempty"`
	Links struct {
		About string `json:"about,omitempty"`
	} `json:"links,omitempty"`
	Status string `json:"status,omitempty"`
	Code   string `json:"code,omitempty"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
	Source struct {
		Pointer   string `json:"pointer,omitempty"`
		Parameter string `json:"parameter,omitempty"`
	} `json:"source,omitempty"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}
