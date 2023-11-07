//go:build wasm

package terminal

func EnableColor(force bool) (Cleanup, error) {
	if force {
		Colorize = true
	}

	DefaultStylist = NewStylist(Colorize)
	return nil, nil
}
