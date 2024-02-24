package lang

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// ChildMarkMergeFunc is a copy of the MergeFunc from cty stdlib, but with the ability to merge marked
// parents into their children. This is necessary for the `merge` function to
// work as expected and sort values based on location in the source file.
var ChildMarkMergeFunc = function.New(&function.Spec{
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

			// check for invalid arguments
			if !ty.IsMapType() && !ty.IsObjectType() {
				return cty.NilType, fmt.Errorf("arguments must be maps or objects, got %#v", ty.FriendlyName())
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
		var markses []cty.ValueMarks // remember any marked maps/objects we find

		for _, arg := range args {
			if arg.IsNull() {
				continue
			}
			arg, argMarks := arg.Unmark()
			// Adapted: instead of marking the result with its new children, we mark the children with thier parents
			for it := arg.ElementIterator(); it.Next(); {
				k, v := it.Element()
				outputMap[k.AsString()] = v.WithMarks(argMarks)
			}
		}

		switch {
		case retType.IsMapType():
			if len(outputMap) == 0 {
				return cty.MapValEmpty(retType.ElementType()).WithMarks(markses...), nil
			}
			return cty.MapVal(outputMap).WithMarks(markses...), nil
		case retType.IsObjectType(), retType.Equals(cty.DynamicPseudoType):
			return cty.ObjectVal(outputMap).WithMarks(markses...), nil
		default:
			panic(fmt.Sprintf("unexpected return type: %#v", retType))
		}
	},
})
