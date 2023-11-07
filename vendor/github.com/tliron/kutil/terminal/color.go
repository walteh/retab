package terminal

import (
	"fmt"
	"strconv"
)

var Colorize = false

type Cleanup func() error

const (
	escapePrefix = "\x1b["
	escapeSuffix = "m"

	ResetCode   = escapePrefix + "0" + escapeSuffix
	RedCode     = escapePrefix + "31" + escapeSuffix
	GreenCode   = escapePrefix + "32" + escapeSuffix
	YellowCode  = escapePrefix + "33" + escapeSuffix
	BlueCode    = escapePrefix + "34" + escapeSuffix
	MagentaCode = escapePrefix + "35" + escapeSuffix
	CyanCode    = escapePrefix + "36" + escapeSuffix
)

func ProcessColorizeFlag(colorize string) (Cleanup, error) {
	if colorize == "force" {
		return EnableColor(true)
	} else if colorize_, err := strconv.ParseBool(colorize); err == nil {
		if colorize_ {
			return EnableColor(false)
		}
	} else {
		return nil, fmt.Errorf("\"--colorize\" must be boolean or \"force\": %s", colorize)
	}
	return nil, nil
}

//
// Colorizer
//

type Colorizer func(name string) string

// Colorizer signature
func ColorRed(s string) string {
	return RedCode + s + ResetCode
}

// Colorizer signature
func ColorGreen(s string) string {
	return GreenCode + s + ResetCode
}

// Colorizer signature
func ColorYellow(s string) string {
	return YellowCode + s + ResetCode
}

// Colorizer signature
func ColorBlue(s string) string {
	return BlueCode + s + ResetCode
}

// Colorizer signature
func ColorMagenta(s string) string {
	return MagentaCode + s + ResetCode
}

// Colorizer signature
func ColorCyan(s string) string {
	return CyanCode + s + ResetCode
}
