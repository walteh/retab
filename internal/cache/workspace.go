package cache

import (
	"errors"

	"github.com/walteh/retab/internal/source"
)

// fileExists reports whether the file has a Content (which may be empty).
// An overlay exists even if it is not reflected in the file system.
func fileExists(fh source.FileHandle) bool {
	_, err := fh.Content()
	return err == nil
}

// errExhausted is returned by findModules if the file scan limit is reached.
var errExhausted = errors.New("exhausted")

// Limit go.mod search to 1 million files. As a point of reference,
// Kubernetes has 22K files (as of 2020-11-24).
//
// Note: per golang/go#56496, the previous limit of 1M files was too slow, at
// which point this limit was decreased to 100K.
const fileLimit = 100_000
