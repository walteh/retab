package yamlfmt_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/walteh/retab/v2/gen/mocks/pkg/formatmock"
	"github.com/walteh/retab/v2/pkg/diff"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/formatters/yamlfmt"
)

func formatYaml(ctx context.Context, cfg format.Configuration, src []byte) (string, error) {
	formatter := yamlfmt.NewFormatter()
	reader, err := formatter.Format(ctx, cfg, bytes.NewReader(src))
	if err != nil {
		return "", err
	}

	result, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func TestYaml(t *testing.T) {
	start := `
cmds:
    - for:
        var: WASM_OUT_DIRS
        cmd: echo "{{.ITEM}}"
        friend: 
            sup: joe
`

	expected := `
cmds:
  - for:
        var: WASM_OUT_DIRS
        cmd: echo "{{.ITEM}}"
        friend:
            sup: joe
`

	cfg := formatmock.NewMockConfiguration(t)
	cfg.EXPECT().UseTabs().Return(false).Maybe()
	cfg.EXPECT().IndentSize().Return(4).Maybe()
	cfg.EXPECT().Raw().Return(map[string]string{
		"cmds": "cmds",
	}).Maybe()

	actual, err := formatYaml(t.Context(), cfg, []byte(start))
	if err != nil {
		t.Fatal(err)
	}

	diff.Require(t).Want(expected).Got(actual).Equals()
}
