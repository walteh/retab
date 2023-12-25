package hclread

import (
	"context"
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const validHCL2 = `
file "default.yaml" {
	dir = "./.github/workflows"
	schema = "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json"
	data = {
		name = "test"
		on = {
		  push = {
			  branches = ["main"]
		  }
	  }
		jobs = {
		  build = {
			  runs-on = "ubuntu-latest"
			  steps = [
				  {
					  name = "Checkout"
					  uses = "actions/checkout@v2"
				  },
				  {
					  name = "Run tests"
					  run = "make test"
				  },
			  ]
			  }

		}
	}
}
`

func TestParseBlocksFromFile(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    *FileBlockEvaluation
		wantErr error
	}{
		{
			name: "valid hcl",
			args: args{
				str: validHCL2,
			},
			want: &FileBlockEvaluation{
				Name:   "default.yaml",
				Schema: "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json",
				Dir:    "./.github/workflows",
				RawOutput: map[string]interface{}{
					"name": "test",
					"on": map[string]interface{}{
						"push": map[string]interface{}{
							"branches": []interface{}{
								"main",
							},
						},
					},
					"jobs": map[string]interface{}{
						"build": map[string]interface{}{
							"runs-on": "ubuntu-latest",
							"steps": []interface{}{
								map[string]interface{}{
									"name": "Checkout",
									"uses": "actions/checkout@v2",
								},
								map[string]interface{}{
									"name": "Run tests",
									"run":  "make test",
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()

			aferoFS := afero.NewMemMapFs()

			err := afero.WriteFile(aferoFS, "test.hcl", []byte(validHCL2), 0644)
			if err != nil {
				t.Fatal(err)
			}

			file, err := aferoFS.Open("test.hcl")
			if err != nil {
				t.Fatal(err)
			}

			f, ectx, got, err := NewEvaluation(ctx, file)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			var resp *FileBlockEvaluation

			for _, block := range got.Blocks {
				if block.Type != "file" {
					continue
				}
				be, err := NewFileBlockEvaluation(ctx, ectx, block, f.Body, false)
				if tt.wantErr == nil {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}

				resp = be
				break
			}

			assert.Equal(t, tt.want.RawOutput, resp.RawOutput)
		})
	}
}

const validHCLWithReference = `

file "default.yaml" {
	dir = "./.github/workflows"
	schema = "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json"
	data = {
		name = "test"
		on = {
		  push = {
			  branches = ["main"]
		  }
	  }
		jobs = {
		  build = {
			  runs-on = "ubuntu-latest"
			  steps = [
				  step.checkout,
				  {
					  name = "Run tests"
					  run = "make test"
				  },
			  ]
			  }

		}
	}
}

step "checkout" {
	name = "Checkout"
	uses = "actions/checkout@v2"
}
`

const validHCLWithError = `
file "default.yaml" {
	dir = "./.github/workflows"
	schema = "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json"
	data = {
		name = "test"
	}
}
`

func TestParseBlocksWithReference(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    *FileBlockEvaluation
		wantErr bool
	}{
		{
			name: "valid hcl",
			args: args{
				str: validHCL2,
			},
			want: &FileBlockEvaluation{
				Name:   "default.yaml",
				Schema: "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json",
				Dir:    "./.github/workflows",
				RawOutput: map[string]interface{}{
					"name": "test",
					"on": map[string]interface{}{
						"push": map[string]interface{}{
							"branches": []interface{}{
								"main",
							},
						},
					},
					"jobs": map[string]interface{}{
						"build": map[string]interface{}{
							"runs-on": "ubuntu-latest",
							"steps": []interface{}{
								map[string]interface{}{
									"name": "Checkout",
									"uses": "actions/checkout@v2",
								},
								map[string]interface{}{
									"name": "Run tests",
									"run":  "make test",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid hcl with error",
			args: args{
				str: validHCLWithError,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()

			aferoFS := afero.NewMemMapFs()

			err := afero.WriteFile(aferoFS, "test.hcl", []byte(tt.args.str), 0644)
			if err != nil {
				t.Fatal(err)
			}

			file, err := aferoFS.Open("test.hcl")
			if err != nil {
				t.Fatal(err)
			}

			_, ectx, got, err := NewEvaluation(ctx, file)
			if err != nil {
				t.Fatal(err)
			}
			be, err := NewFullEvaluation(ctx, ectx, got, false, "somefile")
			if err != nil {
				t.Fatal(err)
			}

			pp.SetDefaultMaxDepth(20)
			pp.Println(be.File.Validation)
			if tt.wantErr {
				assert.True(t, len(be.File.Validation) > 0)
			} else {
				assert.Equal(t, tt.want.RawOutput, be.File.RawOutput)
			}

			// assert.Empty(t, resp)

		})
	}
}
