package lang

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func sanatizeGenPath(path string) string {
	// we want it to work whether or not the user is poining to the current directory, or the retab folder
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "../")
	path = strings.TrimPrefix(path, "/")
	return path
}

func NewRefFunctionFromPath(ctx context.Context, start string) function.Function {
	typ := cty.Object(map[string]cty.Type{"path": cty.String})
	return function.New(&function.Spec{
		Description: fmt.Sprintf(`Returns the path of the given file or directory, relative to the %s.`, start),
		Params: []function.Parameter{
			{
				Name:             "input",
				Type:             typ,
				AllowUnknown:     false,
				AllowDynamicType: false,
				AllowNull:        false,
			},
		},
		Type: function.StaticReturnType(cty.String),
		// RefineResult: refineNonNull,
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			// if len(args) != 1 {
			// 	return cty.NilVal, terrors.Errorf("expected 1 argument, got %d", len(args))
			// }
			// if args[0].IsNull() {
			// 	return cty.StringVal(""), nil
			// }

			// errs := args[0].Type().TestConformance(typ)
			// if len(errs) > 0 {
			// 	return cty.NilVal, terrors.Errorf("expected %s, got %s", typ.GoString(), args[0].Type().GoString())
			// }

			vals := args[0].AsValueMap()

			path := vals["path"]
			if path.IsNull() {
				return cty.NilVal, fmt.Errorf("expected path to be set")
			}

			val := sanatizeGenPath(path.AsString())
			start := sanatizeGenPath(start)

			rel, err := filepath.Rel(start, val)
			if err != nil {
				return cty.NilVal, err
			}

			return cty.StringVal(rel), nil
		},
	})
}
