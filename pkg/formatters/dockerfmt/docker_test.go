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
	"github.com/walteh/retab/v2/pkg/formatters/dockerfmt"
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
		{
			name: "Complex multiline RUN with backslashes",
			src: `FROM alpine:3.14
RUN apk update && \
 apk add --no-cache \
  python3 \
  py3-pip \
  curl \
  jq && \
 pip3 install --upgrade pip
`,
			expected: `FROM alpine:3.14
RUN apk update \
	&& apk add --no-cache \
		python3 \
		py3-pip \
		curl \
		jq \
	&& pip3 install --upgrade pip
`,
		},
		{
			name: "Multi-stage build",
			src: `FROM golang:1.18 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/myapp

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/myapp /app/
EXPOSE 8080
CMD ["/app/myapp"]
`,
			expected: `FROM golang:1.18 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/myapp

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/myapp /app/
EXPOSE 8080
CMD ["/app/myapp"]
`,
		},
		{
			name: "ENV and ARG declarations",
			src: `FROM node:16
ARG NODE_ENV=production
ENV PORT=3000
ENV HOST=0.0.0.0
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE $PORT
CMD ["npm", "start"]
`,
			expected: `FROM node:16
ARG NODE_ENV=production
ENV PORT=3000
ENV HOST=0.0.0.0
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE $PORT
CMD ["npm", "start"]
`,
		},
		{
			name: "LABEL and comments",
			src: `# Base image
FROM ubuntu:22.04

# Set labels for metadata
LABEL maintainer="example@example.com"
LABEL version="1.0"
LABEL description="Example Docker image"

# Install dependencies
RUN apt-get update && apt-get install -y \
  python3 \
	python3-pip

# Setup app directory
WORKDIR /app

# Run the application
CMD ["python3", "app.py"]
`,
			expected: `# Base image
FROM ubuntu:22.04

# Set labels for metadata
LABEL maintainer="example@example.com"
LABEL version="1.0"
LABEL description="Example Docker image"

# Install dependencies
RUN apt-get update && apt-get install -y \
	python3 \
	python3-pip

# Setup app directory
WORKDIR /app

# Run the application
CMD ["python3", "app.py"]
`,
		},
		{
			name: "HEALTHCHECK and ENTRYPOINT",
			src: `FROM alpine:latest
RUN apk add --no-cache curl
HEALTHCHECK --interval=30s --timeout=3s \
    CMD curl -f http://localhost/ || exit 1
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
`,
			expected: `FROM alpine:latest
RUN apk add --no-cache curl
HEALTHCHECK --interval=30s --timeout=3s \
	CMD curl -f http://localhost/ || exit 1
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon", "off;"]
`,
		},
		{
			name: "Complex JSON in CMD and ENTRYPOINT",
			src: `FROM alpine:latest
ENTRYPOINT ["sh", "-c", "echo 'Starting server with config:' && cat /config.json"]
CMD ["{ \"port\": 8080, \"debug\": true, \"database\": { \"host\": \"db\", \"port\": 5432 } }"]
`,
			expected: `FROM alpine:latest
ENTRYPOINT ["sh", "-c", "echo", "Starting server with config:", "&&", "cat", "/config.json"]
CMD ["{", "port:", "8080,", "debug:", "true,", "database:", "{", "host:", "db,", "port:", "5432", "}", "}"]
`,
		},
		{
			name: "Empty lines and comments",
			src: `FROM alpine:latest

# Install dependencies

RUN apk add --no-cache curl

# Configure application
COPY . /app

# Set the entrypoint
CMD ["./app"]
`,
			expected: `FROM alpine:latest

# Install dependencies

RUN apk add --no-cache curl

# Configure application
COPY . /app

# Set the entrypoint
CMD ["./app"]
`,
		},
		{
			name: "Shell form commands",
			src: `FROM ubuntu:20.04
RUN apt-get update && apt-get install -y nginx
CMD nginx -g 'daemon off;'
`,
			expected: `FROM ubuntu:20.04
RUN apt-get update && apt-get install -y nginx
CMD ["nginx", "-g", "daemon off;"]
`,
		},
		{
			name: "Heredoc syntax",
			src: `FROM alpine:latest
RUN <<EOT
  echo "This is multiline shell script"
  echo "Second line"
  echo "Third line"
EOT
`,
			expected: `FROM alpine:latest
RUN <<EOT
echo "This is multiline shell script"
echo "Second line"
echo "Third line"
EOT
`,
		},
		{
			name: "COPY with heredoc syntax",
			src: `FROM alpine:latest
COPY <<EOF /destination/
  content line 1
  content line 2
  content line 3
EOF
`,
			expected: `FROM alpine:latest
COPY <<EOF /destination/
content line 1
content line 2
content line 3
EOF
`,
		},
		{
			name: "RUN with heredoc syntax and arguments",
			src: `FROM alpine:latest
RUN --mount=type=secret,id=mysecret <<EOT
  echo "Reading secret"
  cat /run/secrets/mysecret
EOT
`,
			expected: `FROM alpine:latest
RUN --mount=type=secret,id=mysecret <<EOT
echo "Reading secret"
cat /run/secrets/mysecret
EOT
`,
		},
		{
			name: "OnBuild instructions",
			src: `FROM alpine:latest
ONBUILD RUN echo "This runs when the image is used as a base"
ONBUILD COPY . /app
`,
			expected: `FROM alpine:latest
ONBUILD RUN echo "This runs when the image is used as a base"
ONBUILD COPY . /app
`,
		},
		{
			name: "ADD with URL",
			src: `FROM alpine:latest
ADD https://example.com/file.tar.gz /tmp/
RUN tar -xzf /tmp/file.tar.gz -C /usr/local/
`,
			expected: `FROM alpine:latest
ADD https://example.com/file.tar.gz /tmp/
RUN tar -xzf /tmp/file.tar.gz -C /usr/local/
`,
		},
		{
			name: "USER and WORKDIR with variables",
			src: `FROM alpine:latest
ARG user=appuser
ENV APP_HOME=/home/$user
USER $user
WORKDIR $APP_HOME
`,
			expected: `FROM alpine:latest
ARG user=appuser
ENV APP_HOME=/home/$user
USER $user
WORKDIR $APP_HOME
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg := formatmock.NewMockConfiguration(t)
			cfg.EXPECT().UseTabs().Return(true).Maybe()
			cfg.EXPECT().IndentSize().Return(4).Maybe()
			cfg.EXPECT().Raw().Return(tt.config).Maybe()

			formatted, err := formatDocker(ctx, cfg, []byte(tt.src))
			require.NoError(t, err, "Format returned error")

			// Compare the exact output
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

func TestErrorHandling(t *testing.T) {
	// Test with non-empty but minimal valid input
	ctx := context.Background()
	cfg := formatmock.NewMockConfiguration(t)
	cfg.EXPECT().UseTabs().Return(false).Maybe()
	cfg.EXPECT().IndentSize().Return(4).Maybe()
	cfg.EXPECT().Raw().Return(map[string]string{}).Maybe()

	formatter := dockerfmt.NewFormatter()
	// Use a minimal valid Dockerfile
	_, err := formatter.Format(ctx, cfg, bytes.NewReader([]byte("FROM scratch\n")))
	require.NoError(t, err, "Minimal valid Dockerfile should not cause error")

	// Test with corrupted reader
	corruptedReader := &errorReader{err: context.Canceled}
	_, err = formatter.Format(ctx, cfg, corruptedReader)
	require.Error(t, err, "Should return error for corrupted reader")
}

// errorReader is a mock reader that always returns an error
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func TestComplexDockerfile(t *testing.T) {
	// A complex Dockerfile with various features
	src := `# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine AS builder

# Set build arguments and environment variables
ARG VERSION=dev
ENV CGO_ENABLED=0
ENV GOOS=linux

# Set working directory and install dependencies
WORKDIR /app
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    apk add --no-cache git ca-certificates && \
    update-ca-certificates

# Download dependencies first for better caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy the rest of the code and build
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-X main.version=${VERSION}" -o /app/server ./cmd/server

# Create a minimal runtime image
FROM alpine:3.16
RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/server /usr/local/bin/

# Set metadata and run configuration
LABEL org.opencontainers.image.source="https://github.com/example/app"
LABEL org.opencontainers.image.description="Example application"
LABEL org.opencontainers.image.version="${VERSION}"

# Set up runtime environment
EXPOSE 8080
VOLUME /data
WORKDIR /data

# Health check to verify the application is running properly
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Set the entrypoint and default command
ENTRYPOINT ["/usr/local/bin/server"]
CMD ["--config", "/data/config.yaml"]
`

	ctx := context.Background()
	cfg := formatmock.NewMockConfiguration(t)
	cfg.EXPECT().UseTabs().Return(false).Maybe()
	cfg.EXPECT().IndentSize().Return(4).Maybe()
	cfg.EXPECT().Raw().Return(map[string]string{}).Maybe()

	formatted, err := formatDocker(ctx, cfg, []byte(src))
	require.NoError(t, err, "Format should not return error")

	// Check for various elements in the Dockerfile
	require.Contains(t, formatted, "# syntax=docker/dockerfile:1.4", "Should contain syntax header")
	require.Contains(t, formatted, "FROM golang:1.18-alpine AS builder", "Should contain builder stage")
	require.Contains(t, formatted, "FROM alpine:3.16", "Should contain runtime stage")
	require.Contains(t, formatted, "HEALTHCHECK", "Should contain healthcheck")
	require.Contains(t, formatted, "ENTRYPOINT", "Should contain entrypoint")
	require.Contains(t, formatted, "CMD", "Should contain cmd")
}
