package yaml

import (
	"encoding/json"
	"reflect"
)

// this is a direct copy of the v2 encoder
// https://github.com/go-yaml/yaml/blob/7649d4548cb53a614db133b2a8ac1f31859dda8c/encode.go

var (
	mapItemType = reflect.TypeOf(MapSlice{})

	_ json.Marshaler   = MapSlice{}
	_ json.Unmarshaler = &MapSlice{}
)

type MapSlice []MapItem

type MapItem struct {
	Key   any
	Value any
}

func (e *encoder) itemsv(tag string, slice MapSlice) {
	e.mappingv(tag, func() {
		for _, item := range slice {
			e.marshal("", reflect.ValueOf(item.Key))
			e.marshal("", reflect.ValueOf(item.Value))
		}
	})
}

func (m MapSlice) MarshalJSON() ([]byte, error) {
	mapper, err := NewOrderedMapFromKVPairs(m)
	if err != nil {
		return nil, err
	}
	return json.Marshal(mapper)
}

func (m *MapSlice) UnmarshalJSON(data []byte) error {
	kvp := NewOrderedMap()
	err := kvp.UnmarshalJSON(data)
	if err == nil {
		*m = kvp.ToMapSlice()
	}
	return err
}
