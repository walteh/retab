# syntax=docker/dockerfile:labs

ARG GO_VERSION

ARG BUILDRC_VERSION=
FROM walteh/buildrc:${BUILDRC_VERSION} AS buildrc

FROM golang:${GO_VERSION}-alpine AS tools
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apk add --no-cache git

# Set common working directory
WORKDIR /wrk

# Mockery stage
FROM tools as mockerygen
COPY --from=vektra/mockery:latest /usr/local/bin/mockery /usr/bin/
RUN --mount=type=bind,target=.,rw \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod <<EOT
	set -ex
	mockery --dir ./tmp
	mkdir /out
	cd ./tmp
	git ls-files -m --others -- ':!vendor' '*.mockery.go' | tar -cf - --files-from - | tar -C /out -xf -
EOT

# Final update stage
FROM scratch AS generate
COPY --from=mockerygen /out /

FROM tools AS validate
ARG DESTDIR
RUN --mount=target=/context \
	--mount=from=mockerygen,target=/out,source=/out,type=bind <<EOT
	set -e
	buildrc diff --current="/context/${DESTDIR}" --correct="/out" --glob="*.mockery.go"
EOT
