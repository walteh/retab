package hclwrite_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/tftab/gend/mockery"
	"github.com/walteh/tftab/pkg/hclwrite"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name       string
		useTabs    bool
		indentSize int
		src        []byte
		expected   []byte
	}{
		{
			name:       "Use Tabs with IndentSize 1",
			useTabs:    true,
			indentSize: 1,
			src: []byte(`
variable "DESTDIR" {
  default = "./bin"
  required = true
}`),
			expected: []byte(`
variable "DESTDIR" {
	default  = "./bin"
	required = true
}`),
		},
		{
			name:       "Use Spaces with IndentSize 4",
			useTabs:    false,
			indentSize: 4,
			src: []byte(`
variable "DESTDIR" {
  default = "./bin"
  required = true
}`),
			expected: []byte(`
variable "DESTDIR" {
    default  = "./bin"
    required = true
}`),
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &mockery.MockConfigurationProvider{}
			cfg.EXPECT().UseTabs().Return(tt.useTabs)
			cfg.EXPECT().IndentSize().Return(tt.indentSize)

			// Call the Format function with the provided configuration and source
			result, err := hclwrite.FormatBytes(cfg, tt.src)

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
			assert.Equal(t, tt.expected, buf.Bytes(), "HCL source does not match expected output")
		})
	}
}
