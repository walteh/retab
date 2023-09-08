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


variable "TEST_CASE_OUTPUT" {
	default = "${DESTDIR}/test-cases"
}


variable "TEST_IMAGES_OUTPUT" {
	default = "${DESTDIR}/test-images"
}

variable "GITHUB_REPOSITORY" {
	default = ""
}

variable "GITHUB_RUN_ID" {
	default = ""
}

variable "HTTP_PROXY" {
	default = ""
}
variable "HTTPS_PROXY" {
	default = ""
}
variable "NO_PROXY" {
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
		TEST_IMAGES_OUTPUT            = TEST_IMAGES_OUTPUT
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
	targets = ["build"]
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

target "ghactions" {
	inherits   = ["_common"]
	target     = "validator"
	dockerfile = "./hack/dockerfiles/ghactions.Dockerfile"
	output     = ["type=cacheonly"]
}

target "ghaction" {
	inherits   = ["_common"]
	target     = "validate"
	dockerfile = "./hack/dockerfiles/ghactions.Dockerfile"
	output     = ["type=cacheonly"]
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

/* target "test" {
	inherits = ["_common"]
	target   = "test"
} */

target "test" {
	inherits = ["_common"]
	target   = "case"
	matrix = {
		item = [
			{ ARGS = "-test.skip=Integration -test.skip=E2E", NAME = "unit" },
			{ ARGS = "-test.run=Integration", NAME = "integration" },
			{ ARGS = "-test.run=E2E", NAME = "e2e" }
		]
	}
	name = "${item.NAME}"
	args = {
		ARGS = item.ARGS
		NAME = item.NAME
		E2E  = item.NAME == "e2e" ? 1 : 0
	}
	output = ["type=local,dest=${DESTDIR}/test-cases/${item.NAME}"]
}

/* target "unit" {
	inherits = ["_common", "test"]
	target   = "tester"
	args = {
		ARGS = "-test.skip=Integration -test.skip=E2E"
		NAME = "unit"
	}
	output = ["type=docker,name=unit,dest=${TEST_IMAGES_OUTPUT}/unit.tar"]
}


target "integration" {
	inherits = ["_common", "test"]
	target   = "tester"
	args = {
		ARGS = "-test.run=Integration"
		NAME = "integration"
	}
	output = ["type=docker,name=integration,dest=${TEST_IMAGES_OUTPUT}/integration.tar"]
}

target "integration2" {
	inherits = ["_common", "test"]
	target   = "tester2out"
	args = {
		ARGS = "-test.run=Integration"
		NAME = "integration"
	}
	output = ["type=local,dest=${TEST_IMAGES_OUTPUT}/int"]

}

target "integration3" {
	args = {
		HTTP_PROXY  = HTTP_PROXY
		HTTPS_PROXY = HTTPS_PROXY
		NO_PROXY    = NO_PROXY
		ARGS        = "-test.run=Integration"
		NAME        = "integration"
	}
	target = "tester3out"

	output = ["type=local,dest=${TEST_IMAGES_OUTPUT}/int"]
}

target "e2e" {
	inherits = ["_common", "test"]
	target   = "tester-with-build"
	args = {
		ARGS = "-test.run=E2E"
		NAME = "e2e"
		E2E  = 1
	}
	output = ["type=docker,name=e2e,dest=${TEST_IMAGES_OUTPUT}/e2e.tar"]
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
