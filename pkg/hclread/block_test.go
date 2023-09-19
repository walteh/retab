package hclread

import (
	"context"
	"testing"

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
			  runson = "ubuntu-latest"
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
		want    []*BlockEvaluation
		wantErr error
	}{
		{
			name: "valid hcl",
			args: args{
				str: validHCL2,
			},
			want: []*BlockEvaluation{
				{
					Name:   "default.yaml",
					Schema: "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json",
					Dir:    "./.github/workflows",
					Content: map[string]interface{}{
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
								"runson": "ubuntu-latest",
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

			_, got, err := NewEvaluation(ctx, file)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
