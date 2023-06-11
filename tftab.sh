#!/usr/bin/env bash

file="$1"

function logic() {
    local temp_file
    temp_file=$(mktemp)
    "$1" fmt "$file" && sed -e'':a'' -e's/^\\(\\t*\\)  /\\1\\t/;ta' "$file" >"$temp_file" && mv "$temp_file" "$file"
    exit 0
}

if command -v tfenv >/dev/null 2>&1; then
    logic "tfenv exec terraform"
elif command -v terraform >/dev/null 2>&1; then
    logic "terraform"
else
    echo "terraform not found"
    exit 1
fi
