##################################################################
# GLOBALS
##################################################################

DOCKER_ORG     = "walteh"
DOCKER_REPO    = "retab"
IS_DEFAULT_BIN = BIN_NAME == DOCKER_REPO ? true : false
IS_LOCAL       = CI != "1" && CI != "true" ? true : false
ROOT_DIR       = "."
DEST_DIR       = "${ROOT_DIR}/bin"
GEN_DIR        = "${ROOT_DIR}/gen"

##################################################################
# INPUTS
##################################################################

variable "BIN_NAME" { default = DOCKER_REPO }
variable "HTTP_PROXY" {}
variable "HTTPS_PROXY" {}
variable "NO_PROXY" {}
variable "CI" {}
variable "VERSION_TAG" {}

##################################################################
# GITHUB ACTIONS
##################################################################

variable "GITHUB_REPOSITORY" {}
variable "GITHUB_RUN_ID" {}
variable "GITHUB_SHA" {}
variable "GITHUB_REF" {}
variable "GITHUB_JOB" {}
variable "GITHUB_ACTOR" {}
variable "GITHUB_JOB_NAME" {}
variable "GITHUB_ACTIONS" {}

IS_GITHUB_ACTIONS = GITHUB_ACTIONS == "true" ? true : false

GITHUB_ACTIONS_TAGS = flatten([
	GITHUB_REF == "refs/heads/main" ? ["latest", "main"] : [],
])

target _github_actions {
	cache-to   = ["type=gha,scope=${GITHUB_JOB_NAME != "" ? GITHUB_JOB_NAME : GITHUB_JOB}"]
	cache-from = ["type=gha,mode=max,scope=${GITHUB_JOB_NAME != "" ? GITHUB_JOB_NAME : GITHUB_JOB}"]
	labels = {
		"org.opencontainers.image.url"           = "https://github.com/${GITHUB_REPOSITORY}"
		"org.opencontainers.image.documentation" = "https://github.com/${GITHUB_REPOSITORY}/README.md"
		"org.opencontainers.image.source"        = "https://github.com/${GITHUB_REPOSITORY}"
		"org.opencontainers.image.revision"      = "${GITHUB_SHA}"
		"org.opencontainers.image.authors"       = "${GITHUB_ACTOR}"
	}
}

##################################################################
# TAGS
##################################################################

DOCKER_IMAGE_ROOTS = [for x in [
	!IS_LOCAL ? "${DOCKER_ORG}/${DOCKER_REPO}" : "local/${DOCKER_REPO}",
	IS_GITHUB_ACTIONS ? "ghcr.io/${GITHUB_REPOSITORY}" : null,
] : x if x != null]

DOCKER_IMAGE_TAGS = [for tag in flatten([
	# local tags for local builds
	IS_LOCAL ? ["local"] : [],
	# tags for main branch
	IS_GITHUB_ACTIONS ? GITHUB_ACTIONS_TAGS : [],
	# if version tag exists, use it
	VERSION_TAG != "" ? [VERSION_TAG] : [],
]) : IS_DEFAULT_BIN ? tag : "${tag}-${BIN_NAME}"]

target "_tagged" {
	tags = flatten([
		for image in DOCKER_IMAGE_ROOTS : [for tag in DOCKER_IMAGE_TAGS : "${image}:${tag}"]
	])
}

##################################################################
# COMMON
##################################################################

target "_common" {
	push = false
	inherits = flatten([
		IS_GITHUB_ACTIONS ? ["_github_actions"] : []
	])
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
	labels = {
		"org.opencontainers.image.title"       = "${BIN_NAME}"
		"org.opencontainers.image.description" = "A tool to reformat spaced out text into tabbed text"
		"org.opencontainers.image.created"     = timestamp()
		"org.opencontainers.image.vendor"      = "${DOCKER_ORG}"
		"org.opencontainers.image.version"     = VERSION_TAG != "" ? VERSION_TAG : null
	}
}


