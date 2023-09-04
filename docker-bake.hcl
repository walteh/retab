variable "GO_VERSION" {
	default = "1.21.0"
}
variable "DOCS_FORMATS" {
	default = "md"
}

variable "DESTDIR" {
	default = "./bin"
}

variable "GENDIR" {
	default = "./gen"
}


variable "DOCKER_IMAGE" {
	default = "walteh/retab"
}

variable "BIN_NAME" {
	default = "retab"
}

target "_common" {
	args = {
		GO_VERSION                    = GO_VERSION
		BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
		DOCKER_IMAGE                  = DOCKER_IMAGE
		BIN_NAME                      = BIN_NAME
	}
}

group "default" {
	targets = ["binaries"]
}

group "validate" {
	targets = ["lint", "validate-vendor", "validate-docs"]
}

target "lint" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/lint.Dockerfile"
	output     = ["type=cacheonly"]
}

target "validate-vendor" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
	target     = "validate"
	output     = ["type=cacheonly"]
}

target "validate-docs" {
	inherits = ["_common"]
	args = {
		FORMATS             = DOCS_FORMATS
		BUILDX_EXPERIMENTAL = 1 // enables experimental cmds/flags for docs generation
	}
	dockerfile = "./hack/dockerfiles/docs.Dockerfile"
	target     = "validate"
	output     = ["type=cacheonly"]
}


target "validate-gen" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/gen.Dockerfile"
	target     = "validate"
	output     = ["type=cacheonly"]
}

target "update-vendor" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
	target     = "update"
	output     = ["."]
}

target "update-docs" {
	inherits = ["_common"]
	args = {
		FORMATS             = DOCS_FORMATS
		BUILDX_EXPERIMENTAL = 1 // enables experimental cmds/flags for docs generation
	}
	dockerfile = "./hack/dockerfiles/docs.Dockerfile"
	target     = "update"
	output     = ["./docs/reference"]
}

target "update-gen" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/gen.Dockerfile"
	target     = "update"
	output     = ["${GENDIR}"]
}

target "outdated" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
	target     = "outdated-output"
	output     = ["${DESTDIR}/outdated"]
}

target "binaries" {
	inherits  = ["_common"]
	target    = "binaries"
	output    = ["${DESTDIR}/build"]
	platforms = ["local"]
}

target "binaries-cross" {
	inherits = ["binaries"]
	platforms = [
		"darwin/amd64",
		"darwin/arm64",
		"linux/amd64",
		"linux/arm/v6",
		"linux/arm/v7",
		"linux/arm64",
		"linux/ppc64le",
		"linux/riscv64",
		"linux/s390x",
		"windows/amd64",
		"windows/arm64"
	]
}

target "release" {
	inherits = ["binaries-cross"]
	target   = "release"
	output   = ["${DESTDIR}/release"]
}

target "image" {
	inherits = ["meta-helper", "binaries"]
	output   = ["type=image"]
}

target "image-default" {
	inherits  = ["meta-helper", "binaries"]
	target    = "entry"
	output    = ["type=image"]
	platforms = ["linux/arm64"]
}

target "image-cross" {
	inherits = ["meta-helper", "binaries-cross"]
	target   = "entry"
	output   = ["type=image"]
	attest = [
		"type=provenance",
		"type=sbom"
	]
}

target "image-local" {
	inherits = ["image"]
	output   = ["type=docker"]
}

target "meta" {
	inherits = ["_common"]
	target   = "meta-out"
	output   = ["meta"]
}

# Special target: https://github.com/docker/metadata-action#bake-definition
target "meta-helper" {
	tags = ["${DOCKER_IMAGE}:local"]
}

target "tester" {
	inherits = ["_common"]
	target   = "test-runner"
}

target "integration-test" {
	inherits = ["_common", "tester"]
	args = {
		TEST_ARGS = "-run=Integration"
		DESTDIR   = "${DESTDIR}/testreports/integration"
	}
	output = ["type=docker,name=integration-test"]
}

target "unit-test" {
	inherits = ["_common", "tester"]
	args = {
		TEST_ARGS = "-skip=Integration"
		DESTDIR   = "${DESTDIR}/testreports/unit"
	}
	output = ["type=docker,name=unit-test"]
}
