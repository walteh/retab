package dockerfmt_test

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/walteh/retab/v2/gen/mocks/pkg/formatmock"
	"github.com/walteh/retab/v2/pkg/diff"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/dockerfmt"
)

func formatDocker(ctx context.Context, cfg format.Configuration, src []byte) (string, error) {
	formatter := dockerfmt.NewFormatter()
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

type formatTest struct {
	name     string
	src      string
	expected string
	config   map[string]string
}

//go:embed testdata/simple.dockerfile
var simpleDockerfile []byte

//go:embed testdata/unformatted.dockerfile
var unformattedDockerfile []byte

//go:embed testdata/basic.dockerfile
var basicDockerfile []byte

func TestDockerFormatting(t *testing.T) {
	tests := []formatTest{
		{
			name: "Simple dockerfile",
			src: `FROM ubuntu:20.04
RUN echo hello
`,
			expected: `FROM ubuntu:20.04
RUN echo hello
`,
		},
		{
			name: "With trailing newline",
			src: `FROM ubuntu:20.04
CMD ["echo", "hello"]
`,
			expected: `FROM ubuntu:20.04
CMD ["echo", "hello"]
`,
		},
		{
			name: "Without trailing newline",
			src: `FROM ubuntu:20.04
CMD ["echo", "hello"]`,
			expected: `FROM ubuntu:20.04
CMD ["echo", "hello"]
`,
		},
		{
			name: "Space redirects",
			src: `FROM ubuntu:20.04
RUN echo hello    >     /tmp/file
`,
			config: map[string]string{
				"space_redirects": "true",
			},
			expected: `FROM ubuntu:20.04
RUN echo hello > /tmp/file
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg := formatmock.NewMockConfiguration(t)
			cfg.EXPECT().UseTabs().Return(false).Maybe()
			cfg.EXPECT().IndentSize().Return(4).Maybe()
			cfg.EXPECT().Raw().Return(tt.config).Maybe()

			formatted, err := formatDocker(ctx, cfg, []byte(tt.src))
			require.NoError(t, err, "Format returned error")

			// For other tests, compare the exact output
			diff.Require(t).Want(tt.expected).Got(formatted).Equals()
		})
	}
}

func TestIndentSize(t *testing.T) {
	// Test basic indentation with continuation lines
	src := `FROM ubuntu:20.04
RUN apt-get update && \
    apt-get install -y \
    curl \
    wget
`

	ctx := context.Background()
	cfg := formatmock.NewMockConfiguration(t)
	cfg.EXPECT().UseTabs().Return(false).Maybe()
	cfg.EXPECT().IndentSize().Return(4).Maybe()
	cfg.EXPECT().Raw().Return(map[string]string{}).Maybe()

	formatted, err := formatDocker(ctx, cfg, []byte(src))
	require.NoError(t, err, "Format should not return error")

	// Check that the formatted output has the key elements
	require.Contains(t, formatted, "FROM ubuntu:20.04", "Should contain FROM line")
	require.Contains(t, formatted, "RUN apt-get update", "Should contain apt-get update")
	require.Contains(t, formatted, "curl", "Should contain curl")
	require.Contains(t, formatted, "wget", "Should contain wget")
}
