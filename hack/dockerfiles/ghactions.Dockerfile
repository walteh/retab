# syntax=docker/dockerfile:labs

ARG GO_VERSION=1.21.0
ARG XX_VERSION=1.2.1

ARG DOCKER_VERSION=24.0.2
ARG GOTESTSUM_VERSION=v1.9.0
ARG REGISTRY_VERSION=2.8.0
ARG BUILDKIT_VERSION=v0.11.6

ARG BUILDRC_VERSION=0.14.1
ARG ACT_VERSION=0.2.50

FROM walteh/buildrc:${BUILDRC_VERSION} AS buildrc

##################################################################
# DOCKERD
##################################################################

# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS golatest

FROM --platform=$BUILDPLATFORM registry:$REGISTRY_VERSION AS registry

FROM --platform=$BUILDPLATFORM moby/buildkit:$BUILDKIT_VERSION AS buildkit

FROM --platform=$BUILDPLATFORM docker/buildx-bin:latest AS buildx-bin

FROM golatest AS gobase
COPY --from=xx / /
RUN apk add --no-cache file git bash
ENV GOFLAGS=-mod=vendor
ENV CGO_ENABLED=0
WORKDIR /src

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

FROM gobase AS dockerd
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
COPY --link --from=registry /bin/registry /usr/bin/
COPY --link --from=docker /opt/docker/* /usr/bin/
COPY --link --from=buildkit /usr/bin/buildkitd /usr/bin/
COPY --link --from=buildkit /usr/bin/buildctl /usr/bin/
COPY --link --from=buildx-bin /buildx /usr/libexec/docker/cli-plugins/docker-buildx

FROM alpine:latest AS act
COPY --from=buildrc . .
RUN apk add --no-cache curl bash
RUN <<EOT
	 set -e
# 	buildrc binary --organization=nektos --repository=act --version=${ACT_VERSION} --outfile=/usr/bin/act
#  /usr/bin/act --version
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | bash
act --version
mv $(which act) /usr/bin/act
EOT

# Base tools image with required packages
FROM alpine:latest AS tools
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
COPY --from=act /usr/bin/act /usr/bin/act

FROM alpine:latest AS validator
RUN apk add --no-cache curl bash coreutils ca-certificates git
RUN update-ca-certificates
ENV GIT_SSL_NO_VERIFY=true

RUN git clone https://github.com/walteh/retab

COPY --from=tools /usr/bin/act /usr/bin/act
COPY . /src
ENTRYPOINT cd /src && act


FROM docker:dind-rootless AS validate
ARG DESTDIR
ARG TARGETPLATFORM
COPY --from=tools /usr/bin/act /usr/bin/act
COPY . /src
RUN <<EOT
	 set -e
	 dockerd-rootless-setuptool.sh install
	 cd src && act
EOT
