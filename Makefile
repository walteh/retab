
.PHONY: all
all: binaries

.PHONY: build
build:
	./hack/build

.PHONY: shell
shell:
	./hack/shell

.PHONY: binaries
binaries:
	docker buildx bake binaries

.PHONY: binaries-cross
binaries-cross:
	docker buildx bake binaries-cross

.PHONY: install
install: binaries
	mkdir -p ~/.docker/cli-plugins
	install bin/build/buildx ~/.docker/cli-plugins/docker-buildx

.PHONY: release
release:
	./hack/release

.PHONY: validate-all
validate-all: lint test validate-vendor validate-docs validate-gen

.PHONY: lint
lint:
	docker buildx bake lint

.PHONY: test
test:
	./hack/test

.PHONY: test-unit
test-unit:
	TESTPKGS=./... SKIP_INTEGRATION_TESTS=1 ./hack/test

.PHONY: test
test-integration:
	TESTPKGS=./tests ./hack/test

.PHONY: validate-vendor
validate-vendor:
	docker buildx bake validate-vendor

.PHONY: validate-docs
validate-docs:
	docker buildx bake validate-docs

.PHONY: validate-authors
validate-authors:
	docker buildx bake validate-authors

.PHONY: validate-gen
validate-gen:
	docker buildx bake validate-gen

.PHONY: test-driver
test-driver:
	./hack/test-driver

.PHONY: vendor
vendor:
	./hack/update-vendor

.PHONY: docs
docs:
	./hack/update-docs

.PHONY: authors
authors:
	docker buildx bake update-authors

.PHONY: mod-outdated
mod-outdated:
	docker buildx bake mod-outdated

.PHONY: gen
gen:
	docker buildx bake update-gen
