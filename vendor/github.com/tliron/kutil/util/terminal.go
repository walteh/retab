package util

import (
	"github.com/tliron/kutil/terminal"
)

func InitializeColorization(colorize string) {
	cleanup, err := terminal.ProcessColorizeFlag(colorize)
	FailOnError(err)
	if cleanup != nil {
		OnExitError(cleanup)
	}
}
