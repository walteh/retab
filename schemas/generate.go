//go:build ignore
// +build ignore

package main

import (
	"context"

	"github.com/walteh/retab/schemas"
)

func main() {
	ctx := context.Background()

	err := schemas.DownloadAllJSONSchemas(ctx)
	if err != nil {
		panic(err)
	}
}
