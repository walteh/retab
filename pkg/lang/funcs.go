package lang

import (
	"encoding/base64"
	"path/filepath"
	"strings"

	"github.com/walteh/terrors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"gitlab.com/tozd/go/errors"
)

func NewFunctionMap() map[string]function.Function {

	return map[string]function.Function{
		"jsonencode":             stdlib.JSONEncodeFunc,
		"jsondecode":             stdlib.JSONDecodeFunc,
		"csvdecode":              stdlib.CSVDecodeFunc,
		"equal":                  stdlib.EqualFunc,
		"notequal":               stdlib.NotEqualFunc,
		"format":                 stdlib.FormatFunc,
		"join":                   stdlib.JoinFunc,
		"merge":                  CombinedMergeConcatFunc,
		"length":                 stdlib.LengthFunc,
		"keys":                   stdlib.KeysFunc,
		"values":                 stdlib.ValuesFunc,
		"flatten":                stdlib.FlattenFunc,
		"contains":               stdlib.ContainsFunc,
		"index":                  stdlib.IndexFunc,
		"lookup":                 stdlib.LookupFunc,
		"element":                stdlib.ElementFunc,
		"slice":                  stdlib.SliceFunc,
		"compact":                stdlib.CompactFunc,
		"distinct":               stdlib.DistinctFunc,
		"reverselist":            stdlib.ReverseListFunc,
		"setproduct":             stdlib.SetProductFunc,
		"setunion":               stdlib.SetUnionFunc,
		"setintersection":        stdlib.SetIntersectionFunc,
		"sethaselement":          stdlib.SetHasElementFunc,
		"setsubtract":            stdlib.SetSubtractFunc,
		"setsymmetricdifference": stdlib.SetSymmetricDifferenceFunc,
		"formatdate":             stdlib.FormatDateFunc,
		"timeadd":                stdlib.TimeAddFunc,
		"add":                    stdlib.AddFunc,
		"assertnotnull":          stdlib.AssertNotNullFunc,
		"byteslen":               stdlib.BytesLenFunc,
		"byteslice":              stdlib.BytesSliceFunc,
		"not":                    stdlib.NotFunc,
		"and":                    stdlib.AndFunc,
		"or":                     stdlib.OrFunc,
		"upper":                  stdlib.UpperFunc,
		"lower":                  stdlib.LowerFunc,
		"replace":                stdlib.ReplaceFunc,
		"split":                  stdlib.SplitFunc,
		"substr":                 stdlib.SubstrFunc,
		"trimprefix":             stdlib.TrimPrefixFunc,
		"trimsuffix":             stdlib.TrimSuffixFunc,
		"trimspace":              stdlib.TrimSpaceFunc,
		"trim":                   stdlib.TrimFunc,
		"chomp":                  stdlib.ChompFunc,
		"chunklist":              stdlib.ChunklistFunc,
		"coalesce":               stdlib.CoalesceFunc,
		"indent":                 stdlib.IndentFunc,
		"title":                  stdlib.TitleFunc,
		"abs":                    stdlib.AbsoluteFunc,
		"ceil":                   stdlib.CeilFunc,
		"div":                    stdlib.DivideFunc,
		"mod":                    stdlib.ModuloFunc,
		"floor":                  stdlib.FloorFunc,
		"max":                    stdlib.MaxFunc,
		"min":                    stdlib.MinFunc,
		"mul":                    stdlib.MultiplyFunc,
		"gte":                    stdlib.GreaterThanOrEqualToFunc,
		"gt":                     stdlib.GreaterThanFunc,
		"lte":                    stdlib.LessThanOrEqualToFunc,
		"lt":                     stdlib.LessThanFunc,
		"sub":                    stdlib.SubtractFunc,
		"neg":                    stdlib.NegateFunc,
		"int":                    stdlib.IntFunc,
		"log":                    stdlib.LogFunc,
		"pow":                    stdlib.PowFunc,
		"signum":                 stdlib.SignumFunc,
		"parseint":               stdlib.ParseIntFunc,
		"range":                  stdlib.RangeFunc,
		"formatlist":             stdlib.FormatListFunc,
		"regex":                  stdlib.RegexFunc,
		"regexall":               stdlib.RegexAllFunc,
		"regexreplace":           stdlib.RegexReplaceFunc,
		"zipmap":                 stdlib.ZipmapFunc,
		"coelscelist":            stdlib.CoalesceListFunc,
		"reverse":                stdlib.ReverseFunc,
		"sort":                   stdlib.SortFunc,
		"base64encode": function.New(&function.Spec{
			Description: `Returns the Base64-encoded version of the given string.`,
			Params: []function.Parameter{
				{
					Name:             "str",
					Type:             cty.String,
					AllowUnknown:     false,
					AllowDynamicType: false,
					AllowNull:        false,
				},
			},
			Type: function.StaticReturnType(cty.String),
			// RefineResult: refineNonNull,
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if len(args) != 1 {
					return cty.NilVal, terrors.Errorf("expected 1 argument, got %d", len(args))
				}
				if args[0].IsNull() {
					return cty.StringVal(""), nil
				}

				if args[0].Type() != cty.String {
					return cty.NilVal, terrors.Errorf("expected string, got %s", args[0].GoString())
				}
				return cty.StringVal(base64.StdEncoding.EncodeToString([]byte(args[0].AsString()))), nil
			},
		}),
		"base64decode": function.New(&function.Spec{
			Description: `Returns the Base64-decoded version of the given string.`,
			Params: []function.Parameter{
				{
					Name:             "str",
					Type:             cty.String,
					AllowUnknown:     false,
					AllowDynamicType: false,
					AllowNull:        false,
				},
			},
			Type: function.StaticReturnType(cty.String),
			// RefineResult: refineNonNull,
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if len(args) != 1 {
					return cty.NilVal, errors.Errorf("expected 1 argument, got %d", len(args))
				}
				if args[0].IsNull() {
					return cty.StringVal(""), nil
				}
				if args[0].Type() != cty.String {
					return cty.NilVal, errors.Errorf("expected string, got %s", args[0].GoString())
				}
				dec, err := base64.StdEncoding.DecodeString(args[0].AsString())
				if err != nil {
					return cty.NilVal, err
				}
				return cty.StringVal(string(dec)), nil
			},
		}),
	}
}

