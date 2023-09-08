# syntax=docker/dockerfile:labs

ARG GO_VERSION=
ARG GO_PACKAGE=
ARG GOLANGCI_LINT_VERSION=v1.54.2

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci

FROM golangci as validate
RUN apk add --no-cache git gcc musl-dev
WORKDIR /app
RUN --mount=type=bind,target=/app --mount=target=/root/.cache,type=cache \
	golangci-lint run --timeout 5m0s --skip-dirs vendor
