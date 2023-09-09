variable "BIN_NAME" { default = "retab" }
variable "ROOT_DIR" { default = "." }
variable "DEST_DIR" { default = "${ROOT_DIR}/bin" }
variable "GEN_DIR" { default = "${ROOT_DIR}/gen" }

##################################################################
# LOCALS
##################################################################

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

IS_DOCKER_DEFAULT_BIN = BIN_NAME == DOCKER_REPO ? true : false

DOCKER_IMAGE_TAGS = [for tag in flatten([
	# local tags for local builds
	IS_LOCAL ? ["local"] : [],
	# tags for main branch
	IS_GITHUB_ACTIONS ? GITHUB_ACTIONS_TAGS : [],
	# if version tag exists, use it
	VERSION_TAG != "" ? [VERSION_TAG] : [],
]) : IS_DOCKER_DEFAULT_BIN ? tag : "${BIN_NAME}-${tag}"]

target "_tagged" {
	tags = flatten([
		for image in DOCKER_IMAGE_ROOTS : [for tag in DOCKER_IMAGE_TAGS : "${image}:${tag}"]
	])
}

##################################################################
# COMMON
##################################################################

DOCKER_ORG = "walteh"

DOCKER_REPO = "retab"

IS_LOCAL = CI != "1" && CI != "true" ? true : false

target "_common" {
	push     = false
	inherits = flatten([IS_GITHUB_ACTIONS ? ["_github_actions"] : []])
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
		"org.opencontainers.image.version"     = "local"
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
	inherits  = ["build"]
	target    = "entry"
	output    = ["type=image"]
	platforms = ["local"]
}

target "registry" {
	inherits = ["_cross", "_attest", "_common", "_tagged"]
	target   = "entry"
	output   = ["type=image"]
}