target "_cross" {
	platforms = [
		"darwin/amd64",
		"darwin/arm64",
		"linux/amd64",
		/* "linux/arm/v6",
		"linux/arm/v7", */
		"linux/arm64",
		/* "linux/ppc64le",
		"linux/riscv64",
		"linux/s390x", */
		/* "freebsd/amd64",
		"freebsd/arm64",
		"openbsd/amd64",
		"openbsd/arm64",
		"netbsd/amd64",
		"netbsd/arm64", */
		"windows/amd64",
		"windows/arm64",
	]
	args = {
		BUILDKIT_MULTI_PLATFORM = 1
	}
}

target "_attest" {
	attest = IS_GITHUB_ACTIONS ? [
		"type=provenance,mode=max,builder-id=https://github.com/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}",
		"type=sbom"
		] : [
		"type=provenance,mode=max",
		"type=sbom"
	]
}

group "default" {
	targets = ["image"]
}

##################################################################
# COMMANDS
##################################################################

COMMANDS = {
	lint = {
		dockerfile = "./hack/dockerfiles/lint.Dockerfile"
		validate   = { target = "validate" }
		generate   = null
		dest       = "${ROOT_DIR}"
	}
	vendor = {
		dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
		validate   = { target = "validate" }
		generate   = { target = "generate" }
		dest       = "${ROOT_DIR}"
	}
	docs = {
		dockerfile = "./hack/dockerfiles/docs.Dockerfile"
		validate   = { target = "validate" }
		generate   = { target = "generate" }
		dest       = "${ROOT_DIR}/docs/reference"
	}
	mockery = {
		dockerfile = "./hack/dockerfiles/mockery.Dockerfile"
		validate   = { target = "validate" }
		generate   = { target = "generate" }
		dest       = "${GEN_DIR}/mockery"
	}
	buf = {
		dockerfile = "./hack/dockerfiles/buf.Dockerfile"
		validate   = { target = "validate" }
		generate   = { target = "generate" }
		dest       = "${GEN_DIR}/buf"
	}
}

##################################################################
# GENERATE
##################################################################

target "generate" {
	inherits = ["_common"]
	matrix = {
		item = [for name, item in COMMANDS : merge(item, { name = name }) if item.generate != null]
	}
	name = "generate-${item.name}"
	args = {
		NAME = item.name
	}
	output     = ["type=local,dest=${item.dest}"]
	target     = item.generate.target
	dockerfile = item.dockerfile
}

##################################################################
# VALIDATE
##################################################################

target "validate" {
	inherits = ["_common"]
	matrix = {
		item = [for name, item in COMMANDS : merge(item, { name = name }) if item.validate != null]
	}
	name       = "validate-${item.name}"
	output     = ["type=cacheonly"]
	target     = item.validate.target
	dockerfile = item.dockerfile
	args = {
		NAME    = item.name
		DESTDIR = item.dest
	}
}


##################################################################
# METADATA
##################################################################

target "meta" {
	inherits = ["_common", "_cross"]
	target   = "meta"
	output   = ["type=local,dest=${DEST_DIR}/meta"]
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

target "test-build" {
	inherits = ["_common"]
	target   = "test-build"
	output   = ["type=local,dest=${DEST_DIR}/test-build"]
}

target "test" {
	inherits = ["_common"]
	target   = "test"
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
			},
			{
				name = "all"
				args = ""
			}
		]
	}
	name = "test-${item.name}"
	args = {
		ARGS = item.args
		NAME = item.name
		E2E  = item.name == "e2e" || item.name == "all" ? 1 : 0
	}
	output = ["type=docker,dest=${DEST_DIR}/test-${item.name}.tar,name=test-${item.name}"]
}

##################################################################
# IMAGE
##################################################################

target "image" {
	inherits  = ["_attest", "_common", "_tagged"]
	target    = "entry"
	output    = ["type=image"]
	platforms = ["local"]
}

target "registry" {
	inherits = ["_cross", "_attest", "_common", "_tagged"]
	target   = "entry"
	output   = ["type=image"]
}
