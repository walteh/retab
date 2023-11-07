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
			# find . \( -name '*.go' ! -name '*_test.go'  ! -path '*/generate/*' ! -path '*/testdata/*'  \) -type f | tar -cf - --files-from - | tar -C /out/$name -xf -
			find . \( -name '*.go' ! -path '*/generate/*'  \) -type f | tar -cf - --files-from - | tar -C /out/$name -xf -
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
	copy_pkg internal diff
	copy_pkg internal gcimporter
	copy_pkg internal pkgbits
	copy_pkg internal imports
	copy_pkg internal testenv
	copy_pkg internal gopathwalk
	copy_pkg internal gocommand
	copy_pkg internal fastwalk
	copy_pkg internal facts
	copy_pkg internal stack
	copy_pkg internal typesinternal
	copy_pkg internal packagesinternal
	copy_pkg internal refactor
	copy_pkg internal robustio
	copy_pkg internal memoize
	copy_pkg internal constraints
	copy_pkg internal persistent
	copy_pkg internal tokeninternal
	copy_pkg internal typeparams
	copy_pkg internal packagesinternal
	copy_pkg internal goroot
	copy_pkg internal analysisinternal
	copy_pkg internal fuzzy
	copy_pkg internal apidiff
	copy_pkg internal proxydir
	copy_pkg gopls/internal/lsp fake

	copy_pkg gopls/internal span
	copy_pkg gopls/internal telemetry
	copy_pkg gopls/internal bug
	copy_pkg gopls/internal vulncheck
	copy_pkg gopls/internal astutil
	copy_pkg gopls/internal hooks

	copy_pkg gopls/internal/lsp snippet
	copy_pkg gopls/internal/lsp frob
	copy_pkg gopls/internal/lsp lsprpc
	copy_pkg gopls/internal/lsp debug
	copy_pkg gopls/internal/lsp tests
	copy_pkg gopls/internal/lsp analysis
	copy_pkg gopls/internal/lsp cache
	copy_pkg gopls/internal/lsp progress
	copy_pkg gopls/internal/lsp filecache
	copy_pkg gopls/internal/lsp lru
	copy_pkg gopls/internal/lsp safetoken
	copy_pkg gopls/internal/lsp glob
	copy_pkg gopls/internal/lsp protocol
	copy_pkg gopls/internal/lsp regtest
	copy_pkg gopls/internal/lsp template
	copy_pkg gopls/internal/lsp mod

	rm -rf /out/mod/code_lens.go





	# find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/internal/imports|golang.org/x/tools/imports|g" {} \;
	find /out -type f -name "*.go" -exec sed -i "s|\"golang.org/x/tools/gopls/internal/lsp\"|lsp \"github.com/walteh/retab/internal/server\"|g" {} \;
	find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/gopls/internal/lsp/source|github.com/walteh/retab/internal/source|g" {} \;
	find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/gopls/internal/lsp/command|github.com/walteh/retab/internal/command|g" {} \;
	find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools/gopls/internal/lsp/cmd|github.com/walteh/retab/internal/cmd|g" {} \;

	find /out -type f -name '*.go' -exec sed -i 's|command.New[^\(]*Command|command.NewNoopCommand|g' {} \;


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
