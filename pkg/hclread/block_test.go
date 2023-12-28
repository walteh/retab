package hclread

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
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
		name         string
		args         args
		want         *FileBlockEvaluation
		contextDiags hcl.Diagnostics
		schemaDiags  hcl.Diagnostics
	}{
		{
			name: "valid hcl",
			args: args{
				str: validHCL2,
			},
			want: &FileBlockEvaluation{
				Name:   "default.yaml",
				Schema: "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json",
				Path:   "./.github/workflows",
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
			contextDiags: hcl.Diagnostics{},
			schemaDiags:  hcl.Diagnostics{},
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

			file, err := afero.ReadFile(aferoFS, "test.hcl")
			if err != nil {
				t.Fatal(err)
			}

			_, ectx, got, diags, err := NewContextFromFile(ctx, file, "test.hcl")
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.contextDiags, diags)
			if len(diags) > 0 {
				return
			}

			be, diags, err := NewFileBlockEvaluation(ctx, ectx, got)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.schemaDiags, diags)
			if len(diags) > 0 {
				return
			}

			assert.Equal(t, tt.want.RawOutput, be.RawOutput)
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
		name         string
		args         args
		want         *FileBlockEvaluation
		contextDiags hcl.Diagnostics
		schemaDiags  hcl.Diagnostics
	}{
		{
			name: "valid hcl",
			args: args{
				str: validHCL2,
			},
			want: &FileBlockEvaluation{
				Name:   "default.yaml",
				Schema: "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json",
				Path:   "./.github/workflows",
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
			contextDiags: hcl.Diagnostics{},
			schemaDiags:  hcl.Diagnostics{},
		},
		{
			name: "valid hcl with error",
			args: args{
				str: validHCLWithError,
			},
			contextDiags: hcl.Diagnostics{},
			schemaDiags:  hcl.Diagnostics{},
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

			file, err := afero.ReadFile(aferoFS, "test.hcl")
			if err != nil {
				t.Fatal(err)
			}

			_, ectx, got, diags, err := NewContextFromFile(ctx, file, "test.hcl")
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.contextDiags, diags)
			if len(diags) > 0 {
				return
			}

			be, diags, err := NewFileBlockEvaluation(ctx, ectx, got)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.schemaDiags, diags)
			if len(diags) > 0 {
				return
			}

			assert.Equal(t, tt.want.RawOutput, be.RawOutput)

			// assert.Empty(t, resp)

		})
	}
}
