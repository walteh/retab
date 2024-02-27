package lang

import (
	"errors"
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/rs/xid"
)

type mergeMarker struct {
	ref    cty.ValueMarks
	common string
}

// CombinedAdaptedMergeConcatFunc is a dynamic function that can merge maps or arrays
// it looks at the first argument to determine if it should merge maps or arrays
// and then delegates to the appropriate function.
// The AdaptedMergeFuncSpec and AdapedConcatFuncSpec are copies of the MergeFunc and ConcatFunc specs from cty stdlib,
// but with the ability to merge marked parents into their children.
var CombinedMergeConcatFunc = function.New(&function.Spec{
	Description: `Merges all of the elements from the given maps or arrays into a single map or array`,
	Params:      []function.Parameter{},
	VarParam: &function.Parameter{
		Name:             "items",
		Type:             cty.DynamicPseudoType,
		AllowDynamicType: true,
		AllowNull:        false,
		AllowMarked:      true,
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		// empty args is accepted, so assume an empty object since we have no
		// key-value types.
		if len(args) == 0 {
			return cty.NilType, errors.New("at least one argument is required")
		}

		first := args[0].Type()
		if first.IsListType() || first.IsTupleType() {
			return AdapedConcatFuncSpec.Type(args)
		}

		return AdaptedMergeFuncSpec.Type(args)
	},
	// RefineResult: refineNonNull, // Adapted: we don't need to refine the result
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		if retType.IsListType() || retType.IsTupleType() {
			return AdapedConcatFuncSpec.Impl(args, retType)
		}
		return AdaptedMergeFuncSpec.Impl(args, retType)
	},
})

// AdaptedMergeFuncSpec is a copy of the MergeFunc spec from cty stdlib, but with the ability to merge marked
// parents into their children. This is necessary for the `merge` function to
// work as expected and sort values based on location in the source file.
var AdaptedMergeFuncSpec = &function.Spec{
	Description: `Merges all of the elements from the given maps into a single map, or the attributes from given objects into a single object.`,
	Params:      []function.Parameter{},
	VarParam: &function.Parameter{
		Name:             "maps",
		Type:             cty.DynamicPseudoType,
		AllowDynamicType: true,
		AllowNull:        true,
		AllowMarked:      true,
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		// empty args is accepted, so assume an empty object since we have no
		// key-value types.
		if len(args) == 0 {
			return cty.EmptyObject, nil
		}

		// collect the possible object attrs
		attrs := map[string]cty.Type{}

		first := cty.NilType
		matching := true
		attrsKnown := true
		for i, arg := range args {
			ty := arg.Type()
			// any dynamic args mean we can't compute a type
			if ty.Equals(cty.DynamicPseudoType) {
				return cty.DynamicPseudoType, nil
			}

			// marks are attached to values, so ignore while determining type
			arg, _ = arg.Unmark()

			switch {
			case ty.IsObjectType() && !arg.IsNull():
				for attr, aty := range ty.AttributeTypes() {
					attrs[attr] = aty
				}
			case ty.IsMapType():
				switch {
				case arg.IsNull():
					// pass, nothing to add
				case arg.IsKnown():
					ety := arg.Type().ElementType()
					for it := arg.ElementIterator(); it.Next(); {
						attr, _ := it.Element()
						attrs[attr.AsString()] = ety
					}
				default:
					// any unknown maps means we don't know all possible attrs
					// for the return type
					attrsKnown = false
				}
			}

			// record the first argument type for comparison
			if i == 0 {
				first = arg.Type()
				continue
			}

			if !ty.Equals(first) && matching {
				matching = false
			}
		}

		// the types all match, so use the first argument type
		if matching {
			return first, nil
		}

		// We had a mix of unknown maps and objects, so we can't predict the
		// attributes
		if !attrsKnown {
			return cty.DynamicPseudoType, nil
		}

		return cty.Object(attrs), nil
	},
	// RefineResult: refineNonNull, // Adapted: we don't need to refine the result
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

		outputMap := make(map[string]cty.Value)

		for _, arg := range args {
			xid := xid.New().String()
			if arg.IsNull() {
				continue
			}
			arg, argMarks := arg.Unmark()
			// Adapted: instead of marking the result with its new children, we mark the children with thier parents
			for it := arg.ElementIterator(); it.Next(); {
				k, v := it.Element()
				v, umrk := v.Unmark()
				outputMap[k.AsString()] = v.WithMarks(argMarks).Mark(&mergeMarker{umrk, xid})
			}
		}

		switch {
		case retType.IsObjectType(), retType.Equals(cty.DynamicPseudoType):
			return cty.ObjectVal(outputMap), nil
		default:
			panic(fmt.Sprintf("unexpected return type: %#v", retType))
		}
	},
}

// AdapedConcatFuncSpec is a copy of the ConcatFunc spec from cty stdlib, but with the ability to merge marked
// parents into their children. This is necessary for the `concat` function to
// work as expected and sort values based on location in the source file.
var AdapedConcatFuncSpec = &function.Spec{
	Description: `Concatenates together all of the given lists or tuples into a single sequence, preserving the input order.`,
	Params:      []function.Parameter{},
	VarParam: &function.Parameter{
		Name:        "seqs",
		Type:        cty.DynamicPseudoType,
		AllowMarked: true,
	},
	Type: func(args []cty.Value) (ret cty.Type, err error) {
		if len(args) == 0 {
			return cty.NilType, fmt.Errorf("at least one argument is required")
		}

		etys := make([]cty.Type, 0, len(args))
		for i, val := range args {
			// Discard marks for nested values, as we only need to handle types
			// and lengths.
			val, _ := val.UnmarkDeep()

			ety := val.Type()
			switch {
			case ety.IsTupleType():
				etys = append(etys, ety.TupleElementTypes()...)
			default:
				return cty.NilType, function.NewArgErrorf(
					i, "all arguments must be lists or tuples; got %s",
					ety.FriendlyName(),
				)
			}
		}
		return cty.Tuple(etys), nil
	},
	// RefineResult: refineNonNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		switch {
		case retType.IsTupleType():
			// If retType is a tuple type then we could have a mixture of
			// lists and tuples but we know they all have known values
			// (because our params don't AllowUnknown) and we know that
			// concatenating them all together will produce a tuple of
			// retType because of the work we did in the Type function above.
			vals := make([]cty.Value, 0, len(args))
			// var markses []cty.ValueMarks // remember any marked seqs we find

			for _, seq := range args {
				id := xid.New().String()
				// Adapted: instead of marking the result with its new children, we mark the children with thier parents
				seq, seqMarks := seq.Unmark()
				// Both lists and tuples support ElementIterator, so this is easy.
				it := seq.ElementIterator()
				for it.Next() {
					_, v := it.Element()
					v, umrk := v.Unmark()
					vals = append(vals, v.WithMarks(seqMarks).Mark(&mergeMarker{umrk, id}))
				}
			}

			return cty.TupleVal(vals), nil
		default:
			// should never happen if Type is working correctly above
			panic("unsupported return type")
		}
	},
}
