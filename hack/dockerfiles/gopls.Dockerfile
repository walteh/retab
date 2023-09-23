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
ARG GOPLS_VERSION GO_MODULE DESTDIR
RUN --mount=type=bind,target=/wrk/repo,rw \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod <<SHELL
	set -ex

	git clone --depth=1 --no-checkout https://github.com/golang/tools.git
	cd tools
	git fetch --tags
	git checkout gopls/v${GOPLS_VERSION}
	mkdir -p /out

	function copy_internal_pkg() {
		local src=$1
		(
			mkdir -p /out/$src
			cd ./internal/$src
			find . \( -name '*.go' ! -name '*_test.go' \) -type f | tar -cf - --files-from - | tar -C /out/$src -xf -
		)
	}

	function copy_protocol_file() {
		local src=$1
		(
			cp ./gopls/internal/lsp/protocol/$src /out/$src
			sed -i "s|package protocol|package gopls|g" /out/$src
		)
	}

	copy_internal_pkg event
	copy_internal_pkg jsonrpc2
	copy_internal_pkg jsonrpc2_v2
	copy_internal_pkg xcontext
	copy_internal_pkg tool
	copy_internal_pkg fakenet

	copy_protocol_file tsdocument_changes.go
	copy_protocol_file tsserver.go
	copy_protocol_file tsjson.go
	copy_protocol_file protocol.go
	copy_protocol_file tsprotocol.go
	copy_protocol_file tsclient.go

	find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/internal|${GO_MODULE}/${DESTDIR#./}|g" {} \;
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
