# syntax=docker/dockerfile:labs

##################################################################
# SETUP
##################################################################

ARG GO_VERSION=
ARG XX_VERSION=
ARG GOTESTSUM_VERSION=
ARG BUILDRC_VERSION=
ARG BIN_NAME=
ARG DART_VERSION=

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS golatest
FROM --platform=$BUILDPLATFORM walteh/buildrc:${BUILDRC_VERSION} as buildrc
FROM --platform=$BUILDPLATFORM alpine:latest AS alpinelatest
FROM --platform=$BUILDPLATFORM busybox:musl AS musl
FROM --platform=$BUILDPLATFORM dart:${DART_VERSION} as dart

FROM golatest AS gobase
COPY --from=xx / /
# COPY --from=buildrc /usr/bin/ /usr/bin/
RUN apk add --no-cache file git bash
ENV GOFLAGS=-mod=vendor
ENV CGO_ENABLED=0
WORKDIR /src

##################################################################
# BUILD
##################################################################

FROM gobase AS metarc
ARG TARGETPLATFORM BUILDPLATFORM BIN_NAME
RUN --mount=type=bind,target=/src,readonly <<SHELL

	mkdir -p /meta

	echo "$(git rev-list HEAD -1)" > /meta/revision

	# if we are detached, then rev list -2 (git symbolic-ref -q HEAD)
	if [ "$(git symbolic-ref -q HEAD || echo "d")" != "" ]; then
		echo "$(git rev-list HEAD -2 | tail -n 1)" > /meta/revision
	fi

	echo "$(git describe "$(cat /meta/revision)" --tags || echo "v0.0.0-local+$(git rev-parse --short HEAD)")$(git diff --quiet || echo '.dirty')" > /meta/version
	echo "${BIN_NAME}_$(cat /meta/version)_${TARGETPLATFORM}" | sed -e 's|/|-|g' > /meta/artifact
	echo "${BIN_NAME}" > /meta/executable
	echo "$(go list -m)" > /meta/go-pkg

	# if target contains  windows, then add .exe
	if [ "$(echo ${TARGETPLATFORM} | grep -i windows)" != "" ]; then
		echo "$(cat /meta/executable).exe" > /meta/executable
	fi

	echo "========== [meta] =========="
	cat /meta/*
SHELL
FROM scratch AS meta
COPY --link --from=metarc /meta /

FROM gobase AS builder
ARG TARGETPLATFORM
COPY --link --from=meta . /meta
RUN --mount=type=bind,target=. \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod  <<SHELL
	#!/bin/sh
	set -e -o pipefail

	export CGO_ENABLED=0
 	xx-go --wrap;
	GO_PKG=$(cat /meta/go-pkg);
	LDFLAGS="-s -w -X ${GO_PKG}/version.Version=$(cat /meta/version) -X ${GO_PKG}/version.Revision=$(cat /meta/revision) -X ${GO_PKG}/version.Package=${GO_PKG}";
	go build -mod vendor -trimpath -ldflags "$LDFLAGS" -o /out/$(cat /meta/executable) ./cmd;
  	xx-verify --static /out/$(cat /meta/executable);
SHELL

FROM musl AS symlink
COPY --link --from=meta . /meta
RUN <<SHELL
	#!/bin/sh
	set -e -o pipefail

	mkdir -p /out/symlink
	ln -s ../$(cat /meta/executable) /out/symlink/executable
SHELL

FROM scratch AS build-unix
COPY --from=builder /out /
COPY --from=symlink /out /

FROM build-unix AS build-darwin
FROM build-unix AS build-linux
FROM build-unix AS build-freebsd
FROM build-unix AS build-openbsd
FROM build-unix AS build-netbsd
FROM build-unix AS build-ios

FROM scratch AS build-windows
COPY --from=builder /out/ /
COPY --from=symlink /out /

FROM build-$TARGETOS AS build
# enable scanning for this stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
COPY --from=metarc /meta/artifact /artifact


##################################################################
# TESTING
##################################################################

FROM gobase AS test2json
ARG GOTESTSUM_VERSION
ENV GOFLAGS=
RUN --mount=target=/root/.cache,type=cache <<SHELL
	#!/bin/sh
	set -e -o pipefail

	CGO_ENABLED=0 go build -o /out/test2json -ldflags="-s -w" cmd/test2json
SHELL

FROM gobase AS gotestsum
ARG GOTESTSUM_VERSION
ENV GOFLAGS=
RUN --mount=target=/root/.cache,type=cache \
	go install gotest.tools/gotestsum@${GOTESTSUM_VERSION}

FROM gobase AS test-builder
ARG BIN_NAME
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev libc6-compat
RUN mkdir -p /out
RUN --mount=type=bind,target=. \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod <<SHELL
	#!/bin/sh
	set -e -o pipefail

	for dir in $(go list -test -f '{{if or .ForTest}}{{.Dir}}{{end}}' ./...); do
		pkg=$(echo $dir | sed -e 's/.*\///')
		echo "========== [pkg:${pkg}] =========="
		go test -c -v -cover -fuzz -race -vet='' -covermode=atomic -mod=vendor "$dir" -o /out;
	done
SHELL

FROM scratch AS test-build
COPY --from=test-builder /out /tests
COPY --from=test2json /out /bins
COPY --from=gotestsum /go/bin /bins

FROM docker:dind AS test
ARG CASE_INFOS
RUN	apk add --no-cache jq
RUN mkdir -p /dat/case_info
RUN echo "${CASE_INFOS}" | jq -r '.[] | @base64' | while read case; do \
	echo ${case} | base64 -d > /dat/case_info/$(echo ${case} | base64 -d | jq -r '.name'); \
	done
COPY --from=test-build /tests /usr/bin/
COPY --from=test-build /bins /usr/bin/
COPY --from=build . /usr/bin/

COPY <<-"SHELL" /xxx/run
	#!/bin/sh
	set -e -o pipefail

	GOVERSION=$1
	PKGS=$2
	CASES=$3

	# if pkgs = "empty" then use default
	if [ "${PKGS}" = "empty" ]; then
		# find all executables in /usr/bin that end with .test, remove the .test
		PKGS=$(find /usr/bin -maxdepth 1 -type f -name '*.test' -exec basename {} \; | sed -e 's/\.test//g' | sort | uniq | jq -R . | jq -s .)
	fi

	# if cases = "empty" then use default
	if [ "${CASES}" = "empty" ]; then
		# find all the non executable files in /dat/case_info
		CASES=all
		echo "{\"name\":\"all\",\"args\":\"\"}" > "/dat/case_info/all"
	fi

	for CASE in $(echo "${CASES}" | jq -r '.[]' || echo "$CASES"); do

		case_info=$(cat /dat/case_info/${CASE})

		if [ "${case_info}" = "" ]; then
			echo "case not found"
			exit 1
		fi

		export NAME=$(echo "${case_info}" | jq -r '.name' > /dat/name)
		export ARGS=$(echo "${case_info}" | jq -r '.args' > /dat/args)

		for PKG in $(echo "${PKGS}" | jq -r '.[]' || echo "$PKGS"); do
			name=$(cat /dat/name)
			funcs=$(if [ "${name}" = "fuzz" ]; then "/usr/bin/${PKG}.test" -test.list=Fuzz 2> /dev/null; else echo "-"; fi)
			for FUNC in $funcs; do
				echo ""
				if [ "${name}" = "fuzz" ] && [ "${FUNC}" = "-" ]; then continue; fi
				n=$(if [ "${name}" = "fuzz" ]; then echo "[fuzz:${FUNC}] "; else echo ""; fi)
				echo "========= [pkg:${PKG}] [case:$(cat /dat/name)] $n=========="

				filename=$(if [ "${name}" = "fuzz" ]; then echo "${PKG}-fuzz-${FUNC}"; else echo "${PKG}-${name}"; fi)

				fuzzfunc=$(if [ "${name}" = "fuzz" ]; then echo "-test.fuzz=${FUNC} -test.run=${FUNC}"; fi)

				/usr/bin/gotestsum --format=standard-verbose \
					--jsonfile=/out/go-test-report-${filename}.json \
					--junitfile=/out/junit-report-${filename}.xml \
					--raw-command -- /usr/bin/test2json -t -p ${PKG}  /usr/bin/${PKG}.test $(cat /dat/args) -test.bench=. -test.timeout=10m  ${fuzzfunc} \
					-test.v=test2json -test.coverprofile=/out/coverage-report-${filename}.txt \
					-test.outputdir=/out;
			done;
		done;
	done;

	echo ""
SHELL
RUN chmod +x /xxx/run
ARG GO_VERSION
ENV GOVERSION=${GO_VERSION}
ENV PKGS=empty CASES=empty
ENTRYPOINT /xxx/run "${GOVERSION}" "${PKGS}" ${CASES}

##################################################################
# RELEASE
##################################################################

FROM alpinelatest AS packager
RUN apk add --no-cache file tar jq
COPY --link --from=build . /src/
RUN <<SHELL
	#!/bin/sh
	set -e -o pipefail

	if [ -f /src/artifact ]; then
		searchdir="/src/"
	else
		searchdir="/src/*/"
	fi
	mkdir -p /out
	for pdir in ${searchdir}; do
		(
			cd "${pdir}"
			artifact="$(cat ./artifact)"
			tar -czvf "/out/${artifact}.tar.gz" .
		)
	done

	(
		cd /out
		find . -type f \( -name '*.tar.gz' \) -exec sha256sum -b {} \; >./checksums.txt
		sha256sum -c checksums.txt
	)
SHELL

FROM scratch AS package
COPY --link --from=packager /out/ /

##################################################################
# IMAGE
##################################################################

FROM scratch AS entry
ARG BUILDKIT_SBOM_SCAN_STAGE=true
ARG TARGETOS TARGETARCH TARGETVARIANT BIN_NAME
ARG TGT=${TARGETOS}_${TARGETARCH}*${TARGETVARIANT}
COPY --from=build /symlink* /usr/bin/symlink/
COPY --from=build /${BIN_NAME}* /usr/bin/
COPY --from=build /*.json /usr/bin/
COPY --from=build /${TGT} /usr/bin/
ENTRYPOINT ["/usr/bin/symlink/executable"]
