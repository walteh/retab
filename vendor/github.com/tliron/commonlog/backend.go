package commonlog

import (
	"io"
)

//
// Backend
//

type Backend interface {
	// If "path" is nil will log to stdout, colorized if possible
	// The default "verbosity" 0 will log criticals, errors, warnings, and notices.
	// "verbosity" 1 will add infos. "verbosity" 2 will add debugs.
	// Set "verbostiy" to -1 to disable the log.
	Configure(verbosity int, path *string)

	// Gets the backend's [io.Writer]. Can be nil if unsupported.
	GetWriter() io.Writer

	NewMessage(level Level, depth int, name ...string) Message

	// Returns true if a level is loggable for the given name.
	AllowLevel(level Level, name ...string) bool

	// Sets the maximum loggable level for the given name. Will become the
	// default maximum level for names deeper in the hierarchy unless
	// explicitly set for them.
	SetMaxLevel(level Level, name ...string)

	// Gets the maximum loggable level for the given name.
	GetMaxLevel(name ...string) Level
}
