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
		BUILDX_EXPERIMENTAL           = 1 // enables experimental cmds/flags for docs generation
		FORMATS                       = DOCS_FORMATS
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

target "validate" {
	inherits = ["_common"]
	matrix = {
		item = [
			{
				name = "lint",
				dest = ""
			},
			{
				name = "vendor",
				dest = VENDOR_OUTPUT
			},
			{
				name = "docs",
				dest = DOCS_OUTPUT
			},
			{
				name = "mockery",
				dest = MOCKERY_OUTPUT
			},
			{
				name = "buf",
				dest = BUF_OUTPUT
			},
		]
	}
	name = "validate-${item.name}"
	args = {
		NAME    = item.name
		DESTDIR = item.dest
	}
	output     = ["type=cacheonly"]
	target     = "validate"
	dockerfile = "./hack/dockerfiles/${item.name}.Dockerfile"
}


##################################################################
# GENERATE
##################################################################

target "generate" {
	inherits = ["_common"]
	matrix = {
		item = [
			{
				name = "vendor",
				dest = VENDOR_OUTPUT
			},
			{
				name = "docs",
				dest = DOCS_OUTPUT
			},
			{
				name = "mockery",
				dest = MOCKERY_OUTPUT
			},
			{
				name = "buf",
				dest = BUF_OUTPUT
			},
		]
	}
	name = "generate-${item.name}"
	args = {
		NAME    = item.name
		DESTDIR = item.dest
	}
	output     = ["type=local,dest=${item.dest}"]
	target     = "generate"
	dockerfile = "./hack/dockerfiles/${item.name}.Dockerfile"
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

target "testable" {
	inherits = ["_common"]
	target   = "testable"
	output   = ["type=local,dest=${DESTDIR}/testable"]
}

target "case" {
	inherits = ["_common"]
	target   = "case"
	matrix = {
		item = [
			{
				name = "unit"
				args = "-test.skip=Integration -test.skip=E2E"
			},
			{
				name = "integration"
				args = "-test.run=Integration"
			},
			{
				name = "e2e"
				args = "-test.run=E2E"
			}
		]
	}
	name = "case-${item.name}"
	args = {
		ARGS = item.args
		NAME = item.name
		E2E  = item.name == "e2e" ? 1 : 0
	}
	contexts = {
		build    = BUILD_OUTPUT
		testable = TEST_OUTPUT
	}
	output = ["type=local,dest=${DESTDIR}/case-${item.name}"]
}

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
