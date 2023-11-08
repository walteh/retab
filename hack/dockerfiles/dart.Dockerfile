# syntax=docker/dockerfile:labs

ARG GO_VERSION
ARG BUILDRC_VERSION
ARG DART_VERSION

FROM walteh/buildrc:${BUILDRC_VERSION} AS buildrc
FROM dart:${DART_VERSION} as dart

# Base tools image with required packages
FROM golang:${GO_VERSION}-alpine AS tools
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc

# Set common working directory
WORKDIR /wrk



# Final update stage
FROM scratch AS generate
COPY --from=dart /usr/lib/dart/bin/dart /dist/
COPY <<EOT /dart.go
	package dart

	import "embed"

	//go:embed dist/*
	var StaticAssets embed.FS
	var StaticAssetsDir = "dist"
EOT


FROM tools AS validate
ARG DESTDIR
COPY --from=generate . /expected
RUN --mount=target=/current <<EOT
	set -e
	cd /current
	buildrc diff --current="./${DESTDIR}" --correct="/expected" --glob="*"
EOT
