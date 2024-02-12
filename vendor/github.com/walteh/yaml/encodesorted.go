package yaml

import (
	"encoding/json"
	"reflect"
)

// this is a direct copy of the v2 encoder
// https://github.com/go-yaml/yaml/blob/7649d4548cb53a614db133b2a8ac1f31859dda8c/encode.go

var (
	mapItemType = reflect.TypeOf(MapSlice{})

	_ json.Marshaler   = MapItem{}
	_ json.Marshaler   = MapSlice{}
	_ json.Unmarshaler = &MapItem{}
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
	return json.Marshal([]MapItem(m))
}

func (m *MapSlice) UnmarshalJSON(data []byte) error {
	var slice []MapItem
	err := json.Unmarshal(data, &slice)
	if err == nil {
		*m = MapSlice(slice)
	}
	return err
}

func (m MapItem) MarshalJSON() ([]byte, error) {
	mapd := make(map[string]any)
	mapd[m.Key.(string)] = m.Value
	return json.Marshal(mapd)
}

// Encode encodes the map slice to a YAML byte slice
func (m *MapItem) UnmarshalJSON(data []byte) error {
	mapd := make(map[string]any)
	if err := json.Unmarshal(data, &mapd); err != nil {
		return err
	}

	for k, v := range mapd {
		m.Value = k
		m.Key = v
	}

	return nil
}
