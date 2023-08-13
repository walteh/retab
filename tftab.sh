#!/usr/bin/env bash

file="$1"
temp_file=$(mktemp)

# Run terraform fmt, capturing output and return code
terraform_output=$(terraform fmt -write=false -list=false "$file")
terraform_return_code=$?

# Check if terraform fmt was successful
if [ $terraform_return_code -eq 0 ]; then
	# If successful, process the output and overwrite the file
	echo "$terraform_output" | sed -e'':a'' -e's/^\(\t*\)  /\1\t/;ta' >"$temp_file" && mv "$temp_file" "$file"
else
	# If there was an error, print a message and the terraform fmt output, and leave the original file unchanged
	echo "Error formatting Terraform file:"
	echo "$terraform_output"
	rm "$temp_file"
	exit $terraform_return_code
fi
