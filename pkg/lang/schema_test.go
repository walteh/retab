package lang

import (
	"context"
	"embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	require.Empty(t, diags)

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
