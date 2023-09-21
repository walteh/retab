// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"io/fs"
	"path"
	"path/filepath"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/walteh/retab/internal/lsp/filesystem"
)

func TestDecoder_CodeLensesForFile_concurrencyBug(t *testing.T) {

	testCfg := `data "terraform_remote_state" "vpc" { }
`
	dirNames := []string{"testdir1", "testdir2"}

	mapFs := fstest.MapFS{}
	for _, dirName := range dirNames {
		mapFs[dirName] = &fstest.MapFile{Mode: fs.ModeDir}
		mapFs[path.Join(dirName, "main.tf")] = &fstest.MapFile{Data: []byte(testCfg)}
		mapFs[filepath.Join(dirName, "main.tf")] = &fstest.MapFile{Data: []byte(testCfg)}
	}

	ctx := context.Background()

	d := decoder.NewDecoder(filesystem.NewFilesystem(filesystem.DocumentStore))

	var wg sync.WaitGroup
	for _, dirName := range dirNames {
		dirName := dirName
		wg.Add(1)
		go func(t *testing.T) {
			defer wg.Done()
			_, err := d.CodeLensesForFile(ctx, lang.Path{
				Path:       dirName,
				LanguageID: "retab",
			}, "main.tf")
			if err != nil {
				t.Error(err)
			}
		}(t)
	}
	wg.Wait()
}

func gzipCompressBytes(t *testing.T, b []byte) []byte {
	var compressedBytes bytes.Buffer
	gw := gzip.NewWriter(&compressedBytes)
	_, err := gw.Write(b)
	if err != nil {
		t.Fatal(err)
	}
	err = gw.Close()
	if err != nil {
		t.Fatal(err)
	}
	return compressedBytes.Bytes()
}

var tfSchemaJSON = `{
	"format_version": "1.0",
	"provider_schemas": {
		"terraform.io/builtin/terraform": {
			"data_source_schemas": {
				"terraform_remote_state": {
					"version": 0,
					"block": {
						"attributes": {
							"backend": {
								"type": "string",
								"description": "The remote backend to use, e.g. remote or http.",
								"description_kind": "markdown",
								"required": true
							},
							"config": {
								"type": "dynamic",
								"description": "The configuration of the remote backend. Although this is optional, most backends require some configuration.\n\nThe object can use any arguments that would be valid in the equivalent terraform { backend \"\u003cTYPE\u003e\" { ... } } block.",
								"description_kind": "markdown",
								"optional": true
							},
							"defaults": {
								"type": "dynamic",
								"description": "Default values for outputs, in case the state file is empty or lacks a required output.",
								"description_kind": "markdown",
								"optional": true
							},
							"outputs": {
								"type": "dynamic",
								"description": "An object containing every root-level output in the remote state.",
								"description_kind": "markdown",
								"computed": true
							},
							"workspace": {
								"type": "string",
								"description": "The Terraform workspace to use, if the backend supports workspaces.",
								"description_kind": "markdown",
								"optional": true
							}
						},
						"description_kind": "plain"
					}
				}
			}
		}
	}
}`