func sanitizeFileName(str string) string {

	str = filepath.Base(str)

	if !strings.HasSuffix(str, ".retab") {
		str = str + ".retab"
	}

	// return strings.TrimPrefix(str, ".retab/")

	return str
}

// var t = cty.

var regexType = cty.Object(map[string]cty.Type{
	"key":   cty.List(cty.String),
	"regex": cty.String,
})

func dangerouslyParseRegexArgs(args cty.Value) (keys []string, regex cty.Value, err error) {
	um, _ := args.UnmarkDeep()

	m := um.AsValueMap()

	if m["regex"].IsNull() {
		err = errors.Errorf("expected map with keys 'regex' as string, got %s", m["regex"].GoString())
		return
	}

	reg := m["regex"]
	key := m["key"].AsValueSlice()

	if len(key) == 0 {
		err = errors.Errorf("expected more than 0 keys, got %s", m["key"].GoString())
		return
	}

	strs := make([]string, len(key))
	for i, v := range key {
		strs[i] = v.AsString()
	}

	return strs, reg, nil
}

func NewContextualizedFunctionMap(ectx *SudoContext, file string) map[string]function.Function {

	file = sanitizeFileName(file)

	mapp := function.New(&function.Spec{
		Description: `Returns a map of all blocks w\ the given label`,
		Params: []function.Parameter{
			{
				Name:             "block",
				Type:             cty.String,
				AllowUnknown:     true,
				AllowDynamicType: true,
				AllowNull:        false,
				AllowMarked:      true,
			},
		},
		VarParam: &function.Parameter{
			Name:             "regex",
			Type:             regexType,
			AllowUnknown:     false,
			AllowDynamicType: false,
			AllowNull:        false,
			AllowMarked:      true,
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

			unmarked, mrk := args[0].Unmark()

			fle := ectx.Root().Map[FilesKey].Map[file]

			if fle == nil {
				return cty.NilVal, errors.Errorf("file %s not found", file)
			}

			ok, err := fle.BlocksOfType(unmarked.AsString())
			if err != nil {
				return cty.NilVal, err
			}
			if len(ok) == 0 {
				return cty.NilVal, errors.Errorf("block %s not found", unmarked.AsString())
			}

			err = CheckForAnyIncompletedBlock(ok)
			if err != nil {
				return cty.NilVal, err
			}

			if len(args) > 1 && !args[1].IsNull() {

				strs, reg, err := dangerouslyParseRegexArgs(args[1])
				if err != nil {
					return cty.NilVal, err
				}

				filtered, err := FilterSudoContextWithRegex(ok, strs, reg)
				if err != nil {
					return cty.NilVal, err
				}

				ok = filtered

			}

			resp := make(map[string]cty.Value, len(ok))
			for _, v := range ok {
				resp[v.ParentKey] = v.ToValueWithExtraContext()
			}

			return cty.ObjectVal(resp).WithMarks(mrk), nil
		},
	})

	list := function.New(&function.Spec{
		Description: `Returns a list of all blocks w\ the given label`,
		Params: []function.Parameter{
			{
				Name:             "block",
				Type:             cty.String,
				AllowUnknown:     true,
				AllowDynamicType: true,
				AllowNull:        false,
				AllowMarked:      true,
			},
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

			unmarked, mrk := args[0].Unmark()

			// we do not care whether the file is complete or not, just the internal blocks
			fle := ectx.Root().Map[FilesKey].Map[file]

			if fle == nil {
				return cty.NilVal, errors.Errorf("file %s not found", file)
			}

			ok, err := fle.BlocksOfType(unmarked.AsString())
			if err != nil {
				return cty.NilVal, err
			}
			if len(ok) == 0 {
				return cty.NilVal, errors.Errorf("block %s not found", unmarked.AsString())
			}

			err = CheckForAnyIncompletedBlock(ok)
			if err != nil {
				return cty.NilVal, err
			}

			resp := make([]cty.Value, len(ok))
			for i, v := range ok {
				resp[i] = v.ToValue()
			}

			return cty.TupleVal(resp).WithMarks(mrk), nil
		}},
	)

	filed := function.New(&function.Spec{
		Description: "Returns the contents of another .retab file",
		Params: []function.Parameter{
			{
				Name:             "file",
				Type:             cty.String,
				AllowUnknown:     false,
				AllowDynamicType: false,
				AllowNull:        false,
			},
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

			unmarked, mrk := args[0].Unmark()

			fne := sanitizeFileName(unmarked.AsString())

			resp, err := CheckForCompletedBlock(ectx.Map[FilesKey], fne)
			if err != nil {
				return cty.NilVal, err
			}

			r := resp.ToValueWithExtraContext()

			return r.WithMarks(mrk), nil
		},
	})

	return map[string]function.Function{
		"file":       filed,
		"allof":      mapp,
		"alloflist":  list,
		"allofarray": list,
	}
}

