# syntax=docker/dockerfile:labs

ARG GO_VERSION

ARG BUILDRC_VERSION=
FROM walteh/buildrc:${BUILDRC_VERSION} AS buildrc

# Base tools image with required packages
FROM golang:${GO_VERSION}-bookworm AS tools
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apt-get update && apt-get --no-install-recommends install -y git unzip

# Set common working directory
WORKDIR /wrk

# Buf stage
FROM tools AS bufgen
COPY --from=bufbuild/buf:latest /usr/local/bin/buf /usr/bin/
RUN --mount=type=bind,target=.,rw \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod <<EOT
	set -ex
	buf generate --exclude-path ./vendor --output . --include-imports --include-wkt || echo "buf generate failed - ignoring"
	mkdir /out
	git ls-files -m --others -- ':!vendor' '**/*.pb.go' | tar -cf - --files-from - | tar -C /out -xf -
EOT

# Final update stage
FROM scratch AS update
COPY --from=bufgen /out /

FROM tools AS validate
ARG DESTDIR
RUN --mount=target=/context \
	--mount=from=bufgen,target=/out,source=/out,type=bind <<EOT
	set -e
	buildrc diff --current="/context/${DESTDIR}" --correct="/out" --glob="**/*.pb.go"
EOT
