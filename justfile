##################################################################
# GENERATE
##################################################################

generate: vendor docs gen meta

gen:
    docker buildx bake update-gen

meta:
    docker buildx bake meta

vendor:
    ./hack/update-vendor

docs:
    ./hack/update-docs

##################################################################
# VALIDATE
##################################################################

validate: lint outdated validate-vendor validate-docs validate-gen

lint:
    docker buildx bake lint

validate-vendor:
    docker buildx bake validate-vendor

validate-docs:
    docker buildx bake validate-docs

validate-gen:
    docker buildx bake validate-gen

outdated:
	docker buildx bake outdated
	cat ./bin/outdated/outdated.txt

##################################################################
# TEST
##################################################################

test: unit-test integration-test

unit-test:
	docker buildx bake unit-test --set "*.args.DESTDIR=/out"
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -v ./bin:/out unit-test

integration-test:
	docker buildx bake integration-test --set "*.args.DESTDIR=/out"
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -v ./bin:/out integration-test

##################################################################
# BUILD
##################################################################

binaries:
    docker buildx bake binaries

binaries-cross:
    docker buildx bake binaries-cross

release:
    ./hack/release $(PLATFORM) $(TARGET)

local:
	docker buildx bake image-default

install: binaries
	binname=$(docker buildx bake _common --print | jq -cr '.target._common.args.BIN_NAME') && \
	./bin/build/${binname} install && ${binname} --version
