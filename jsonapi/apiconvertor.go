package jsonapi

import (
	"reflect"
)

// ApiConvertor interface should be implemented by all models
// that need explicit conversion to JSONAPI Resource
type ApiConvertor interface {
	JsonApiResource(includes *Includes) *Resource
}

// Helper function to extract all ApiConvertors
func Apis(data interface{}) []ApiConvertor {
	t := reflect.ValueOf(data)

	if t.Kind() == reflect.Ptr {
		elm := t.Elem()
		if elm.Kind() == reflect.Slice {
			result := make([]ApiConvertor, 0, elm.Len())
			for i := 0; i < elm.Len(); i++ {
				value := reflect.New(elm.Index(i).Type())
				value.Elem().Set(elm.Index(i))
				if api, ok := value.Interface().(ApiConvertor); ok {
					result = append(result, api)
				}
			}

			return result
		}
	}

	if api, ok := data.(ApiConvertor); ok {
		return []ApiConvertor{api}
	} else {
		return []ApiConvertor{}
	}
}
