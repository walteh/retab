package yamlwrite

import (
	"context"
	"io"
	"reflect"
	"unsafe"

	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
	"gopkg.in/yaml.v3"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewYamlFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.yaml", "*.yml"}
}

type FlowInterface struct {
	Sup map[string]interface{} `yaml:"sup,flow"`
}

func (me *Formatter) Format(_ context.Context, cfg configuration.Provider, read io.Reader) (io.Reader, error) {

	dec := yaml.NewDecoder(read)

	v := interface{}(nil)

	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	read, write := io.Pipe()
	// defer write.Close()

	// wrap the writer so that everything that is written with spaces is replaced with tabs
	// tw := new(tabwriter.Writer)
	// write2 := tw.Init(write, 4, 0, 1, ' ', 0)

	enc := yaml.NewEncoder(write)
	enc.SetIndent(4)
	rs := reflect.ValueOf(enc).Elem()
	rf := rs.FieldByName("encoder").Elem().FieldByName("flow")
	// rf can't be read or set.
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()

	// // rf is now a reflect.Value representing the flow field of the encoder.
	// // We can now read and set it.
	rf.Set(reflect.ValueOf(true))

	// rf = rs.FieldByName("encoder").Elem().FieldByName("emitter")
	// // rf can't be read or set.
	// rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()

	// rf = rf.FieldByName("indent")

	// rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()

	// // rf is now a reflect.Value representing the flow field of the encoder.
	// // We can now read and set it.
	// rf.Set(reflect.ValueOf(1))

	// pp.Println(v)

	// start the formatter in a goroutine

	go func() {
		// defer tw.Flush()
		defer enc.Close()

		enc.SetIndent(4)

		if err := enc.Encode(&v); err != nil {
			err := write.CloseWithError(err)
			if err != nil {
				panic(err)
			}
			return
		}
		if err := write.Close(); err != nil {
			panic(err)
		}
	}()

	return read, nil
}
