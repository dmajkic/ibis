// Package jsonapi implements JSONAPI v1.0 from http://jsonapi.org/format/1.0
package jsonapi

import (
	"encoding/json"
)

// VersionMeta represents JSONApiObject optional version info
type VersionMeta struct {
	Version string                 `json:"version,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// DocItem represents JSONAPI document where
// Data represents single resource
type DocItem struct {
	Data     *Resource              `json:"data,omitempty"`
	Errors   []Err                  `json:"errors,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
	JSONApi  *VersionMeta           `json:"jsonapi,omitempty"`
	Links    *Links                 `json:"links,omitempty"`
	Included []*Resource            `json:"included,omitempty"`
}

// DocCollection represents JSONAPI document where
// Data represents collection of resources
type DocCollection struct {
	Data     []*Resource            `json:"data"`
	Errors   []Err                  `json:"errors,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
	JSONApi  *VersionMeta           `json:"jsonapi,omitempty"`
	Links    *Links                 `json:"links,omitempty"`
	Included []*Resource            `json:"included,omitempty"`
}

// ResourceIdentifier represents object by type and id
type ResourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Resource JSONAPI representation of single item
type Resource struct {
	ID            string                   `json:"id,omitempty"`
	Type          string                   `json:"type"`
	Attributes    map[string]interface{}   `json:"attributes,omitempty"`
	Relationships map[string]*Relationship `json:"relationships,omitempty"`
	Links         map[string]interface{}   `json:"links,omitempty"`
	Meta          map[string]interface{}   `json:"meta,omitempty"`
}

// Links represent JSONAPI links related to document or resource
type Links struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
	First   string `json:"first,omitempty"`
	Last    string `json:"last,omitempty"`
	Prev    string `json:"prev,omitempty"`
	Next    string `json:"next,omitempty"`
}

// Relationship represents resource ToOne and ToMany relationships.
type Relationship struct {
	Links Links                  `json:"links,omitempty"`
	Data  RelationshipData       `json:"data,omitempty"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
}

// RelationshipData represents relationships information
type RelationshipData struct {
	ResourceIds []ResourceIdentifier
	IsSingle    bool
}

// String representation for Resource as JSON
func (r *Resource) String() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// String representation for DocItem as JSON
func (d *DocItem) String() string {
	data, _ := json.Marshal(d)
	return string(data)
}

// String representation for DocCollection as JSON
func (d *DocCollection) String() string {
	data, _ := json.Marshal(d)
	return string(data)
}

// UnmarshalJSON handles that Relationship.Data can be ResourceIdentifier or []ResourceIdentifier
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

	var array []ResourceIdentifier
	if err = json.Unmarshal(b, &array); err == nil {
		r.IsSingle = false
		r.ResourceIds = array
		return nil
	}

	return err
}

// MarshalJSON marshals JSONAPI relationship data to json
func (r *RelationshipData) MarshalJSON() ([]byte, error) {

	if r.IsSingle {
		if len(r.ResourceIds) > 0 {
			return json.Marshal(r.ResourceIds[0])
		}
		return json.Marshal(nil)
	}

	return json.Marshal(r.ResourceIds)
}

// Err is representation of JSONAPI Error resource
type Err struct {
	ID    string `json:"id,omitempty"`
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
