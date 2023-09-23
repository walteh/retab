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

	git clone https://github.com/golang/tools.git && cd tools
	# put this back once "telemetry" is added to gopls
	# git clone --depth=1 --no-checkout https://github.com/golang/tools.git && cd tools
	# git fetch --tags
	# git checkout gopls/v${GOPLS_VERSION}

	mkdir -p /out

	function copy_pkg() {
		local from=$1
		local name=$2
		(
			mkdir -p /out/$name
			cd ./$from/$name
			find . \( -name '*.go' ! -name '*_test.go'  ! -name 'generate/*' \) -type f | tar -cf - --files-from - | tar -C /out/$name -xf -
		)
	}

	function copy_protocol_file() {
		local src=$1
		(
			cp ./gopls/internal/lsp/protocol/$src /out/$src
			sed -i "s|package protocol|package gopls|g" /out/$src
		)
	}

	function copy_lsrpc_file() {
		local src=$1
		(
			mkdir -p /out/lsprpc
			cp ./gopls/internal/lsp/lsprpc/$src /out/lsprpc/$src
		)
	}

	copy_pkg internal event
	copy_pkg internal jsonrpc2
	copy_pkg internal jsonrpc2_v2
	copy_pkg internal xcontext
	copy_pkg internal tool
	copy_pkg internal fakenet
	copy_pkg gopls/internal/lsp progress
	copy_pkg gopls/internal span
	copy_pkg gopls/internal telemetry
	copy_pkg gopls/internal/lsp safetoken
	copy_pkg gopls/internal bug
	copy_pkg gopls/internal/lsp protocol
	copy_pkg internal diff
	copy_pkg internal robustio
	copy_pkg internal memoize
	copy_pkg internal constraints
	copy_pkg internal persistent
	copy_pkg internal tokeninternal
	copy_pkg internal typeparams



	# copy_protocol_file tsdocument_changes.go
	# copy_protocol_file tsserver.go
	# copy_protocol_file tsjson.go
	# copy_protocol_file protocol.go
	# copy_protocol_file tsprotocol.go
	# copy_protocol_file tsclient.go
	# copy_protocol_file context.go
	# copy_protocol_file log.go

	# find any lines with "vulncheck" or Vulncheck and remove them
	# find /out -type f -name "*.go" -exec sed -i "/vulncheck/d" {} \;
	# find /out -type f -name "*.go" -exec sed -i "/Vulncheck/d" {} \;

	# more specific are first
	# find /out -type f -name "*.go" -exec sed -i "s|\"golang.org/x/tools/gopls/internal/lsp/protocol\"|protocol \"${GO_MODULE}/${DESTDIR#./}\"|g" {} \;
	find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/internal/imports|golang.org/x/tools/imports|g" {} \;

	# important extra slash after command because commandmeta needs it
	# find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/gopls/internal/lsp/command/|${GO_MODULE}/${DESTDIR#./}/|g" {} \;
	find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/gopls/internal/lsp|${GO_MODULE}/${DESTDIR#./}|g" {} \;
	find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/gopls/internal|${GO_MODULE}/${DESTDIR#./}|g" {} \;
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
