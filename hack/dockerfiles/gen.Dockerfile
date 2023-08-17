# syntax=docker/dockerfile:1

# Forked from https://github.com/moby/buildkit/blob/e1b3b6c4abf7684f13e6391e5f7bc9210752687a/hack/dockerfiles/generated-files.Dockerfile
# Copyright The BuildKit Authors.
# Copyright The Buildx Authors.
# Licensed under the Apache License, Version 2.0

ARG GO_VERSION

# Base tools image with required packages
FROM golang:${GO_VERSION}-bookworm AS tools
RUN apt-get update && apt-get --no-install-recommends install -y git unzip

# Mounting common volumes
RUN --mount=type=bind,target=.,rw \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod

# Set common working directory
WORKDIR /wrk

# Mockery stage
FROM tools as mockerygen
COPY --from=vektra/mockery:latest /usr/local/bin/mockery /usr/bin/
RUN --mount=type=bind,target=.,rw <<EOT
	set -ex
	mockery --dir .
	mkdir /out
	git ls-files -m --others -- ':!vendor' '*.mockery.go' | tar -cf - --files-from - | tar -C /out -xf -
EOT

# Buf stage
FROM tools AS bufgen
COPY --from=bufbuild/buf:latest /usr/local/bin/buf /usr/bin/
RUN --mount=type=bind,target=.,rw <<EOT
  set -ex
  buf generate --exclude-path ./vendor --output . --include-imports --include-wkt || echo "buf generate failed - ignoring"
  mkdir /out
  git ls-files -m --others -- ':!vendor' '**/*.pb.go' | tar -cf - --files-from - | tar -C /out -xf -
EOT

# Final update stage
FROM scratch AS update
COPY --from=bufgen /out /buf/
COPY --from=mockerygen /out /mockery/

FROM tools AS validate
ENV GIT_DISCOVERY_ACROSS_FILESYSTEM=true
RUN --mount=type=bind,target=.,rw \
	--mount=type=bind,from=update,source=/buf,target=./buf \
	--mount=type=bind,from=update,source=/mockery,target=./mockery <<EOT
  set -e
  ls -la

	pwd

	(
		cd ./buf
		git add -A

		diff=$(git status --porcelain -- ':!vendor' '**/*.pb.go')
		if [ -n "$diff" ]; then
			echo >&2 'ERROR: The result of "buf generate" differs. Please update with "make gen"'
			echo "$diff"
			exit 1
		fi
	)

	(
		cd ./mockery
		git add -A

		diff=$(git status --porcelain -- ':!vendor' '**/*.mockery.go')
		if [ -n "$diff" ]; then
			echo >&2 'ERROR: The result of "mockery" differs. Please update with "make gen"'
			echo "$diff"
			exit 1
		fi
	)
EOT
