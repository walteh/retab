package hclread

import (
	"context"
	"embed"
	"testing"

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

	fle, err := afero.ReadFile(aferoFS, "test.hcl")
	if err != nil {
		t.Fatal(err)
	}

	// load schema file
	_, ectx, got, diags, errd := NewContextFromFile(ctx, fle, "test.hcl")
	assert.NoError(t, errd)
	assert.NoError(t, diags)

	ran := false

	for _, b := range got.Blocks {
		if b.Type != "file" {
			continue
		}

		ran = true

		blk, diags, err := NewGenBlockEvaluation(ctx, ectx, got)
		if err != nil {
			t.Fatal(err)
		}

		assert.NoError(t, err)
		assert.NoError(t, diags)

		out, erry := yaml.Marshal(blk.RawOutput)
		if erry != nil {
			t.Fatal(erry)
		}

		t.Log(string(out))
	}

	assert.True(t, ran)

}

//go:embed testdata
var testdata embed.FS

func TestRetab3Schema(t *testing.T) {
	ctx := context.Background()
	// pp.SetDefaultMaxDepth(5)

	data, err := testdata.ReadFile("testdata/retab3.retab")
	assert.NoError(t, err)

	// load schema file
	_, ectx, got, diags, errd := NewContextFromFile(ctx, data, "test.hcl")
	assert.NoError(t, errd)
	assert.NoError(t, diags)

	_, diags, err = NewGenBlockEvaluation(ctx, ectx, got)
	assert.NoError(t, err)
	assert.NoError(t, diags)

}
