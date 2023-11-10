##################################################################
# GENERATE
##################################################################

generate:
	docker buildx bake generate

generate-buf:
	docker buildx bake generate-buf

generate-mockery:
    docker buildx bake generate-mockery

generate-meta:
	docker buildx bake meta

generate-vendor:
	docker buildx bake generate-vendor

generate-docs:
    docker buildx bake generate-docs

##################################################################
# VALIDATE
##################################################################

validate:
	docker buildx bake validate --no-cache

lint:
    docker buildx bake lint

validate-vendor:
    docker buildx bake validate-vendor

validate-docs:
    docker buildx bake validate-docs

validate-gen:
    docker buildx bake validate-gen

ghactions:
	mkdir -p ./bin/images && \
	docker buildx bake ghactions --set "*.output=type=docker,dest=./bin/images/runner.tar,name=runner" --set "*.platform=linux/amd64" && \
	docker load -i ./bin/images/runner.tar && \
	docker run --platform=linux/amd64 --network host -v /var/run/docker.sock:/var/run/docker.sock -v ./bin/test-output:/out runner

ghaction:
	docker buildx bake ghaction

outdated:
	docker buildx bake outdated
	cat ./bin/outdated/outdated.txt

##################################################################
# TEST
##################################################################

test-pkg PACKAGE:
	just test all {{PACKAGE}}

test CASE PACKAGE:
	docker buildx bake test-{{CASE}} && \
	docker load -i ./bin/test-{{CASE}}.tar && \
	docker run -e PKGS='{{PACKAGE}}' --network host -v /var/run/docker.sock:/var/run/docker.sock -v ./bin/test-reports:/out test-{{CASE}}:latest  && \
	echo "test-{{CASE}}: {{PACKAGE}}"

test-all:
	pkgs=$(go list -test ./... | grep "\.test$" | jq -R -c -s 'split("\n") | map(select(. != "")) | map(split("/")[-1]) | map(split(".")[0])') && \
	just test all "$pkgs"

test-fuzz:
	pkgs=$(go list -test ./... | grep "\.test$" | jq -R -c -s 'split("\n") | map(select(. != "")) | map(split("/")[-1]) | map(split(".")[0])') && \
	just test fuzz "$pkgs"

##################################################################
# BUILD
##################################################################

build:
    docker buildx bake build

build-local:
    docker buildx bake build --set "*.platform=local"

package:
	BUILD_OUTPUT=$(mktemp -d -t release-XXXXXXXXXX) && \
	docker buildx bake build --set "*.output=${BUILD_OUTPUT}" && \
	docker buildx bake package --set "*.contexts.build=${BUILD_OUTPUT}" && \
	docker buildx bake registry --set "*.contexts.build=${BUILD_OUTPUT}" && \
	rm -rf ${BUILD_OUTPUT}

local:
	docker buildx bake image-default

install: build-local
	binname=$(docker buildx bake _common --print | jq -cr '.target._common.args.BIN_NAME') && \
	./bin/build/darwin_arm64/${binname} install && ${binname} --version
