//go:build !windows && !wasm

package terminal

import (
	"github.com/muesli/termenv"
)

func EnableColor(force bool) (Cleanup, error) {
	if force {
		Colorize = true
	} else {
		Colorize = termenv.EnvColorProfile() != termenv.Ascii
	}

	DefaultStylist = NewStylist(Colorize)
	return nil, nil
}
