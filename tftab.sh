#!/usr/bin/env bash

file="$1"
temp_file=$file-notab
terraform fmt "$file" && sed -e'':a'' -e's/^\\(\\t*\\)  /\\1\\t/;ta' "$file" >"$temp_file" && mv "$temp_file" "$file"
