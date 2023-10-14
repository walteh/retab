//go:build windows

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

	if Colorize {
		DefaultStylist = NewStylist(true)
		if mode, err := termenv.EnableWindowsANSIConsole(); err == nil {
			return func() error {
				return termenv.RestoreWindowsConsole(mode)
			}, nil
		} else {
			return nil, err
		}
	} else {
		DefaultStylist = NewStylist(false)
		return nil, nil
	}
}
