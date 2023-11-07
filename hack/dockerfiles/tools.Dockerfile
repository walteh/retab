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

	find . -name "*.go" -type f | tar -cf - --files-from - | tar -C /out -xf -

	# replace all "golang.org/x/tools" imports with "${GO_MODULE}/${DESTDIR}" # imported from "golang.org/x/tools"
	find /out -type f -name "*.go" -exec sed -i "s|golang.org/x/tools|${GO_MODULE}/${DESTDIR#./}|g" {} \;
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
	buildrc diff --current="./${DESTDIR}" --correct="/out" --glob="./**/*.go"
SHELL
