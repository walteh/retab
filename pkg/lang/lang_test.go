package lang_test

import (
	"context"
	"embed"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/retab/pkg/diff"
	"github.com/walteh/retab/pkg/lang"
	"github.com/walteh/yaml"
)

//go:embed testdata/issue22/*
var issue22_files embed.FS

//go:embed testdata/issue22/task.retab
var issue22_input []byte

//go:embed testdata/issue22/task.expected.yaml
var issue22_expected []byte

func TestIssue22(t *testing.T) {

	ctx := context.Background()

	fs := afero.FromIOFS{FS: issue22_files}

	bse := afero.NewBasePathFs(fs, "testdata/issue22/sample")

	env, err := lang.LoadGlobalEnvVars(bse, &lang.LoadGlobalEnvVarOpts{
		GoModFileName:  "go.modz",
		DotEnvFileName: "dotenv",
	})
	require.NoError(t, err)

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue22_input,
	}, env)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["tmp.yaml"])

	out, err := blk["tmp.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue22_expected), string(out)))

}

//go:embed testdata/sampleA.input.retab
var sampleAInput []byte

//go:embed testdata/sampleA.expected.yaml
var sampleAExpected []byte

//go:embed testdata/sampleB.input.retab
var sampleBInput []byte

//go:embed testdata/sampleB.expected.yaml
var sampleBExpected []byte

func TestEncoding(t *testing.T) {

	type caseT struct {
		name     string
		input    []byte
		expected []byte
	}

	cases := []caseT{
		{"A", sampleAInput, sampleAExpected},
		// {"B", sampleBInput, sampleBExpected},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			ctx := context.Background()

			_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{"sampleA.input.retab": sampleAInput}, nil)
			require.NoError(t, err)
			for _, c := range diags {
				t.Log(c)
			}
			require.Empty(t, diags)

			blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
			require.NoError(t, err)
			for _, c := range diags {
				t.Log(c)
			}
			require.Empty(t, diags)

			require.Len(t, blk, 1)
			require.NotNil(t, blk["sampleA.actual.yaml"])

			out, err := blk["sampleA.actual.yaml"].Encode()
			require.NoError(t, err)

			require.Empty(t, diff.DiffExportedOnly(string(sampleAExpected), string(out)))
		})
	}
}

func TestFileRef(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"sampleA.input.retab": sampleAInput,
		"sampleB.input.retab": sampleBInput,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	for range 1000 {

		blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
		require.NoError(t, err)
		for _, c := range diags {
			t.Log(c)
		}
		require.Empty(t, diags)

		require.Len(t, blk, 2)
		require.NotNil(t, blk["sampleA.actual.yaml"])
		require.NotNil(t, blk["sampleB.actual.yaml"])

		out, err := blk["sampleB.actual.yaml"].Encode()
		require.NoError(t, err)

		require.Empty(t, diff.DiffExportedOnly(string(sampleBExpected), string(out)))
	}
}

//go:embed testdata/issue31/task.retab
var issue31_taskInput []byte

//go:embed testdata/issue31/mockery.retab
var issue31_mockeryInput []byte

//go:embed testdata/issue31/buf.retab
var issue31_bufInput []byte

func TestIssue31(t *testing.T) {

	ctx := context.Background()
	ok := 0
	broken := 0
	for range 25 {

		_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
			"mockery.retab": issue31_mockeryInput,
			"buf.retab":     issue31_bufInput,
			"task.retab":    issue31_taskInput,
		}, nil)
		require.NoError(t, err)
		for _, c := range diags {
			t.Log(c)
		}
		require.Empty(t, diags)

		blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
		require.NoError(t, err)
		for _, c := range diags {
			t.Log(c)
		}
		require.Empty(t, diags)

		require.Len(t, blk, 3)
		require.NotNil(t, blk["taskfile.yaml"])

		tsks := yaml.MapSlice{}

		for _, v := range blk["taskfile.yaml"].OrderedOutput {
			if v.Key == "tasks" {
				tsks = v.Value.(yaml.MapSlice)
				break
			}
		}

		if len(tsks) != 15 {
			broken++
		} else {
			ok++
		}
	}

	assert.Equal(t, 0, broken)

}

