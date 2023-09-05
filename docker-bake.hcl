variable "GO_VERSION" {
	default = "1.21.0"
}

variable "BUILDRC_VERSION" {
	default = "0.12.9"
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

variable "VENDOR_OUTPUT" {
	default = "."
}

variable "DOCS_OUTPUT" {
	default = "./docs/reference"
}

variable "GEN_OUTPUT" {
	default = "./gen"
}

variable "MOCKERY_OUTPUT" {
	default = "${GEN_OUTPUT}/mockery"
}

variable "BUF_OUTPUT" {
	default = "${GEN_OUTPUT}/buf"
}

variable "BUILD_OUTPUT" {
	default = "${DESTDIR}/build"
}

variable "RELEASE_OUTPUT" {
	default = "${DESTDIR}/release"
}

variable "PACKAGE_OUTPUT" {
	default = "${DESTDIR}/package"
}

variable "TEST_OUTPUT" {
	default = "${DESTDIR}/testreports"
}

variable "GITHUB_REPOSITORY" {
	default = ""
}

variable "GITHUB_RUN_ID" {
	default = ""
}

target "_common" {
	args = {
		GO_VERSION                    = GO_VERSION
		BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
		DOCKER_IMAGE                  = DOCKER_IMAGE
		BIN_NAME                      = BIN_NAME
		VENDOR_OUTPUT                 = VENDOR_OUTPUT
		DOCS_OUTPUT                   = DOCS_OUTPUT
		GEN_OUTPUT                    = GEN_OUTPUT
		BUILD_OUTPUT                  = BUILD_OUTPUT
		RELEASE_OUTPUT                = RELEASE_OUTPUT
		TEST_OUTPUT                   = TEST_OUTPUT
		BUILDRC_VERSION               = BUILDRC_VERSION
		PACKAGE_OUTPUT                = PACKAGE_OUTPUT
	}
}

target "_cross" {
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
	args = {
		BUILDKIT_MULTI_PLATFORM = true
	}
}

target "_attest" {
	attest = GITHUB_REPOSITORY != "" ? [
		"type=provenance,mode=max,builder-id=https://github.com/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}",
		"type=sbom"
		] : [
		"type=provenance,mode=max",
		"type=sbom"
	]
}

group "default" {
	targets = ["binaries"]
}

target "_vendor" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
	output     = ["type=local,dest=${VENDOR_OUTPUT}"]
	args = {
		BUILDX_EXPERIMENTAL = 1 // enables experimental cmds/flags for docs generation
		DESTDIR             = VENDOR_OUTPUT
	}
}

target "_docs" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/docs.Dockerfile"
	output     = [DOCS_OUTPUT]
	args = {
		FORMATS             = DOCS_FORMATS
		BUILDX_EXPERIMENTAL = 1 // enables experimental cmds/flags for docs generation
		DESTDIR             = DOCS_OUTPUT
	}
}

target "_mockery" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/mockery.Dockerfile"
	output     = [MOCKERY_OUTPUT]
	args = {
		DESTDIR = MOCKERY_OUTPUT
	}
}

target "_buf" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/buf.Dockerfile"
	output     = [BUF_OUTPUT]
	args = {
		DESTDIR = BUF_OUTPUT
	}
}

##################################################################
# VALIDATE
##################################################################

group "validate" {
	targets = ["lint", "validate-vendor", "validate-docs", "validate-mockery", "validate-buf", "outdated"]
}

target "lint" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/lint.Dockerfile"
	output     = ["type=cacheonly"]
}

target "validate-vendor" {
	inherits = ["_vendor"]
	target   = "validate"
	output   = ["type=cacheonly"]
}

target "validate-docs" {
	inherits = ["_docs"]
	target   = "validate"
	output   = ["type=cacheonly"]
}

target "validate-mockery" {
	inherits = ["_mockery"]
	target   = "validate"
	output   = ["type=cacheonly"]
}

target "validate-buf" {
	inherits = ["_buf"]
	target   = "validate"
	output   = ["type=cacheonly"]
}

target "outdated" {
	inherits   = ["_common"]
	dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
	target     = "outdated-output"
	output     = ["${DESTDIR}/outdated"]
}

##################################################################
# GENERATE
##################################################################

group "generate" {
	targets = ["generate-vendor", "generate-docs", "generate-mockery", "generate-buf"]
}

target "generate-vendor" {
	inherits = ["_vendor"]
	target   = "update"
}

target "generate-docs" {
	inherits = ["_docs"]
	target   = "update"
}

target "generate-mockery" {
	inherits = ["_mockery"]
	target   = "update"
}

target "generate-buf" {
	inherits = ["_buf"]
	target   = "update"
}

##################################################################
# METADATA
##################################################################

target "meta" {
	inherits = ["_common", "_cross"]
	target   = "meta"
	output   = ["${DESTDIR}"]
}

# Special target: https://github.com/docker/metadata-action#bake-definition
target "meta-helper" {
	tags = ["${DOCKER_IMAGE}:local"]
}

##################################################################
# BUILD
##################################################################

target "local" {
	inherits  = ["_common"]
	target    = "build"
	output    = [BUILD_OUTPUT]
	platforms = ["local"]
}

target "build" {
	inherits = ["_common", "_cross", "_attest"]
	target   = "build"
	output   = [BUILD_OUTPUT]
}

target "package" {
	inherits  = ["_common"]
	target    = "package"
	output    = [PACKAGE_OUTPUT]
	platforms = ["local"]
	contexts = {
		build = BUILD_OUTPUT
	}
}

##################################################################
# TESTING
##################################################################

target "tester" {
	inherits = ["_common"]
	target   = "tester"
}

/* target "integration-test" {
	inherits = ["_common"]
	target   = "tester"
	output   = ["type=docker,name=integration-test"]
	args = {
		TEST_ARGS = "-run=Integration"
		TEST_NAME = "integration"
	}
}

target "unit-test" {
	inherits = ["_common"]
	output   = ["type=docker,name=unit-test"]
	args = {
		TEST_ARGS = "-skip=Integration"
		TEST_NAME = "unit"
	}
} */

##################################################################
# IMAGE
##################################################################

target "image" {
	inherits  = ["meta-helper", "build"]
	target    = "entry"
	output    = ["type=image"]
	platforms = ["local"]
}

target "registry" {
	inherits = ["meta-helper", "_cross", "_attest"]
	target   = "entry"
	output   = ["type=image"]
}
