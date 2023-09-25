// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nonewvars_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"github.com/walteh/retab/gen/gopls/analysis/nonewvars"
	"github.com/walteh/retab/gen/gopls/typeparams"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	tests := []string{"a"}
	if typeparams.Enabled {
		tests = append(tests, "typeparams")
	}
	analysistest.RunWithSuggestedFixes(t, testdata, nonewvars.Analyzer, tests...)
}
