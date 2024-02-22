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
TOOLS_DIR = "tools"

bins = {
	"task"          = "github.com/go-task/task/v3/cmd/task"
	"buf"           = "github.com/bufbuild/buf/cmd/buf",
	"mockery"       = "github.com/vektra/mockery/v2",
	"gotestsum"     = "gotest.tools/gotestsum",
	"golangci-lint" = "github.com/golangci/golangci-lint/cmd/golangci-lint",
}

gen taskfile {
	// schema = "https://taskfile.dev/schema.json"
	path   = "../taskfile.yaml"
	data = {
		version = 3
		tasks = merge(
			{
				for bin, zpath in bins : "${bin}-bin" => {
					generates = ["bin/${bin}"]
					dir       = "${TOOLS_DIR}"
					sources   = ["go.mod", "go.sum", "main.go"]
					cmds = [
						"go build -mod=vendor -o ./bin/${bin} ${zpath}"
					]
				}
			},
			allof("task"),
			{
				default = { deps = ["tools", "gen", "lint", "test"] }
				tools = {
					deps = [for bin, _ in bins : "${bin}-bin"]
				}
				gen = {
					deps = ["retab-gen", "buf-gen", "mockery-gen"]
				}
		})

		roll = allof("task")
	}
}


task tidy {
	cmds = [
		"go mod tidy",
		"cd ${TOOLS_DIR} && go mod tidy",
		"go work vendor"
	]
}


task update {
	cmds = [
		"go get -u -v ./... && go mod tidy",
		"cd ${TOOLS_DIR} && go get -u -v ./... && go mod tidy",
		"go work vendor"
	]
}

BUF_OUTPUT     = "a"
MOCKERY_OUTPUT = "b"

task buf-gen {
	deps      = ["buf-bin", "retab-gen"]
	generates = ["gen/buf/**/*"]
	dir       = "."
	sources   = ["proto/**/*.proto", "./${TOOLS_DIR}/bin/buf", "./${BUF_OUTPUT}"]
	cmds = [
		<<EOF
		./${TOOLS_DIR}/bin/buf generate --include-imports --include-wkt --exclude-path="./vendor" --template ${BUF_OUTPUT} --path=./proto
	EOF
		,
		"find ./gen/buf -maxdepth 5 -type f -mmin +1 -delete"
	]
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
	_, ectx, _, flebdy, diags, errd := NewContextFromFiles(ctx, map[string][]byte{"test.hcl": fle}, nil)
	assert.NoError(t, errd)
	assert.Empty(t, diags)

	blk, diags, err := NewGenBlockEvaluation(ctx, ectx, flebdy["test.hcl"])
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
