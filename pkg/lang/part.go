package lang

import (
	"reflect"

	"github.com/walteh/yaml"
	"github.com/zclconf/go-cty/cty"
)

var (
	yamlcapsule = cty.Capsule("yaml", reflect.TypeOf(&yaml.MapSlice{}).Elem())
)

// type Finalizer struct {
// 	finalized bool
// }

// func (me *Finalizer) Finalized() bool {
// 	return me.finalized
// }

// type Part interface {
// 	Range() hcl.Range
// 	Finalized() bool
// }

// type PartList []Part

// type BlockLabelPart struct {
// 	Finalizer

// 	Name  string
// 	Range hcl.Range

// 	// optional
// 	Next *BlockLabelPart

// 	// optional
// 	Body *BlockBodyPart
// }

// type BlockBodyPart struct {
// 	Finalizer

// 	Range  hcl.Range
// 	Object ObjectPart
// 	HCL    *hcl.Block
// }

// type FilePart struct {
// 	Finalizer

// 	Range hcl.Range
// 	Items PartList
// 	HCL   *hcl.File
// }

// type AttributePart struct {
// 	Finalizer

// 	Range hcl.Range
// 	Name  string
// 	HCL   hcl.Attribute
// }

// type ObjectPart struct {
// 	Finalizer

// 	Range hcl.Range
// 	Items PartList
// }
