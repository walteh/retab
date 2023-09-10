# syntax=docker/dockerfile:labs

##################################################################
# SETUP
##################################################################

ARG GO_VERSION=
ARG XX_VERSION=
ARG GOTESTSUM_VERSION=

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
COPY --link --from=metarc /meta /

FROM gobase AS builder
ARG TARGETPLATFORM
COPY --link --from=meta . /meta
RUN --mount=type=bind,target=. \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod  <<EOT
  	set -e
	export CGO_ENABLED=0
 	xx-go --wrap;
	GO_PKG=$(cat /meta/go-pkg);
	LDFLAGS="-s -w -X ${GO_PKG}/version.Version=$(cat /meta/version) -X ${GO_PKG}/version.Revision=$(cat /meta/revision) -X ${GO_PKG}/version.Package=${GO_PKG}";
	go build -mod vendor -trimpath -ldflags "$LDFLAGS" -o /out/$(cat /meta/executable) ./cmd;
  	xx-verify --static /out/$(cat /meta/executable);
EOT

FROM scratch AS build-unix
ARG BIN_NAME
COPY --link --from=builder /out/${BIN_NAME} /${BIN_NAME}

FROM build-unix AS build-darwin
FROM build-unix AS build-linux
FROM build-unix AS build-freebsd
FROM build-unix AS build-openbsd
FROM build-unix AS build-netbsd
FROM build-unix AS build-ios

FROM scratch AS build-windows
ARG BIN_NAME
COPY --link --from=builder /out/${BIN_NAME} /${BIN_NAME}.exe

FROM build-$TARGETOS AS build
# enable scanning for this stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
COPY --link --from=meta /buildrc.json /


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

FROM scratch AS test-build
COPY --from=test-builder /out /tests
COPY --from=gotestsum /out /bins
COPY --from=test2json /out /bins

FROM alpine AS case
ARG NAME= ARGS= E2E=
COPY --from=test-build /bins /bins
COPY --from=test-build /tests /bins
COPY --from=build . /bins

RUN <<EOT
	set -e -x -o pipefail
	mkdir -p /dat

	echo "${ARGS}" > /dat/args
	echo "${E2E}" > /dat/e2e
	echo "${NAME}" > /dat/name

	for file in /bins/*; do
		chmod +x $file
	done
EOT

# FROM scratch AS case-build
# COPY --link --from=case-builder /dat /dat
# COPY --link --from=case-builder /bins /usr/bin

# FROM alpine AS case-build-runner
# ARG GO_VERSION
# ENV GOVERSION=${GO_VERSION}
# COPY --link --from=case . .
# RUN <<EOT
# 	set -e -x -o pipefail
# 	PKG=$(cat /dat/pkg)
# 	NAME=$(cat /dat/name)
# 	ARGS=$(cat /dat/args)
# 	/usr/bin/gotestsum --format=standard-verbose \
# 		--jsonfile=/out/go-test-report-$PKG-$NAME.json \
# 		--junitfile=/out/junit-report-$PKG-$NAME.xml \
# 		--raw-command -- /usr/bin/test2json -t -p $PKG /usr/bin/$PKG.test  -test.bench=.  -test.timeout=10m \
# 		-test.v -test.coverprofile=/out/coverage-report-$PKG-$NAME.txt $ARGS \
# 		-test.outputdir=/out;
# 	done
# EOT

# FROM scratch AS case-built
# COPY --link --from=case-build-runner /out /

FROM alpine AS test
ARG GO_VERSION
ENV GOVERSION=${GO_VERSION}
RUN apk add --no-cache jq
COPY --from=case /bins /usr/bin
COPY --from=case /dat /dat
ENV PKGS=
ENTRYPOINT for PKG in $(echo "${PKGS}" | jq -r '.[]' || echo "$PKGS"); do \
	export E2E=$(cat /dat/e2e) && \
	echo "" && echo "---------- ${PKG}: $(cat /dat/name) ----------" && /usr/bin/gotestsum --format=standard-verbose \
	--jsonfile=/out/go-test-report-${PKG}-$(cat /dat/name).json \
	--junitfile=/out/junit-report-${PKG}-$(cat /dat/name).xml \
	--raw-command -- /usr/bin/test2json -t -p ${PKG} /usr/bin/${PKG}.test  -test.bench=.  -test.timeout=10m \
	-test.v -test.coverprofile=/out/coverage-report-${PKG}-$(cat /dat/name).txt $(cat /dat/args) \
	-test.outputdir=/out; done

##################################################################
# RELEASE
##################################################################

FROM alpine AS packager
RUN apk add --no-cache file tar jq
COPY --link --from=build . /src/
RUN <<EOT
	set -e -x -o pipefail
	if [ -f /src/buildrc.json ]; then
		searchdir="/src/"
	else
		searchdir="/src/*/"
	fi
	mkdir -p /out
	for pdir in ${searchdir}; do
		(
			cd "${pdir}"
			artifact="$(jq -r '.artifact' ./buildrc.json)"
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
COPY --link --from=packager /out/ /

##################################################################
# IMAGE
##################################################################

FROM scratch AS entry-unix
ARG BIN_NAME
COPY --link --from=build . /usr/bin
ENTRYPOINT /usr/bin/${BIN_NAME}

FROM scratch AS entry-windows
ARG BIN_NAME
COPY --link --from=build . /usr/bin
ENTRYPOINT /usr/bin/${BIN_NAME}.exe

FROM entry-unix AS entry-darwin
FROM entry-unix AS entry-linux
FROM entry-unix AS entry-freebsd
FROM entry-unix AS entry-openbsd
FROM entry-unix AS entry-netbsd
FROM entry-unix AS entry-ios

FROM entry-$TARGETOS AS entry
# enable scanning for this stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true


