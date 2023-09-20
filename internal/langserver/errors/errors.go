// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errors

func EnrichTfExecError(err error) error {
	// if module.IsTerraformNotFound(err) {
	// 	return e.New("Terraform (CLI) is required. " +
	// 		"Please install Terraform or make it available in $PATH")
	// }
	return err
}
