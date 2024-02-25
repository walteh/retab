package lang_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/retab/pkg/diff"
	"github.com/walteh/retab/pkg/lang"
	"github.com/walteh/yaml"
)

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

			_, ectx, bb, diags, err := lang.NewContextFromFiles(ctx, map[string][]byte{"sampleA.input.retab": sampleAInput})
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
	})
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
		})
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

		// printer := pp.New()
		// printer.SetColoringEnabled(false)
		//		bufd := ectx.Map[lang.FilesKey].Map["task.retab"]

		// if broken == 1 {
		// 	err = afero.WriteFile(afero.NewOsFs(), "./testdata/tmp/broken.txt", []byte(printer.Sprint(bufd)), 0755)
		// 	require.NoError(t, err)
		// }

		// if ok == 1 {
		// 	err = afero.WriteFile(afero.NewOsFs(), "./testdata/tmp/ok.txt", []byte(printer.Sprint(bufd)), 0755)
		// 	require.NoError(t, err)
		// }

		// fmt.Println()
	}

	assert.Equal(t, 0, broken)

}
