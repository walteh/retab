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

# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS golatest

FROM --platform=$BUILDPLATFORM walteh/buildrc:0.12.7 as buildrc

FROM golatest AS gobase
COPY --from=xx / /
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apk add --no-cache file git bash
ENV GOFLAGS=-mod=vendor
ENV CGO_ENABLED=0
WORKDIR /src

FROM --platform=$BUILDPLATFORM registry:$REGISTRY_VERSION AS registry

FROM --platform=$BUILDPLATFORM moby/buildkit:$BUILDKIT_VERSION AS buildkit

FROM --platform=$BUILDPLATFORM docker/buildx-bin:latest AS buildx-bin

FROM --platform=$BUILDPLATFORM docker/compose-bin:latest AS compose-bin

##################################################################
# BINARIES
##################################################################

FROM gobase AS meta
ARG TARGETPLATFORM
RUN --mount=type=bind,target=/src,rw <<EOT
	buildrc full --git-dir=/src --files-dir=/meta
EOT

FROM gobase AS binary-cache
ARG DESTDIR
RUN --mount=type=bind,target=/src \
	--mount=type=bind,from=meta,source=/meta,target=/meta,readonly <<EOT
	mkdir -p /binary-cache
	echo "checking for binary cache in /src/rebin/$(cat /meta/artifact).tar.gz"
	if [ -f "/src/rebin/$(cat /meta/artifact).tar.gz" ]; then
		echo "found binary cache in /src/rebin/$(cat /meta/artifact).tar.gz";
		tar xzf "/src/rebin/$(cat /meta/artifact).tar.gz" -C /binary-cache;
	fi
EOT

FROM gobase AS builder
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod \
	--mount=type=bind,from=binary-cache,source=/binary-cache,target=/binary-cache,readonly \
	--mount=type=bind,from=meta,source=/meta,target=/meta,readonly <<EOT
  	set -e
  	if [ -z "${TARGETPLATFORM}" ]; then echo "TARGETPLATFORM is not set" && exit 1; fi
	if [ -f /binary-cache/$(cat /meta/executable) ];
		then cp /binary-cache/$(cat /meta/executable) /usr/bin/$(cat /meta/name);
		echo "FOUND BINARY CACHE"
	else
		echo "no binary cache found in /binary-cache/$(cat /meta/name) - building from source";
  		xx-go --wrap;
  		DESTDIR=/usr/bin GO_PKG=$(cat /meta/go-pkg) BIN_NAME=$(cat /meta/name) BIN_VERSION=$(cat /meta/version) BIN_REVISION=$(cat /meta/revision) GO_EXTRA_LDFLAGS="-s -w" ./hack/build;
  	fi;
  	xx-verify --static /usr/bin/$(cat /meta/name)
EOT

FROM scratch AS binaries-unix
ARG BIN_NAME
COPY --link --from=builder /usr/bin/${BIN_NAME} /${BIN_NAME}

FROM binaries-unix AS binaries-darwin
FROM binaries-unix AS binaries-linux

FROM scratch AS binaries-windows
ARG BIN_NAME
COPY --link --from=builder /usr/bin/${BIN_NAME} /${BIN_NAME}.exe

FROM binaries-$TARGETOS AS binaries
# enable scanning for this stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true


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

FROM gobase AS test-runner
COPY --link --from=binaries /${BIN_NAME} /usr/bin/
COPY --link --from=gotestsum /out/gotestsum /usr/bin/
COPY --link --from=meta /meta /meta
COPY . .
ARG TEST_RUN
ENV TEST_RUN=${TEST_RUN}
ARG DESTDIR
ENV DESTDIR=${DESTDIR}
ENTRYPOINT gotestsum \
	--format=standard-verbose \
	--jsonfile=${DESTDIR}/go-test-report.json \
	--junitfile=${DESTDIR}/junit-report.xml \
	-- -v -mod=vendor -coverprofile=${DESTDIR}/coverage-report.txt -covermode=atomic \
	./... -run ${TEST_RUN}

##################################################################
# RELEASE
##################################################################

FROM --platform=$BUILDPLATFORM alpine:latest AS releaser
WORKDIR /work
ARG TARGETPLATFORM
RUN --mount=from=binaries \
	--mount=type=bind,from=meta,source=/meta,target=/meta <<EOT
	set -e
	mkdir -p /out
	cp "$(cat /meta/name)"* "/out/$(cat /meta/executable)"
EOT

FROM scratch AS meta-out
COPY --from=meta /meta/ /

FROM scratch AS release
COPY --from=releaser /out/ /
COPY --from=meta /meta/buildrc.json /buildrc.json


##################################################################
# IMAGE
##################################################################

FROM scratch AS entry
ARG BIN_NAME
ENV BIN_NAME=${BIN_NAME}
COPY --link --from=meta /meta/buildrc.json /usr/bin/${BIN_NAME}/buildrc.json
COPY --link --from=builder /usr/bin/${BIN_NAME} /usr/bin/
ENTRYPOINT /usr/bin/${BIN_NAME}
