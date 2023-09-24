// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"fmt"
)

type AlreadyExistsError struct {
	Idx string
}

func (e *AlreadyExistsError) Error() string {
	if e.Idx != "" {
		return fmt.Sprintf("%s already exists", e.Idx)
	}
	return "already exists"
}

type NoSchemaError struct{}

func (e *NoSchemaError) Error() string {
	return "no schema found"
}
