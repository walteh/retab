// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
)

type PathReader struct {
}

var _ decoder.PathReader = &PathReader{}

func (mr *PathReader) Paths(ctx context.Context) []lang.Path {
	paths := make([]lang.Path, 0)

	return paths
}

func (mr *PathReader) PathContext(path lang.Path) (*decoder.PathContext, error) {

	return nil, fmt.Errorf("unknown language ID: %q", path.LanguageID)
}
