#!/usr/bin/env bash

file="$1"
temp_file=$(mktemp --suffix=.tf) # Create a temp file with .tf extension
temp_output_file=$(mktemp)

# Copy the content of the original file to the temporary .tf file
cp "$file" "$temp_file"

# Run terraform fmt on the temporary .tf file, capturing output and return code
terraform_output=$(terraform fmt -write=false -list=false "$temp_file")
terraform_return_code=$?

# Check if terraform fmt was successful
if [ $terraform_return_code -eq 0 ]; then
	# If successful, process the output and overwrite the original file
	echo "$terraform_output" | sed -e'':a'' -e's/^\(\t*\)  /\1\t/;ta' >"$temp_output_file" && mv "$temp_output_file" "$file"
else
	# If there was an error, print a message and the terraform fmt output, and leave the original file unchanged
	echo "Error formatting Terraform file:"
	echo "$terraform_output"
	rm "$temp_file"
	rm "$temp_output_file"
	exit $terraform_return_code
fi

# Clean up temporary .tf file
rm "$temp_file"
