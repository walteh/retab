# syntax=docker/dockerfile:labs

ARG GO_VERSION
ARG MOCKERY_VERSION
ARG BUILDRC_VERSION

FROM vektra/mockery:v${MOCKERY_VERSION} AS mockery
FROM walteh/buildrc:${BUILDRC_VERSION} AS buildrc

FROM golang:${GO_VERSION}-alpine AS tools
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apk add --no-cache git curl

# Set common working directory
WORKDIR /wrk

# Mockery stage
FROM tools as generator
ARG GOPLS_VERSION
RUN --mount=type=bind,target=.,rw \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod <<SHELL
	set -ex
	function dl() {
		local ver="$1"
		local file="$2"
		local dest="https://raw.githubusercontent.com/golang/tools/gopls/v$ver/gopls/internal/lsp/protocol/$file"
		echo "Downloading $dest"
		curl -sSL -o "/out/$file" "$dest"
	}
	mkdir -p /out
	dl ${GOPLS_VERSION} tsprotocol.go
	dl ${GOPLS_VERSION} tsdocument_changes.go
SHELL

# Final update stage
FROM scratch AS generate
COPY --from=generator /out /

FROM tools AS validate
ARG DESTDIR
RUN --mount=target=/context \
	--mount=from=generator,target=/out,source=/out,type=bind <<SHELL
	set -e
	cd /context
	buildrc diff --current="./${DESTDIR}" --correct="/out" --glob="*.go"
SHELL
