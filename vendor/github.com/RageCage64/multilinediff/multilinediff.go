package multilinediff

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Get the diff between two strings.
func Diff(a, b, lineSep string) (string, int) {
	reporter := Reporter{LineSep: lineSep}
	cmp.Diff(
		a, b,
		cmpopts.AcyclicTransformer("multiline", func(s string) []string {
			return strings.Split(s, lineSep)
		}),
		cmp.Reporter(&reporter),
	)
	return reporter.String(), reporter.DiffCount
}
