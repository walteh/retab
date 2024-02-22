package lang

import (
	"context"
	"embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/retab/pkg/diff"
)

//go:embed testdata/sampleA.input.retab
var sampleAInput []byte

//go:embed testdata/sampleA.output.yaml
var sampleAOutput []byte

func TestValidHCLDecoding(t *testing.T) {
	ctx := context.Background()
	// pp.SetDefaultMaxDepth(5)

	// load schema file
	_, ectx, bb, diags, errd := NewContextFromFiles(ctx, map[string][]byte{"sampleA.input.retab": sampleAInput})
	require.NoError(t, errd)
	require.Empty(t, diags)

	blk, diags, err := NewGenBlockEvaluation(ctx, ectx, bb)
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, err)
	require.Empty(t, diags)

	require.Len(t, blk, 1)

	out, err := blk[0].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(sampleAOutput), string(out)))

}

//go:embed testdata
var testdata embed.FS

func TestRetab3Schema(t *testing.T) {
	ctx := context.Background()
	// pp.SetDefaultMaxDepth(5)

	data, err := testdata.ReadFile("testdata/retab3.retab")
	require.NoError(t, err)

	// load schema file
	_, ectx, got, diags, errd := NewContextFromFile(ctx, data, "test.hcl")
	require.NoError(t, errd)
	require.Empty(t, diags)

	_, diags, err = NewGenBlockEvaluation(ctx, ectx, got)
	assert.NoError(t, err)
	for _, c := range diags {
		fmt.Println(c)
	}

	data, err = testdata.ReadFile("testdata/retab4.retab")
	require.NoError(t, err)

	_, ectx, got, diags, errd = NewContextFromFile(ctx, data, "test2.hcl")
	require.NoError(t, errd)
	require.Empty(t, diags)

	_, diags, err = NewGenBlockEvaluation(ctx, ectx, got)
	assert.NoError(t, err)
	for _, c := range diags {
		fmt.Println(c)
	}
	require.NotEmpty(t, diags)

}
