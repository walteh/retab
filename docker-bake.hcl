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

variable "GO_PKG" {
  default = "github.com/walteh/tftab"
}

variable "DOCKER_IMAGE" {
  default = "walteh/tftab"
}

variable "BIN_NAME" {
  default = "tftab"
}

target "_common" {
  args = {
    GO_VERSION                    = GO_VERSION
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
    DOCKER_IMAGE                  = DOCKER_IMAGE
    GO_PKG                        = GO_PKG
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
  output   = ["${DESTDIR}/coverage"]
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

target "bundle" {
  inherits = ["_common"]
  target   = "bundle"
  output   = ["${DESTDIR}/bundle"]
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


target "test" {
  inherits = ["_common"]
  target   = "test-starter"
  output   = ["type=cacheonly"]
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
