# syntax=docker/dockerfile:labs

ARG GO_VERSION=
ARG MODOUTDATED_VERSION=v0.8.0

ARG BUILDRC_VERSION=
FROM walteh/buildrc:${BUILDRC_VERSION} AS buildrc


FROM golang:${GO_VERSION}-alpine AS tools
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apk add --no-cache git rsync
WORKDIR /src

FROM psampaz/go-mod-outdated:${MODOUTDATED_VERSION} AS go-mod-outdated

FROM tools AS vendored
RUN --mount=target=/context \
	--mount=target=.,type=tmpfs \
	--mount=target=/go/pkg/mod,type=cache <<EOT
	set -e
	rsync -a /context/. .
	go mod tidy
	go mod vendor
	mkdir /out
	cp -r go.mod go.sum vendor /out
EOT

FROM scratch AS update
COPY --from=vendored /out /

FROM tools AS validate
ARG DESTDIR
RUN --mount=target=/context \
	--mount=from=vendored,target=/out,source=/out,type=bind <<EOT
	set -e
	buildrc diff --current="/context/${DESTDIR}" --correct="/out" --glob="**/vendor/**" --glob="**/go.sum" --glob="**/go.mod"
EOT

FROM vendored AS outdated
COPY --from=go-mod-outdated /home/go-mod-outdated /usr/bin/go-mod-outdated
RUN --mount=target=/context \
	--mount=target=.,type=tmpfs \
	--mount=target=/go/pkg/mod,type=cache <<EOT
	set -e
	cd /out
	go list -mod=readonly -u -m -json all | go-mod-outdated -update -direct >/outdated.txt
EOT

FROM scratch AS outdated-output
COPY --from=outdated /outdated.txt /outdated.txt
