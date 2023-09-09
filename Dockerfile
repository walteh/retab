# syntax=docker/dockerfile:labs

##################################################################
# SETUP
##################################################################

ARG GO_VERSION=1.21.0
ARG XX_VERSION=1.2.1

ARG DOCKER_VERSION=24.0.5
ARG GOTESTSUM_VERSION=v1.10.1
ARG REGISTRY_VERSION=latest
ARG BUILDKIT_VERSION=nightly

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS golatest

FROM --platform=$BUILDPLATFORM walteh/buildrc:0.14.1 as buildrc

FROM --platform=$BUILDPLATFORM alpine:latest AS alpine
FROM --platform=$BUILDPLATFORM busybox:musl AS musl

FROM golatest AS gobase
COPY --from=xx / /
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apk add --no-cache file git bash
ENV GOFLAGS=-mod=vendor
ENV CGO_ENABLED=0
WORKDIR /src

##################################################################
# BUILD
##################################################################

FROM gobase AS metarc
ARG TARGETPLATFORM
RUN --mount=type=bind,target=/src,readonly \
	buildrc full --git-dir=/src --files-dir=/meta

FROM scratch AS meta
COPY --from=metarc /meta /meta

FROM gobase AS builder
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod \
	--mount=type=bind,from=meta,source=/meta,target=/meta,readonly <<EOT
  	set -e
 	xx-go --wrap;
	LDFLAGS="-s -w -X $(cat /meta/go-pkg)/version.Version=$(cat /meta/version) -X $(cat /meta/go-pkg)/version.Revision=$(cat /meta/revision) -X $(cat /meta/go-pkg)/version.Package=$(cat /meta/go-pkg)";
	CGO_ENABLED=0 go build -mod vendor -trimpath -ldflags "$LDFLAGS" -o /out/$(cat /meta/executable) ./cmd;
  	xx-verify --static /out/$(cat /meta/executable);
EOT

FROM scratch AS build-unix
ARG BIN_NAME
COPY --link --from=builder /out/${BIN_NAME} /${BIN_NAME}

FROM build-unix AS build-darwin
FROM build-unix AS build-linux

FROM scratch AS build-windows
ARG BIN_NAME
COPY --link --from=builder /out/${BIN_NAME} /${BIN_NAME}.exe

FROM build-$TARGETOS AS build
# enable scanning for this stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
COPY --link --from=meta /meta/buildrc.json /


##################################################################
# TESTING
##################################################################

FROM gobase AS gotestsum
ARG GOTESTSUM_VERSION
ENV GOFLAGS=
RUN --mount=target=/root/.cache,type=cache <<EOT
	GOBIN=/out/ go install "gotest.tools/gotestsum@${GOTESTSUM_VERSION}" &&
	/out/gotestsum --version
EOT

FROM gobase AS test2json
ARG GOTESTSUM_VERSION
ENV GOFLAGS=
RUN --mount=target=/root/.cache,type=cache <<EOT
	CGO_ENABLED=0 go build -o /out/test2json -ldflags="-s -w" cmd/test2json
EOT

FROM gobase AS test-builder
ARG BIN_NAME
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev libc6-compat
RUN mkdir -p /out
RUN --mount=type=bind,target=. \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod \
	go test -coverprofile=/reports/coverage-report.txt -c -race -vet='' -covermode=atomic -mod=vendor ./... -o /out

FROM scratch AS test
COPY --link --from=test-builder /out /
COPY --link --from=gotestsum /out /
COPY --link --from=test2json /out /

FROM alpine AS case-builder
ARG PKG= NAME= ARGS= E2E=
RUN <<EOT
	set -e -x -o pipefail
	mkdir -p /dat

	echo "${ARGS}" > /dat/args
	echo "${E2E}" > /dat/e2e
	echo "${PKG##*/}" > /dat/pkg
	echo "${NAME}" > /dat/name
EOT

FROM scratch AS case
COPY --link --from=case-builder /dat /dat
COPY --link --from=test . /
COPY --link --from=build . /

FROM alpine AS test-runner
ARG GO_VERSION
ENV GOVERSION=${GO_VERSION}
ARG DOCKER_HOST=tcp://0.0.0.0:2375
COPY --link --from=case . .
RUN apk add --no-cache file
RUN --network=host /usr/bin/gotestsum --format=standard-verbose \
	--jsonfile=/out/go-test-report-$(cat /dat/pkg)-$(cat /dat/name).json \
	--junitfile=/out/junit-report-$(cat /dat/pkg)-$(cat /dat/name).xml \
	--raw-command -- /usr/bin/test2json -t -p $(cat /dat/pkg) /usr/bin/$(cat /dat/pkg).test  -test.bench=.  -test.timeout=10m \
	-test.v -test.coverprofile=/out/coverage-report-$(cat /dat/pkg)-$(cat /dat/name).txt $(cat /dat/args) \
	-test.outputdir=/out;

FROM scratch AS tester
COPY --link --from=test-runner /out /


##################################################################
# RELEASE
##################################################################

FROM alpine AS packager
ARG BUILDKIT_MULTI_PLATFORM
RUN apk add --no-cache file tar jq
COPY --link  --from=build . /src/
RUN <<EOT
	set -e
	if [ "$BUILDKIT_MULTI_PLATFORM" != 'true' && "$BUILDKIT_MULTIPLATFORM" != '1' ]; then
		searchdir="/src/"
	else
		searchdir="/src/*/"
	fi
	mkdir -p /out
	for pdir in ${searchdir}; do
		(
			cd "${pdir}"
			artifact="$(jq -r '.artifact' buildrc.json)"
			tar -czf "/out/${artifact}.tar.gz" .
		)
	done
	(
		cd /out
		find . -type f \( -name '*.tar.gz' \) -exec sha256sum -b {} \; >./checksums.txt
		sha256sum -c checksums.txt
	)
EOT

FROM scratch AS package
COPY --from=packager /out/ /

##################################################################
# IMAGE
##################################################################

FROM scratch AS entry
ARG BIN_NAME
ENV BIN_NAME=${BIN_NAME}
COPY --link --from=meta /meta/buildrc.json /usr/bin/${BIN_NAME}/buildrc.json
COPY --link --from=builder /usr/bin/${BIN_NAME} /usr/bin/
ENTRYPOINT /usr/bin/${BIN_NAME}
