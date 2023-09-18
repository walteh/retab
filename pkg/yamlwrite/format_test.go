package yamlwrite_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/retab/gen/mockery"
	"github.com/walteh/retab/pkg/yamlwrite"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name                   string
		useTabs                bool
		indentSize             int
		trimMultipleEmptyLines bool
		src                    []byte
		expected               []byte
	}{
		{
			name:                   "Use Tabs with IndentSize 1",
			useTabs:                true,
			indentSize:             1,
			trimMultipleEmptyLines: false,
			src: []byte(`
sup:
  hi: {there: true}
  ok: true`),
			expected: []byte(`
sup: {
	hi: there,
	ok: true
}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &mockery.MockProvider_configuration{}
			cfg.EXPECT().UseTabs().Return(tt.useTabs)
			cfg.EXPECT().IndentSize().Return(tt.indentSize)
			cfg.EXPECT().TrimMultipleEmptyLines().Return(tt.trimMultipleEmptyLines)

			fmt := yamlwrite.NewYamlFormatter()

			ctx := context.Background()

			// Call the Format function with the provided configuration and source
			result, err := fmt.Format(ctx, cfg, bytes.NewReader(tt.src))

			// Check for errors
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Read the result into a buffer
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(result)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Compare the result with the expected outcome
			assert.Equal(t, string(tt.expected), buf.String(), "HCL source does not match expected output")
		})
	}
}
