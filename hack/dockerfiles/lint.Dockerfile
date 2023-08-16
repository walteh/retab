# syntax=docker/dockerfile:1

ARG GO_VERSION=1.20.7
ARG GO_PACKAGE=

FROM golang:${GO_VERSION}-alpine
RUN apk add --no-cache git gcc musl-dev
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.51.1
WORKDIR /go/src/${GO_PACKAGE}
RUN --mount=target=/go/src/${GO_PACKAGE} --mount=target=/root/.cache,type=cache \
	golangci-lint run
