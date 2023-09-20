package ctydebug

import (
	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
)

// CmpOptions is a set of options for package github.com/google/go-cmp/cmp
// that help it to work well with cty.Type and cty.Value when they appear as
// part of a pair of data structures being compared.
//
// Firstly, it converts collection and structural types into Go named
// types for either map[string]cty.Value or []cty.Value, so that type
// checking will still find these types to be distinct but cmp is able to
// understand how to recursively check inside them.
//
// Secondly, it knows how to compare leaf cty.Type and cty.Value values using
// their built-in definitions of equality.
var CmpOptions cmp.Option

var transformValueOp = cmp.Transformer("ctydebug.TransformValueForCmp", transformValueForCmp)
var transformTypeOp = cmp.Transformer("ctydebug.TransformTypeForCmp", transformTypeForCmp)

func init() {
	CmpOptions = cmp.Options{
		cmp.FilterValues(
			valuesCanCompareDeep,
			transformValueOp,
		),
		cmp.FilterValues(func(a, b cty.Value) bool {
			return !valuesCanCompareDeep(a, b)
		}, cmp.Comparer(cty.Value.RawEquals)),
		cmp.FilterValues(
			typesCanCompareDeep,
			transformTypeOp,
		),
		cmp.FilterValues(func(a, b cty.Type) bool {
			return !typesCanCompareDeep(a, b)
		}, cmp.Comparer(cty.Type.Equals)),
	}
}

func valuesCanCompareDeep(a, b cty.Value) bool {
	if a == cty.NilVal || b == cty.NilVal {
		return false
	}
	aTy := a.Type()
	bTy := b.Type()

	return (aTy.IsCollectionType() || aTy.IsTupleType() || aTy.IsObjectType()) &&
		(bTy.IsCollectionType() || bTy.IsTupleType() || bTy.IsObjectType())
}

func typesCanCompareDeep(a, b cty.Type) bool {
	if a == cty.NilType || b == cty.NilType {
		return false
	}

	return (a.IsCollectionType() || a.IsTupleType() || a.IsObjectType()) &&
		(b.IsCollectionType() || b.IsTupleType() || b.IsObjectType())
}

func transformValueForCmp(v cty.Value) interface{} {
	if v == cty.NilVal {
		return v
	}
	ty := v.Type()
	switch {

	case v.IsNull() || !v.IsKnown():
		return v

	case ty.IsObjectType():
		return ctyObjectVal(v.AsValueMap())

	case ty.IsMapType():
		return ctyMapVal(v.AsValueMap())

	case ty.IsTupleType():
		return ctyTupleVal(v.AsValueSlice())

	case ty.IsListType():
		return ctyListVal(v.AsValueSlice())

	case ty.IsSetType():
		return ctySetVal(v.AsValueSlice())

	default:
		return v
	}
}

type ctyTupleVal []cty.Value

func (w ctyTupleVal) ctyValue() cty.Value { return cty.TupleVal(w) }

type ctyListVal []cty.Value

func (w ctyListVal) ctyValue() cty.Value {
	if len(w) == 0 {
		return cty.ListValEmpty(cty.DynamicPseudoType) // lossy
	}
	return cty.ListVal(w)
}

type ctySetVal []cty.Value

func (w ctySetVal) ctyValue() cty.Value {
	if len(w) == 0 {
		return cty.SetValEmpty(cty.DynamicPseudoType) // lossy
	}
	return cty.SetVal(w)
}

type ctyObjectVal map[string]cty.Value

func (w ctyObjectVal) ctyValue() cty.Value { return cty.ObjectVal(w) }

type ctyMapVal map[string]cty.Value

func (w ctyMapVal) ctyValue() cty.Value {
	if len(w) == 0 {
		return cty.MapValEmpty(cty.DynamicPseudoType) // lossy
	}
	return cty.MapVal(w)
}

// transformTypeForCmp is a function suitable for use with cmp.Transformer
// on package github.com/google/go-cmp/cmp that turns cty collection and
// structural types into Go maps and slices so that cmp can understand
//.how to recursively compare them.
func transformTypeForCmp(ty cty.Type) interface{} {
	if ty == cty.NilType {
		return ty
	}

	switch {

	case ty.IsObjectType():
		return ctyObjectType(ty.AttributeTypes())

	case ty.IsMapType():
		return ctyMapType{ty.ElementType()}

	case ty.IsTupleType():
		return ctyTupleType(ty.TupleElementTypes())

	case ty.IsListType():
		return ctyListType{ty.ElementType()}

	case ty.IsSetType():
		return ctySetType{ty.ElementType()}

	default:
		return ty
	}
}

type ctyObjectType map[string]cty.Type

func (w ctyObjectType) ctyType() cty.Type { return cty.Object(w) }

type ctyTupleType []cty.Type

func (w ctyTupleType) ctyType() cty.Type { return cty.Tuple(w) }

type ctyListType [1]cty.Type

func (w ctyListType) ctyType() cty.Type { return cty.List(w[0]) }

type ctyMapType [1]cty.Type

func (w ctyMapType) ctyType() cty.Type { return cty.Map(w[0]) }

type ctySetType [1]cty.Type

func (w ctySetType) ctyType() cty.Type { return cty.Set(w[0]) }
