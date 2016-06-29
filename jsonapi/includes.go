package jsonapi

// Includes is implementation of JSONAPI resource array
// included in JSONAPI document
type Includes struct {
	m map[string]*Resource
}

func NewIncludes() *Includes {
	return &Includes{make(map[string]*Resource)}
}

func (includes *Includes) Set(key string, value *Resource) {
	includes.m[key] = value
}

func (includes *Includes) Get(key string) *Resource {
	return includes.m[key]
}

func (includes *Includes) ToArray() []*Resource {
	result := make([]*Resource, len(includes.m))

	i := 0
	for _, v := range includes.m {
		result[i] = v
		i++
	}
	return result
}
