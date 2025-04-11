package hclfmt_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/walteh/retab/v2/gen/mocks/pkg/formatmock"
	"github.com/walteh/retab/v2/pkg/diff"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
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
  ok = [{abc = 1}
]
}`),
			expected: []byte(`
variable "DESTDIR" {
	default  = "./bin"
	required = true
	ok = [
		{
			abc = 1
		},
	]
}
`),
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
}
`),
		},
		{
			name:       "trim multiple empty lines - on",
			useTabs:    true,
			indentSize: 1,

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
}
`),
		},
		{
			name:       "trim multiple empty lines - off",
			useTabs:    true,
			indentSize: 1,

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
}
`),
		},
		{name: "add newline at end of file even if no bracket",
			useTabs:    true,
			indentSize: 1,

			src: []byte(`default = "./bin"`),
			expected: []byte(`default = "./bin"
`),
		},
		{name: "trailing comma",
			useTabs:    true,
			indentSize: 1,

			src: []byte(`variable "example_list" {
  type = list(string)
  default = [
    "value1",
    "value2",
    "value3"
  ]
}`),
			expected: []byte(`variable "example_list" {
	type = list(string)
	default = [
		"value1",
		"value2",
		"value3",
	]
}
`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := formatmock.NewMockConfiguration(t)
			cfg.EXPECT().UseTabs().Return(tt.useTabs).Maybe()
			cfg.EXPECT().IndentSize().Return(tt.indentSize).Maybe()

			// Call the Format function with the provided configuration and source
			result, err := hclfmt.FormatBytes(cfg, tt.src)
			require.NoError(t, err)

			diff.Require(t).Want(tt.expected).Got(result).Equals()
		})
	}
}
