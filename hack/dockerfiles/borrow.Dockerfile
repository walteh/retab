# syntax=docker/dockerfile:labs

ARG GO_VERSION
ARG MOCKERY_VERSION
ARG BUILDRC_VERSION

FROM vektra/mockery:v${MOCKERY_VERSION} AS mockery
FROM walteh/buildrc:${BUILDRC_VERSION} AS buildrc

FROM golang:${GO_VERSION}-alpine AS tools
COPY --from=buildrc /usr/bin/buildrc /usr/bin/buildrc
RUN apk add --no-cache git curl jq

# Set common working directory
WORKDIR /wrk

# Mockery stage
FROM tools as generator
ARG GOPLS_VERSION GO_MODULE DESTDIR REPOS DELIMITER
RUN --mount=type=bind,target=/wrk/repo,rw \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go/pkg/mod <<SHELL
	set -ex
# REPOS = jsonencode([
# 				{
# 					repo    = "github.com/hashicorp/terraform-ls"
# 					commit  = "94e47bd3a6371c6d56c2ab92d0d33b1ce84e9c0d"
# 					include = ["internal/**/*.go"]
# 					exclude = ["internal/hooks/*"]
# 					replacements = {
# 						"github.com/hashicorp/terraform-ls/internal/terraform" = "${GO_MODULE}/pkg/hclschema"
# 					}
# 				}
# 			])
	for item in $(echo "$REPOS" | jq -r '.[] | @base64'); do
		json_item=$(echo "$item" | base64 -d)
		repo=$(echo "$json_item" | jq -r '.repo')
		commit=$(echo "$json_item" | jq -r '.commit')

git clone https://${repo}.git working
(
	cd working
	git checkout "${commit}"
	mkdir -p "/out/${repo}"

	# Prepare include and exclude patterns
    include_patterns=$(echo "$json_item" | jq -r '.include | map("-o -path */" + . + " ") | join("")'  | cut -c 3-)
    exclude_patterns=$(echo "$json_item" | jq -r '.exclude | map("-o -path */" + . + " ") | join("")'  | cut -c 3-)





	# Use find to identify files and copy them
find . -type f \( $include_patterns \) ! \( $exclude_patterns \) -exec cp --parents '{}' /out/${repo} \;
)

		find "/out/${repo}" -type f -name "*.go" -exec sed -i "s|${repo}|${GO_MODULE}/${DESTDIR#./}/${repo}|g" {} \;

		replacements=$(echo "$json_item" | jq -r '.replacements | to_entries[] | @base64')
		for wrk in $replacements; do
			json_wrk=$(echo "$wrk" | base64 -d)
			key=$(echo "$json_wrk" | jq -r '.key')
			value=$(echo "$json_wrk" | jq -r '.value')
			find "/out/${repo}" -type f -name "*.go" -exec sed -i "s|${key}|${value}|g" {} \;
		done

	done

SHELL

# Final update stage
FROM scratch AS generate
COPY --from=generator /out /

FROM tools AS validate
ARG DESTDIR
RUN --mount=target=/context \
	--mount=from=generator,target=/out,source=/out,type=bind <<SHELL
	set -e
	cd /context
	buildrc diff --current="./${DESTDIR}" --correct="/out" --glob="*.go"
SHELL
