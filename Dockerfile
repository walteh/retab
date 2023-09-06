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

FROM --platform=$BUILDPLATFORM walteh/buildrc:0.13.0 as buildrc

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
COPY --link --from=gotestsum /out /usr/bin/
RUN --mount=type=bind,target=. \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod \
	gotestsum \
	--format=standard-verbose \
	--jsonfile=/reports/go-test-report.json \
	--junitfile=/reports/junit-report.xml \
	-- -coverprofile=/reports/coverage-report.txt -c -race -vet='' -covermode=atomic -mod=vendor ./... -o /out

FROM scratch AS test
COPY --link --from=test-builder /reports /reports
COPY --link --from=test-builder /out /
COPY --link --from=gotestsum /out /
COPY --link --from=test2json /out /
COPY --link --from=build . /

FROM alpine:latest AS tester
COPY --link --from=test . /usr/bin/
ENV PATH=/usr/bin:$PATH
ARG GO_VERSION
ENV PKG= NAME= ARGS= GOVERSION=${GO_VERSION}
ENTRYPOINT gotestsum \
	--format=standard-verbose \
	--jsonfile=/out/go-test-report-${PKG##*/}-${NAME}.json \
	--junitfile=/out/junit-report-${PKG##*/}-${NAME}.xml \
	--raw-command -- test2json -t -p pkgname ${PKG##*/}.test  -test.bench=.  -test.timeout=10m \
	-test.v -test.coverprofile=/out/coverage-report-${PKG##*/}-${NAME}.txt ${ARGS} \
	-test.outputdir=/out

##################################################################
# RELEASE
##################################################################

FROM alpine:latest AS packager
ARG BUILDKIT_MULTI_PLATFORM
RUN apk add --no-cache file tar jq
COPY --link  --from=build . /src/
RUN <<EOT
	set -e
	if [ "$BUILDKIT_MULTI_PLATFORM" != true ] ; then
		searchdir="/src/"
	else
		searchdir="/src/*/"
	fi
	mkdir -p /out
	for pdir in "${searchdir}"; do
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
