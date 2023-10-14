package util

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
