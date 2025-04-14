package shfmt

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/walteh/retab/v2/pkg/diff"
	"github.com/walteh/retab/v2/pkg/format"
)

type configWithOptions struct {
	format.Configuration
	options map[string]string
}

func (c *configWithOptions) Raw() map[string]string {
	baseRaw := c.Configuration.Raw()
	for k, v := range c.options {
		baseRaw[k] = v
	}
	return baseRaw
}

// Helper for creating a configuration with specific options
func createTestConfig(useTabs bool, indent int, options map[string]string) format.Configuration {
	cfg := format.NewBasicConfigurationProvider(useTabs, indent)
	return &configWithOptions{
		Configuration: cfg,
		options:       options,
	}
}

func TestShellFormatter(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
		useTabs  bool
		indent   int
	}{
		{
			name:     "basic_echo_command",
			source:   "echo 'Hello, world!'",
			expected: "echo 'Hello, world!'\n",
			useTabs:  true,
			indent:   4,
		},
		{
			name: "if_statement_indentation",
			source: `if [ $a -eq 1 ]; then
echo "a equals 1"
else
echo "a is not 1"
fi`,
			expected: `if [ $a -eq 1 ]; then
	echo "a equals 1"
else
	echo "a is not 1"
fi
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "if_statement_with_spaces",
			source: `if [ $a -eq 1 ]; then
echo "a equals 1"
else
echo "a is not 1"
fi`,
			expected: `if [ $a -eq 1 ]; then
  echo "a equals 1"
else
  echo "a is not 1"
fi
`,
			useTabs: false,
			indent:  2,
		},
		{
			name: "for_loop_formatting",
			source: `for i in 1 2 3; do
echo $i
done`,
			expected: `for i in 1 2 3; do
	echo $i
done
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "multiline_function_definition",
			source: `function hello() {
echo "Hello"
echo "World"
}`,
			expected: `function hello() {
	echo "Hello"
	echo "World"
}
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "comment_preservation",
			source: `# This is a comment
echo 'test' # inline comment`,
			expected: `# This is a comment
echo 'test' # inline comment
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "redirects_with_spaces",
			source: `cat file.txt>  output.txt
echo "test">file.txt`,
			expected: `cat file.txt > output.txt
echo "test" > file.txt
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "case_statement_indentation",
			source: `case "$var" in
"option1")
echo "Selected option 1"
;;
"option2")
echo "Selected option 2"
;;
esac`,
			expected: `case "$var" in
	"option1")
		echo "Selected option 1"
		;;
	"option2")
		echo "Selected option 2"
		;;
esac
`,
			useTabs: true,
			indent:  4,
		},
		{
			name:   "binary_operators_with_newlines",
			source: `[ -f /etc/passwd ] && echo "exists" || echo "not found"`,
			expected: `[ -f /etc/passwd ] && echo "exists" || echo "not found"
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "nested_structures",
			source: `if [ $a -eq 1 ]; then
for i in 1 2 3; do
echo $i
done
fi`,
			expected: `if [ $a -eq 1 ]; then
	for i in 1 2 3; do
		echo $i
	done
fi
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "nested_structures_with_comments",
			source: `if [ $a -eq 1 ]; then # comment
for i in 1 2 3; do # comment
echo $i # comment
done # comment
fi # comment
`,
			expected: `if [ $a -eq 1 ]; then  # comment
	for i in 1 2 3; do # comment
		echo $i        # comment
	done               # comment
fi                     # comment
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "heredoc_formatting",
			source: `cat <<   EOF
This is a heredoc
with multiple lines
EOF`,
			expected: `cat << EOF
This is a heredoc
with multiple lines
EOF
`,
			useTabs: true,
			indent:  4,
		},
		{
			name: "complex_script_with_multiple_elements",
			source: `#!/bin/zsh
# Script header

VAR="value"

function test_func() {
  local var="local"
  echo $var
}

if [ -z "$VAR" ]; then
echo "VAR is empty"
else
echo "VAR=$VAR"
fi

for i in {1..3}; do
  echo $i
done`,
			expected: `#!/bin/zsh
# Script header

VAR="value"

function test_func() {
	local var="local"
	echo $var
}

if [ -z "$VAR" ]; then
	echo "VAR is empty"
else
	echo "VAR=$VAR"
fi

for i in {1..3}; do
	echo $i
done
`,
			useTabs: true,
			indent:  4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create formatter
			formatter := NewFormatter()

			// Create basic configuration
			cfg := format.NewBasicConfigurationProvider(tc.useTabs, tc.indent)

			// Format the file
			ctx := context.Background()
			result, err := formatter.Format(ctx, cfg, strings.NewReader(tc.source))
			if err != nil {
				t.Fatalf("Failed to format file: %v", err)
			}

			bytes, err := io.ReadAll(result)
			if err != nil {
				t.Fatalf("Failed to read formatted result: %v", err)
			}

			diff.Require(t).Want(tc.expected).Got(string(bytes)).Equals()
		})
	}
}
