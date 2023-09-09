variable "DOCKER_IMAGE" {
	default = "walteh/retab"
}

variable "BIN_NAME" {
	default = "retab"
}

variable "ROOT_DIR" {
	default = "."
}

variable "DEST_DIR" {
	default = "${ROOT_DIR}/bin"
}

variable "GEN_DIR" {
	default = "${ROOT_DIR}/gen"
}


##################################################################
# LOCALS
##################################################################

variable "HTTP_PROXY" {
	default = ""
}
variable "HTTPS_PROXY" {
	default = ""
}
variable "NO_PROXY" {
	default = ""
}

##################################################################
# GITHUB ACTIONS
##################################################################


variable "GITHUB_REPOSITORY" {
	default = ""
}

variable "GITHUB_RUN_ID" {
	default = ""
}

variable "GITHUB_SHA" {
	default = ""
}

variable "GITHUB_REF" {
	default = ""
}

variable "IS_GITHUB_ACTIONS" {
	default = GITHUB_REPOSITORY != "" ? 1 : 0
}

##################################################################
# COMMON
##################################################################


target "_common" {
	args = {
		GO_VERSION                    = "1.21.0"
		BUILDRC_VERSION               = "0.12.9"
		XX_VERSION                    = "1.2.1"
		GOTESTSUM_VERSION             = "v1.10.1"
		GOLANGCI_LINT_VERSION         = "v1.54.2"
		BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
		BIN_NAME                      = BIN_NAME
		BUILDX_EXPERIMENTAL           = 1
		DOCS_FORMATS                  = "md"
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
		BUILDKIT_MULTI_PLATFORM = 1
	}
}

target "_tagged" {
	tags = flatten([
		IS_GITHUB_ACTIONS == 1 ? ["latest"] : [],
		[]
	])
	labels = merge({}, IS_GITHUB_ACTIONS == 1 ? {
		"org.opencontainers.image.source" = "https://github.com/${GITHUB_REPOSITORY}"
	} : {})
}

target "_attest" {
	attest = IS_GITHUB_ACTIONS == 1 ? [
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

##################################################################
# GENERATE
##################################################################

target "generate" {
	inherits = ["_common"]
	matrix = {
		item = [
			{
				name = "vendor",
				dest = "${ROOT_DIR}"
			},
			{
				name = "docs",
				dest = "${ROOT_DIR}/docs/reference"
			},
			{
				name = "mockery",
				dest = "${GEN_DIR}/mockery"
			},
			{
				name = "buf",
				dest = "${GEN_DIR}/buf"
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
# VALIDATE
##################################################################

target "validate" {
	matrix = {
		item = [
			{
				name     = "lint"
				inherits = []
			},
			{
				name     = "vendor",
				inherits = ["generate-vendor"],
			},
			{
				name     = "docs",
				inherits = ["generate-docs"],
			},
			{
				name     = "mockery",
				inherits = ["generate-mockery"],
			},
			{
				name     = "buf",
				inherits = ["generate-buf"],
			},
		]
	}
	inherits   = flatten([["_common"], item.inherits])
	name       = "validate-${item.name}"
	output     = ["type=cacheonly"]
	target     = "validate"
	dockerfile = "./hack/dockerfiles/${item.name}.Dockerfile"
}


##################################################################
# METADATA
##################################################################

target "meta" {
	inherits = ["_common", "_cross"]
	target   = "meta"
	output   = ["type=local,dest=${DEST_DIR}/meta"]
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
	output    = ["type=local,dest=${DEST_DIR}/build"]
	platforms = ["local"]
}

target "build" {
	inherits = ["_common", "_cross", "_attest"]
	target   = "build"
	output   = ["type=local,dest=${DEST_DIR}/build"]
}

target "package" {
	inherits  = ["_common", ]
	target    = "package"
	output    = ["type=local,dest=${DEST_DIR}/package"]
	platforms = ["local"]
	contexts = {
		build = "${DEST_DIR}/build"
	}
}

##################################################################
# TESTING
##################################################################

target "testable" {
	inherits = ["_common"]
	target   = "testable"
	output   = ["type=local,dest=${DEST_DIR}/testable"]
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
	output = ["type=local,dest=${DEST_DIR}/case-${item.name}"]
}

target "test" {
	matrix = {
		item = [
			{
				name = "unit"
			},
			{
				name = "integration"
			},
			{
				name = "e2e"
			}
		]
	}
	name     = "test-${item.name}"
	inherits = ["_common", "case-${item.name}"]
	target   = "test"
	output   = ["type=local,dest=${DEST_DIR}/test"]
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
