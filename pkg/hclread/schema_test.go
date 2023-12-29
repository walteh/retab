package hclread

import (
	"context"
	"embed"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/walteh/yaml"
)

const validHCL = `
gen "default" {
	path = "./.github/workflows/def.yaml"
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
	assert.Empty(t, diags)

	blk, diags, err := NewGenBlockEvaluation(ctx, ectx, got)
	if err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, err)
	assert.Empty(t, diags)

	out, erry := yaml.Marshal(blk.RawOutput)
	if erry != nil {
		t.Fatal(erry)
	}

	t.Log(string(out))

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
	assert.Empty(t, diags)

	_, diags, err = NewGenBlockEvaluation(ctx, ectx, got)
	assert.NoError(t, err)
	for _, c := range diags {
		fmt.Println(c)
	}
	assert.Empty(t, diags)

}
