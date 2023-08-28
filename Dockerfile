# syntax=docker/dockerfile:1

ARG GO_VERSION=1.21.0
ARG XX_VERSION=1.2.1

ARG DOCKER_VERSION=24.0.2
ARG GOTESTSUM_VERSION=v1.9.0
ARG REGISTRY_VERSION=2.8.0
ARG BUILDKIT_VERSION=v0.11.6

# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS golatest

FROM --platform=$BUILDPLATFORM walteh/buildrc:0.12.1 as buildrc

FROM golatest AS gobase
COPY --from=xx / /
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apk add --no-cache file git bash
ENV GOFLAGS=-mod=vendor
ENV CGO_ENABLED=0
WORKDIR /src

FROM registry:$REGISTRY_VERSION AS registry

FROM moby/buildkit:$BUILDKIT_VERSION AS buildkit

FROM docker/buildx-bin:latest AS buildx-bin

FROM gobase AS meta
ARG TARGETPLATFORM
RUN --mount=type=bind,target=/src,rw <<EOT
    set -e
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

FROM binaries AS entry
ARG BIN_NAME
ENV BIN_NAME=${BIN_NAME}
COPY --link --from=builder /usr/bin/${BIN_NAME} /usr/bin/${BIN_NAME}
ENTRYPOINT [ "/usr/bin/${BIN_NAME}" ]

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

FROM gobase AS test
ENV SKIP_INTEGRATION_TESTS=1
RUN --mount=type=bind,target=. \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod <<EOT
	go test -v -coverprofile=/tmp/coverage.txt -covermode=atomic ./... &&
	go tool cover -func=/tmp/coverage.txt
EOT

FROM scratch AS test-coverage
COPY --from=test /tmp/coverage.txt /coverage.txt

FROM gobase AS docker
ARG TARGETPLATFORM
ARG DOCKER_VERSION
WORKDIR /opt/docker
RUN <<EOT
CASE=${TARGETPLATFORM:-linux/amd64}
DOCKER_ARCH=$(
	case ${CASE} in
	"linux/amd64") echo "x86_64" ;;
	"linux/arm/v6") echo "armel" ;;
	"linux/arm/v7") echo "armhf" ;;
	"linux/arm64/v8") echo "aarch64" ;;
	"linux/arm64") echo "aarch64" ;;
	"linux/ppc64le") echo "ppc64le" ;;
	"linux/s390x") echo "s390x" ;;
	*) echo "" ;; esac
)
echo "DOCKER_ARCH=$DOCKER_ARCH" &&
wget -qO- "https://download.docker.com/linux/static/stable/${DOCKER_ARCH}/docker-${DOCKER_VERSION}.tgz" | tar xvz --strip 1
EOT
RUN ./dockerd --version && ./containerd --version && ./ctr --version && ./runc --version

FROM gobase AS integration-test-base
ARG BIN_NAME
# https://github.com/docker/docker/blob/master/project/PACKAGERS.md#runtime-dependencies
RUN apk add --no-cache \
	btrfs-progs \
	e2fsprogs \
	e2fsprogs-extra \
	ip6tables \
	iptables \
	openssl \
	shadow-uidmap \
	xfsprogs \
	xz
COPY --link --from=gotestsum /out/gotestsum /usr/bin/
COPY --link --from=registry /bin/registry /usr/bin/
COPY --link --from=docker /opt/docker/* /usr/bin/
COPY --link --from=buildkit /usr/bin/buildkitd /usr/bin/
COPY --link --from=buildkit /usr/bin/buildctl /usr/bin/
COPY --link --from=binaries /${BIN_NAME} /usr/bin/
COPY --link --from=buildx-bin /buildx /usr/libexec/docker/cli-plugins/docker-buildx

FROM integration-test-base AS integration-test
COPY . .

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

FROM binaries
