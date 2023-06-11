#!/usr/bin/env bash

file="$1"
temp_file=$(mktemp)
terraform fmt -write=false -list=false "$file" | sed -e'':a'' -e's/^\(\t*\)  /\1\t/;ta' >"$temp_file" && mv "$temp_file" "$file"
