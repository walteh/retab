package hclfmt_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/retab/gen/mockery"
	"github.com/walteh/retab/pkg/format/hclfmt"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name                   string
		useTabs                bool
		indentSize             int
		trimMultipleEmptyLines bool
		oneBracketPerLine      bool
		src                    []byte
		expected               []byte
	}{
		{
			name:                   "Use Tabs with IndentSize 1",
			useTabs:                true,
			indentSize:             1,
			trimMultipleEmptyLines: false,
			oneBracketPerLine:      true,
			src: []byte(`
variable "DESTDIR" {
  default = "./bin"
  required = true
  ok = [{abc = 1}]
}`),
			expected: []byte(`
variable "DESTDIR" {
	default  = "./bin"
	required = true
	ok = [
		{
			abc = 1
		}
	]
}
`),
		},
		{
			name:                   "Use Spaces with IndentSize 4",
			useTabs:                false,
			indentSize:             4,
			trimMultipleEmptyLines: false,
			oneBracketPerLine:      false,
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
			name:                   "trim multiple empty lines - on",
			useTabs:                true,
			trimMultipleEmptyLines: true,
			indentSize:             1,
			oneBracketPerLine:      false,

			src: []byte(`
variable "DESTDIR" {
  default = "./bin"
  required = true
}


variable "DESTDIR1" {
	default = "./bin"
	required = true
  }`),
			expected: []byte(`
variable "DESTDIR" {
	default  = "./bin"
	required = true
}

variable "DESTDIR1" {
	default  = "./bin"
	required = true
}`),
		},
		{
			name:                   "trim multiple empty lines - off",
			useTabs:                true,
			trimMultipleEmptyLines: false,
			oneBracketPerLine:      false,
			indentSize:             1,

			src: []byte(`
variable "DESTDIR" {
  default = "./bin"
  required = true
}


variable "DESTDIR1" {
	default = "./bin"
	required = true
}`),
			expected: []byte(`
variable "DESTDIR" {
	default  = "./bin"
	required = true
}


variable "DESTDIR1" {
	default  = "./bin"
	required = true
}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &mockery.MockConfiguration_configuration{}
			cfg.EXPECT().UseTabs().Return(tt.useTabs)
			cfg.EXPECT().IndentSize().Return(tt.indentSize)
			cfg.EXPECT().TrimMultipleEmptyLines().Return(tt.trimMultipleEmptyLines)
			cfg.EXPECT().OneBracketPerLine().Return(tt.oneBracketPerLine)

			// Call the Format function with the provided configuration and source
			result, err := hclfmt.FormatBytes(cfg, tt.src)

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