//go:embed testdata/issue33/mockery.retab
var issue33_input []byte

//go:embed testdata/issue33/mockery.expected.yaml
var issue33_expected []byte

func TestIssue33(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"mockery.retab": issue33_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk[".mockery.yaml"])

	out, err := blk[".mockery.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue33_expected), string(out)))

}

//go:embed testdata/issue34/task.retab
var issue34_input []byte

//go:embed testdata/issue34/task.expected.yaml
var issue34_expected []byte

func TestIssue34(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue34_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["taskfile.yaml"])

	out, err := blk["taskfile.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue34_expected), string(out)))

}

//go:embed testdata/issue38/task.retab
var issue38_input []byte

//go:embed testdata/issue38/task.expected.yaml
var issue38_expected []byte

func TestIssue38(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue38_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["some/nested/new.yaml"])
	require.NotNil(t, blk["some/nested/dir/tmp.yaml"])

	out, err := blk["some/nested/new.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue38_expected), string(out)))

}

//go:embed testdata/issue39/task.retab
var issue39_input []byte

//go:embed testdata/issue39/task.expected.yaml
var issue39_expected []byte

func TestIssue39(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue39_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["taskfile.yaml"])

	out, err := blk["taskfile.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue39_expected), string(out)))

}

//go:embed testdata/issue41/task.retab
var issue41_input []byte

//go:embed testdata/issue41/task.expected.yaml
var issue41_expected []byte

func TestIssue41(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue41_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["tmp.yaml"])

	out, err := blk["tmp.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue41_expected), string(out)))

}

//go:embed testdata/issue43/task.retab
var issue43_input []byte

//go:embed testdata/issue43/task.expected.yaml
var issue43_expected []byte

func TestIssue43(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue43_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["some/nested/new.yaml"])
	require.NotNil(t, blk["some/nested/turkey/bacon.yaml"])
	require.NotNil(t, blk["some/nested/chicken/egg.yaml"])

	out, err := blk["some/nested/new.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue43_expected), string(out)))

}

//go:embed testdata/issue44/task.retab
var issue44_input []byte

//go:embed testdata/issue44/task.expected.yaml
var issue44_expected []byte

func TestIssue44(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue44_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["out/file.yaml"])

	out, err := blk["out/file.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue44_expected), string(out)))

}

//go:embed testdata/issue46/task.retab
var issue46_input []byte

//go:embed testdata/issue46/task.expected.yaml
var issue46_expected []byte

func TestIssue46(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue46_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["out/file.yaml"])

	out, err := blk["out/file.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue46_expected), string(out)))

}

//go:embed testdata/issue51/task.retab
var issue51_input []byte

//go:embed testdata/issue51/task.expected.yaml
var issue51_expected []byte

func TestIssue51(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"task.retab": issue51_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["out/file.yaml"])

	out, err := blk["out/file.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue51_expected), string(out)))

}

//go:embed testdata/issue53/gha.retab
var issue53_input []byte

//go:embed testdata/issue53/gha.expected.yaml
var issue53_expected []byte

func TestIssue53(t *testing.T) {

	ctx := context.Background()

	_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{
		"gha.retab": issue53_input,
	}, nil)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	blk, diags, err := lang.NewGenBlockEvaluation(ctx, ectx, bb)
	require.NoError(t, err)
	for _, c := range diags {
		t.Log(c)
	}
	require.Empty(t, diags)

	require.NotNil(t, blk["out/file.yaml"])

	out, err := blk["out/file.yaml"].Encode()
	require.NoError(t, err)

	require.Empty(t, diff.DiffExportedOnly(string(issue53_expected), string(out)))

}