func CheckForCompletedBlock(ectx *SudoContext, file string) (*SudoContext, error) {
	resp := ectx.Map[file]

	if resp == nil {
		options := []string{}
		for k := range ectx.Map {
			options = append(options, k)
		}
		return nil, errors.Errorf("block %q not found: (options: %v)", file, options)
	}

	_, ok := resp.Meta.(*IncomleteBlockMeta)
	if ok {
		return nil, errors.Errorf("the block %q is not complete", file)
	}
	return resp, nil
}

func CheckForAnyIncompletedBlock(ectx []*SudoContext) error {

	for _, v := range ectx {
		// fmt.Println(v.ParentKey, reflect.TypeOf(v.Meta).String())
		_, okd := v.Meta.(*IncomleteBlockMeta)
		if okd {
			return errors.Errorf("the block %s is not complete", v.ParentKey)
		}
	}

	return nil
}

func NewDynamicContextualizedFunctionMap(ectx *SudoContext) map[string]function.Function {
	// takes in some negative number and returns the nested parent -x levels
	selfer := function.New(&function.Spec{
		Description: `Returns the parent block of the current block`,
		Params:      []function.Parameter{},
		VarParam: &function.Parameter{
			Name: "levels",
			Type: cty.Number,
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

			if len(args) != 1 {
				// default to 0
				args = append(args, cty.NumberIntVal(0))
			}

			num := args[0].AsBigFloat()
			if !num.IsInt() {
				return cty.NilVal, errors.Errorf("expected int, got %s", args[0].GoString())
			}

			count, _ := num.Int64()

			if count > 0 {
				return cty.NilVal, errors.Errorf("expected negative int, got %s", args[0].GoString())
			}

			wrk := ectx

			for range count * -1 {
				if wrk.Parent == nil {
					return cty.NilVal, nil
				}
				wrk = wrk.Parent
			}

			return wrk.ToValue(), nil
		},
	})

	return map[string]function.Function{
		"self": selfer,
	}
}
