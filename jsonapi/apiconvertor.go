package jsonapi

import (
	"reflect"
)

// ResourceConvertor interface should be implemented by all models
// that need explicit conversion to JSONAPI Resource
type ResourceConvertor interface {
	ToResource(includes *Includes) *Resource
}

// Apis is helper function to extract all ApiConvertors
func Apis(data interface{}) []ResourceConvertor {
	t := reflect.ValueOf(data)

	if t.Kind() == reflect.Ptr {
		elm := t.Elem()
		if elm.Kind() == reflect.Slice {
			result := make([]ResourceConvertor, 0, elm.Len())
			for i := 0; i < elm.Len(); i++ {
				value := reflect.New(elm.Index(i).Type())
				value.Elem().Set(elm.Index(i))
				if api, ok := value.Interface().(ResourceConvertor); ok {
					result = append(result, api)
				}
			}

			return result
		}
	}

	if api, ok := data.(ResourceConvertor); ok {
		return []ResourceConvertor{api}
	}

	return []ResourceConvertor{}
}
