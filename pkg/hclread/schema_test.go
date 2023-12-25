package hclread

import (
	"context"
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/walteh/yaml"
)

const validHCL = `
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
					  run  = <<SHELL
					  echo "Hello world"
				  SHELL
				  },
			  ]
			}
		}
	}
}
`

func TestValidHCLDecoding(t *testing.T) {
	ctx := context.Background()
	// pp.SetDefaultMaxDepth(5)

	aferoFS := afero.NewMemMapFs()

	err := afero.WriteFile(aferoFS, "test.hcl", []byte(validHCL), 0644)
	if err != nil {
		t.Fatal(err)
	}

	fle, err := aferoFS.Open("test.hcl")
	if err != nil {
		t.Fatal(err)
	}

	defer fle.Close()

	// load schema file
	bd, ectx, got, errd := NewEvaluation(ctx, fle)
	assert.NoError(t, errd)

	for _, b := range got.Blocks {

		blk, err := NewFileBlockEvaluation(ctx, ectx, b, bd.Body, false)
		if err != nil {
			t.Fatal(err)
		}

		if blk.Validation != nil {
			for _, v := range blk.Validation {
				for _, err := range v.Problems {
					pp.Println(err)
				}
			}
			t.Fatal(blk.Validation)
		}

		out, erry := yaml.Marshal(blk.RawOutput)
		if erry != nil {
			t.Fatal(erry)
		}

		t.Log(string(out))
	}

}
