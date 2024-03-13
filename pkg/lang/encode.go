package lang

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/walteh/yaml"
	"github.com/zclconf/go-cty/cty"
)

type Sorter struct {
	Range      []hcl.Range
	MergeRange []mergeRange
	Key        string
	Value      cty.Value
	Ignored    bool
	KnownMarks []any
}

type SortEnhancer interface {
	Enhance(cty.Value, []any) (cty.Value, error)
}

type ignoreFromYaml struct{}

type mergeRange struct {
	rng  hcl.Range
	self string
}

type makePathRelative struct {
	to string
}

type isIncompleteBlock struct {
}

func (me *SudoContext) ToYAML() (yaml.MapSlice, error) {
	resp := yaml.MapSlice{}

	item, err := UnmarkToSortedArray(me.ToValue(), me.ParentBlockMeta())
	if err != nil {
		return nil, err
	}

	if x, ok := item.(yaml.MapSlice); ok {
		resp = append(resp, x...)
	}

	if x, ok := item.(yaml.MapItem); ok {
		resp = append(resp, x)
	}

	return resp, nil
}

func NewSorter(key string, me cty.Value) *Sorter {

	me, r := me.Unmark()

	knownMarks := make([]any, 0)
	ranges := make([]hcl.Range, 0, len(r))
	mergedRanges := make([]mergeRange, 0)
	ignored := false

	if len(r) == 0 {
		// panic(fmt.Sprintf("no range or ignore found for %s", me.GoString()))
		return &Sorter{Key: key, Value: me, Ignored: false, Range: ranges, MergeRange: mergedRanges, KnownMarks: knownMarks}
	}

	for z := range r {
		switch e := z.(type) {
		case hcl.Range:
			ranges = append(ranges, e)
		case *ignoreFromYaml:
			ignored = true
		case *mergeMarker:
			for z := range e.ref {
				switch g := z.(type) {
				case hcl.Range:
					mergedRanges = append(mergedRanges, mergeRange{g, e.common})
				}
			}
		case *makePathRelative:
			knownMarks = append(knownMarks, e)
		default:
			fmt.Println("unknown mark", e)
			knownMarks = append(knownMarks, e)
		}
	}

	slices.SortFunc(mergedRanges, func(a, b mergeRange) int {
		if a.self == b.self {
			if a.rng.Start.Line == b.rng.Start.Line {
				if a.rng.Start.Column == b.rng.Start.Column {
					return 0
				}
				return a.rng.Start.Column - b.rng.Start.Column
			}
			return a.rng.Start.Line - b.rng.Start.Line
		}
		return 0
	})

	slices.SortFunc(ranges, func(a, b hcl.Range) int {
		if a.Start.Line == b.Start.Line {
			return a.Start.Byte - b.Start.Byte
		}
		return a.Start.Line - b.Start.Line
	})

	return &Sorter{Range: ranges, Key: key, Ignored: ignored, Value: me, MergeRange: mergedRanges, KnownMarks: knownMarks}
}

func NewSorterList(val cty.Value) (*Sorter, []*Sorter) {
	self := NewSorter("", val)
	out := make([]*Sorter, 0)

	if self.Value.Type().IsObjectType() {
		objs := self.Value.AsValueMap()
		for k, v := range objs {
			rng := NewSorter(k, v)
			out = append(out, rng)
		}
	} else if self.Value.Type().IsTupleType() {
		objs := self.Value.AsValueSlice()
		for _, v := range objs {
			rng := NewSorter("", v)
			out = append(out, rng)
		}
	} else {
		return self, out
	}

	slices.SortFunc(out, func(a, b *Sorter) int {

		for i, x := range a.MergeRange {
			if i >= len(b.MergeRange) {
				continue
			}
			y := b.MergeRange[i]
			if x.self == y.self {
				if x.rng.Start.Line == y.rng.Start.Line {
					if x.rng.Start.Column == y.rng.Start.Column {
						continue
					}
					return x.rng.Start.Column - y.rng.Start.Column
				}
				return x.rng.Start.Line - y.rng.Start.Line
			}
		}

		if len(a.Range) == 0 {
			if len(b.Range) == 0 {
				return 0
			} else {
				return -1
			}
		}

		for i, x := range a.Range {
			if i >= len(b.Range) {
				return 1
			}
			y := b.Range[i]
			if x.Start.Line == y.Start.Line {
				if x.Start.Column == y.Start.Column {
					continue
				}
				return x.Start.Column - y.Start.Column
			}
			return x.Start.Line - y.Start.Line
		}

		return 0
	})

	return self, out
}

func UnmarkToSortedArray(me cty.Value, enhance SortEnhancer) (any, error) {

	self, out := NewSorterList(me)
	if self.Ignored {
		return nil, nil
	}

	if len(out) == 0 {
		return noMetaJsonEncode(self.Value)
	}

	if me.Type().IsObjectType() {

		wrk := make(yaml.MapSlice, 0, len(out))
		for _, v := range out {
			if v.Key == MetaKey || strings.Contains(v.Key, FuncKey) || v.Ignored {
				continue
			}
			res, err := UnmarkToSortedArray(v.Value, enhance)
			if err != nil {
				return nil, err
			}

			wrk = append(wrk, yaml.MapItem{Key: v.Key, Value: res})
		}
		return wrk, nil
	}

	if me.Type().IsTupleType() {
		wrk := make([]any, 0, len(out))
		for _, v := range out {
			if v.Ignored {
				continue
			}
			res, err := UnmarkToSortedArray(v.Value, enhance)
			if err != nil {
				return nil, err
			}
			wrk = append(wrk, res)
		}
		return wrk, nil
	}

	if len(out) != 1 {
		return nil, fmt.Errorf("not a list or map")
	}

	// TODO - this does not work, we will need to more string template handling to properly pull this off

	// if enhance != nil {
	// 	fmt.Println("enhancing", out[0].KnownMarks)
	// 	res, err := enhance.Enhance(out[0].Value, out[0].KnownMarks)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	fmt.Println("enhancing", res)

	// 	out[0].Value = res
	// }

	return nil, fmt.Errorf("not a list or map")
}
