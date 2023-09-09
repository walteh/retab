# syntax=docker/dockerfile:labs

ARG GO_VERSION=
ARG BUILDRC_VERSION=

FROM walteh/buildrc:${BUILDRC_VERSION} AS buildrc

FROM alpine AS tools
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apk add --no-cache git rsync


FROM golang:${GO_VERSION}-alpine AS docsgen
WORKDIR /src
RUN --mount=target=. \
	--mount=target=/root/.cache,type=cache \
	go build -mod=vendor -o /out/docsgen ./docs/generate.go

FROM tools AS gen
RUN apk add --no-cache rsync git
WORKDIR /src
COPY --from=docsgen /out/docsgen /usr/bin
ARG DOCS_FORMATS
ARG BUILDX_EXPERIMENTAL
RUN --mount=target=/context \
	--mount=target=.,type=tmpfs <<EOT
	set -e
	rsync -a /context/. .
	FORMATS="${DOCS_FORMATS}" docsgen "docs/reference"
	mkdir /out
	cp -r docs/reference/* /out
EOT

FROM scratch AS generate
COPY --from=gen /out /

FROM tools AS validate
ARG DESTDIR
RUN --mount=target=/context \
	--mount=from=gen,target=/out,source=/out,type=bind <<EOT
	set -e
	ls -l /out
	buildrc diff --current="/context/${DESTDIR}" --correct="/out" --glob="**/*"
EOT
