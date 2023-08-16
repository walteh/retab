variable "GO_VERSION" {
    default = "1.20.7"
}
variable "DOCS_FORMATS" {
    default = "md"
}

variable "DEST_DIR" {
    default = "./bin"
}

variable "GEN_DIR" {
    default = "./gen"
}

variable "GO_PACKAGE" {
    default = "github.com/walteh/tftab"
}

variable "DOCKER_IMAGE" {
    default = "walteh/tftab"
}

# Special target: https://github.com/docker/metadata-action#bake-definition
target "meta-helper" {
    tags = ["${DOCKER_IMAGE}:local"]
}

target "_common" {
    args = {
        GO_VERSION                    = GO_VERSION
        BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
        DOCKER_IMAGE                  = DOCKER_IMAGE
        GO_PACKAGE                    = GO_PACKAGE
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

target "validate-authors" {
    inherits   = ["_common"]
    dockerfile = "./hack/dockerfiles/authors.Dockerfile"
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

target "update-authors" {
    inherits   = ["_common"]
    dockerfile = "./hack/dockerfiles/authors.Dockerfile"
    target     = "update"
    output     = ["."]
}

target "update-gen" {
    inherits   = ["_common"]
    dockerfile = "./hack/dockerfiles/gen.Dockerfile"
    target     = "update"
    output     = ["${GEN_DIR}"]
}

target "mod-outdated" {
    inherits        = ["_common"]
    dockerfile      = "./hack/dockerfiles/vendor.Dockerfile"
    target          = "outdated"
    no-cache-filter = ["outdated"]
    output          = ["type=cacheonly"]
}

target "test" {
    inherits = ["_common"]
    target   = "test-coverage"
    output   = ["${DEST_DIR}/coverage"]
}

target "binaries" {
    inherits  = ["_common"]
    target    = "binaries"
    output    = ["${DEST_DIR}/build"]
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
    output   = ["${DEST_DIR}/release"]
}

target "image" {
    inherits = ["meta-helper", "binaries"]
    output   = ["type=image"]
}

target "image-cross" {
    inherits = ["meta-helper", "binaries-cross"]
    output   = ["type=image"]
}

target "image-local" {
    inherits = ["image"]
    output   = ["type=docker"]
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

target "integration-test-base" {
    inherits = ["_common"]
    args = {
        HTTP_PROXY  = HTTP_PROXY
        HTTPS_PROXY = HTTPS_PROXY
        NO_PROXY    = NO_PROXY
    }
    target = "integration-test-base"
    output = ["type=cacheonly"]
}

target "integration-test" {
    inherits = ["integration-test-base"]
    target   = "integration-test"
}
