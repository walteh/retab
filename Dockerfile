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
# BUILD
##################################################################

FROM gobase AS metarc
ARG TARGETPLATFORM
RUN --mount=type=bind,target=/src,rw <<EOT
	buildrc full --git-dir=/src --files-dir=/meta
EOT

FROM scratch AS meta
COPY --from=metarc /meta/ /meta/

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
		LDFLAGS="-s -w -X $(cat /meta/go-pkg)/version.Version=$(cat /meta/version) -X $(cat /meta/go-pkg)/version.Revision=$(cat /meta/revision) -X $(cat /meta/go-pkg)/version.Package=$(cat /meta/go-pkg)";
		CGO_ENABLED=0 go build -mod vendor -trimpath -ldflags "$LDFLAGS" -o /usr/bin/$(cat /meta/name) ./cmd;
  	fi;
  	xx-verify --static /usr/bin/$(cat /meta/name)
EOT

FROM scratch AS build-unix
ARG BIN_NAME
COPY --link --from=builder /usr/bin/${BIN_NAME} /${BIN_NAME}

FROM build-unix AS build-darwin
FROM build-unix AS build-linux

FROM scratch AS build-windows
ARG BIN_NAME
COPY --link --from=builder /usr/bin/${BIN_NAME} /${BIN_NAME}.exe

FROM build-$TARGETOS AS build
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
RUN apk add --no-cache gcc musl-dev libc6-compat
COPY --link --from=build /${BIN_NAME} /usr/bin/
COPY --link --from=gotestsum /out/gotestsum /usr/bin/
COPY --link --from=meta /meta /meta
COPY . .
ARG TEST_ARGS
ENV TEST_ARGS=${TEST_ARGS}
ARG TEST_NAME
ENV TEST_NAME=${TEST_NAME}
ENTRYPOINT CGO_ENABLED=1 gotestsum \
	--format=standard-verbose \
	--jsonfile=/out/${TEST_NAME}-go-test-report.json \
	--junitfile=/out/${TEST_NAME}-junit-report.xml \
	-- -v -mod=vendor -coverprofile=/out/${TEST_NAME}-coverage-report.txt \
	-race -covermode=atomic --timeout=10m -vet='' -parallel=100 -bench=. -benchmem -fuzzminimizetime=100x -fullpath  \
	./... ${TEST_ARGS} -tags=${TEST_NAME}

##################################################################
# RELEASE
##################################################################

FROM --platform=$BUILDPLATFORM alpine:latest AS releaser
WORKDIR /work
ARG TARGETPLATFORM
RUN --mount=from=build \
	--mount=type=bind,from=meta,source=/meta,target=/meta <<EOT
	set -e
	mkdir -p /out
	cp "$(cat /meta/name)"* "/out/$(cat /meta/executable)"
EOT

FROM scratch AS release
COPY --from=releaser /out/ /
COPY --from=meta /meta/buildrc.json /buildrc.json

FROM alpine:latest AS packager
RUN apk add --no-cache file tar jq
COPY --from=released . /src/
RUN <<EOT
	set -e
	mkdir -p /out
	for pdir in /src/*/; do
	(
		cd "${pdir}"
		artifact="$(jq -r '.artifact' buildrc.json)"
		tar -czf "/out/${artifact}.tar.gz" .
	)
	done
EOT

FROM scratch AS package
COPY --from=packager /out/ /

FROM alpine:latest AS checksumer
COPY --from=packaged . /src/
RUN <<EOT
	cd /src/
	find . -type f \( -name '*.tar.gz' \) -exec sha256sum -b {} \; >./checksums.txt
	sha256sum -c checksums.txt
EOT

FROM scratch AS checksum
COPY --from=checksumer /src/checksums.txt /checksums.txt

##################################################################
# IMAGE
##################################################################

FROM scratch AS entry
ARG BIN_NAME
ENV BIN_NAME=${BIN_NAME}
COPY --link --from=meta /meta/buildrc.json /usr/bin/${BIN_NAME}/buildrc.json
COPY --link --from=builder /usr/bin/${BIN_NAME} /usr/bin/
ENTRYPOINT /usr/bin/${BIN_NAME}
