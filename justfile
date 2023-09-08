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
	docker buildx bake validate

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

cni PACKAGE:
	# docker buildx rm cni2 || true
	# docker buildx build --tag buildkit-cni:local --file ./hack/dockerfiles/cni.Dockerfile --load .
	# docker buildx create --use --bootstrap --name cni2 --driver docker-container --driver-opt "image=buildkit-cni:local" --driver-opt network=host --buildkitd-flags "--oci-worker-net=cni" --config ./buildkitd.toml
	docker buildx bake integration2 --set "*.args.PKG={{PACKAGE}}"

test PACKAGE:
	docker buildx bake integration3 --set "*.args.PKG={{PACKAGE}}"

unit-test2 PACKAGE:
	mkdir -p ./bin/test-images && \
	docker buildx bake test  && \
	docker build --target tester . --output type=local,dest=./bin/help --build-arg "PKG={{PACKAGE}}" --build-arg "NAME=integration" --build-arg "ARGS=-test.run=Integration" --allow "network.host"

unit-test PACKAGE:
	mkdir -p ./bin/test-images && \
	docker buildx bake unit --set "*.args.PKG={{PACKAGE}}" && \
	docker load -i ./bin/test-images/unit.tar && \
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -v ./bin/test-output:/out unit

integration-test PACKAGE:
	mkdir -p ./bin/test-images && \
	docker buildx bake integration --set "*.args.PKG={{PACKAGE}}" && \
	docker load -i ./bin/test-images/integration.tar && \
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -v ./bin/test-output:/out integration

e2e-test PACKAGE:
	mkdir -p ./bin/test-images && \
	docker buildx bake e2e --set "*.args.PKG={{PACKAGE}}" && \
	docker load -i ./bin/test-images/e2e.tar && \
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -v ./bin/test-output:/out e2e

##################################################################
# BUILD
##################################################################

build:
    docker buildx bake build

package:
	BUILD_OUTPUT=$(mktemp -d -t release-XXXXXXXXXX) && \
	docker buildx bake build --set "*.output=${BUILD_OUTPUT}" && \
	docker buildx bake package --set "*.contexts.build=${BUILD_OUTPUT}" && \
	rm -rf ${BUILD_OUTPUT}

local:
	docker buildx bake image-default

install: build
	binname=$(docker buildx bake _common --print | jq -cr '.target._common.args.BIN_NAME') && \
	./bin/build/${binname} install && ${binname} --version
