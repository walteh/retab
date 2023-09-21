// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

// type PathReader struct {
// }

// var _ decoder.PathReader = &PathReader{}

// func (mr *PathReader) Paths(ctx context.Context) []lang.Path {
// 	paths := make([]lang.Path, 0)

// 	// langId, _ := LanguageId(ctx)

// 	// paths = append(paths, lang.Path{
// 	// 	Path:       mod.Path,
// 	// 	LanguageID: langId.String(),
// 	// })

// 	paths = append(paths, lang.Path{
// 		Path:       ".",
// 		LanguageID: lsp.Retab.String(),
// 	})
// 	// if len(mod.ParsedVarsFiles) > 0 {
// 	// 	paths = append(paths, lang.Path{
// 	// 		Path:       mod.Path,
// 	// 		LanguageID: lsp.fvars.String(),
// 	// 	})
// 	// }

// 	return paths
// }

// func (mr *PathReader) PathContext(path lang.Path) (*decoder.PathContext, error) {

// 	// fmt.Errorf("unknown language ID: %q", path.LanguageID)
// 	return varsPathContext()
// }
