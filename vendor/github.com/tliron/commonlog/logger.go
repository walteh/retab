package commonlog

import (
	"fmt"
)

//
// Logger
//

// While [NewMessage] is the "true" API entry point, this interface enables
// a familiar logger API. Because it's an interface, references can easily
// replace the implementation, for example setting a reference to
// [MOCK_LOGGER] will disable the logger.
//
// See [GetLogger].
type Logger interface {
	// Returns true if a level is loggable for this logger.
	AllowLevel(level Level) bool

	// Sets the maximum loggable level for this logger.
	SetMaxLevel(level Level)

	// Gets the maximum loggable level for this logger.
	GetMaxLevel() Level

	// Creates a new message for this logger. Will return nil if
	// the level is not loggable.
	//
	// The depth argument is used for skipping frames in callstack
	// logging, if supported.
	NewMessage(level Level, depth int, keysAndValues ...any) Message

	// Convenience method to create and send a message with at least
	// the "message" key. Additional keys can be set by providing
	// a sequence of key-value pairs.
	Log(level Level, depth int, message string, keysAndValues ...any)

	// Convenience method to create and send a message with just
	// the "message" key, where the message is created via the format
	// and args similarly to fmt.Printf.
	Logf(level Level, depth int, format string, args ...any)

	Critical(message string, keysAndValues ...any)
	Criticalf(format string, args ...any)
	Error(message string, keysAndValues ...any)
	Errorf(format string, args ...any)
	Warning(message string, keysAndValues ...any)
	Warningf(format string, args ...any)
	Notice(message string, keysAndValues ...any)
	Noticef(format string, args ...any)
	Info(message string, keysAndValues ...any)
	Infof(format string, args ...any)
	Debug(message string, keysAndValues ...any)
	Debugf(format string, args ...any)
}

//
// BackendLogger
//

// Default [Logger] implementation that logs to the current backend set with
// [SetBackend].
type BackendLogger struct {
	name []string
}

func NewBackendLogger(name ...string) BackendLogger {
	return BackendLogger{name: name}
}

// ([Logger] interface)
func (self BackendLogger) AllowLevel(level Level) bool {
	return AllowLevel(level, self.name...)
}

// ([Logger] interface)
func (self BackendLogger) SetMaxLevel(level Level) {
	SetMaxLevel(level, self.name...)
}

// ([Logger] interface)
func (self BackendLogger) GetMaxLevel() Level {
	return GetMaxLevel(self.name...)
}

// ([Logger] interface)
func (self BackendLogger) NewMessage(level Level, depth int, keysAndValues ...any) Message {
	if message := NewMessage(level, depth, self.name...); message != nil {
		SetMessageKeysAndValue(message, keysAndValues...)
		return message
	} else {
		return nil
	}
}

// ([Logger] interface)
func (self BackendLogger) Log(level Level, depth int, message string, keysAndValues ...any) {
	if message_ := self.NewMessage(level, depth+1, keysAndValues...); message_ != nil {
		message_.Set("message", message)
		message_.Send()
	}
}

// ([Logger] interface)
func (self BackendLogger) Logf(level Level, depth int, format string, args ...any) {
	if message := self.NewMessage(level, depth+1); message != nil {
		message.Set("message", fmt.Sprintf(format, args...))
		message.Send()
	}
}

// ([Logger] interface)
func (self BackendLogger) Critical(message string, keysAndValues ...any) {
	self.Log(Critical, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self BackendLogger) Criticalf(format string, args ...any) {
	self.Logf(Critical, 1, format, args...)
}

// ([Logger] interface)
func (self BackendLogger) Error(message string, keysAndValues ...any) {
	self.Log(Error, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self BackendLogger) Errorf(format string, args ...any) {
	self.Logf(Error, 1, format, args...)
}

// ([Logger] interface)
func (self BackendLogger) Warning(message string, keysAndValues ...any) {
	self.Log(Warning, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self BackendLogger) Warningf(format string, args ...any) {
	self.Logf(Warning, 1, format, args...)
}

// ([Logger] interface)
func (self BackendLogger) Notice(message string, keysAndValues ...any) {
	self.Log(Notice, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self BackendLogger) Noticef(format string, args ...any) {
	self.Logf(Notice, 1, format, args...)
}

// ([Logger] interface)
func (self BackendLogger) Info(message string, keysAndValues ...any) {
	self.Log(Info, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self BackendLogger) Infof(format string, args ...any) {
	self.Logf(Info, 1, format, args...)
}

// ([Logger] interface)
func (self BackendLogger) Debug(message string, keysAndValues ...any) {
	self.Log(Debug, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self BackendLogger) Debugf(format string, args ...any) {
	self.Logf(Debug, 1, format, args...)
}

//
// ScopeLogger
//

// Wrapping [Logger] that calls [Message.Set] with a "scope" key
// on all messages. There is special support for nesting scope loggers
// such that a nested scope string is appended to the wrapped scope with
// a "." notation.
type ScopeLogger struct {
	logger Logger
	scope  string
}

func NewScopeLogger(logger Logger, scope string) ScopeLogger {
	if subLogger, ok := logger.(ScopeLogger); ok {
		scope = subLogger.scope + "." + scope
		logger = subLogger.logger
	}

	return ScopeLogger{
		logger: logger,
		scope:  scope,
	}
}

// ([Logger] interface)
func (self ScopeLogger) AllowLevel(level Level) bool {
	return self.logger.AllowLevel(level)
}

// ([Logger] interface)
func (self ScopeLogger) SetMaxLevel(level Level) {
	self.logger.SetMaxLevel(level)
}

// ([Logger] interface)
func (self ScopeLogger) GetMaxLevel() Level {
	return self.logger.GetMaxLevel()
}

// ([Logger] interface)
func (self ScopeLogger) NewMessage(level Level, depth int, keysAndValues ...any) Message {
	if message := self.logger.NewMessage(level, depth, keysAndValues...); message != nil {
		message.Set("scope", self.scope)
		return message
	} else {
		return nil
	}
}

// ([Logger] interface)
func (self ScopeLogger) Log(level Level, depth int, message string, keysAndValues ...any) {
	if message_ := self.NewMessage(level, depth+1, keysAndValues...); message_ != nil {
		message_.Set("message", message)
		message_.Send()
	}
}

// ([Logger] interface)
func (self ScopeLogger) Logf(level Level, depth int, format string, args ...any) {
	if message := self.NewMessage(level, depth+1); message != nil {
		message.Set("message", fmt.Sprintf(format, args...))
		message.Send()
	}
}

// ([Logger] interface)
func (self ScopeLogger) Critical(message string, keysAndValues ...any) {
	self.Log(Critical, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self ScopeLogger) Criticalf(format string, args ...any) {
	self.Logf(Critical, 1, format, args...)
}

// ([Logger] interface)
func (self ScopeLogger) Error(message string, keysAndValues ...any) {
	self.Log(Error, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self ScopeLogger) Errorf(format string, args ...any) {
	self.Logf(Error, 1, format, args...)
}

// ([Logger] interface)
func (self ScopeLogger) Warning(message string, keysAndValues ...any) {
	self.Log(Warning, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self ScopeLogger) Warningf(format string, args ...any) {
	self.Logf(Warning, 1, format, args...)
}

// ([Logger] interface)
func (self ScopeLogger) Notice(message string, keysAndValues ...any) {
	self.Log(Notice, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self ScopeLogger) Noticef(format string, args ...any) {
	self.Logf(Notice, 1, format, args...)
}

// ([Logger] interface)
func (self ScopeLogger) Info(message string, keysAndValues ...any) {
	self.Log(Info, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self ScopeLogger) Infof(format string, args ...any) {
	self.Logf(Info, 1, format, args...)
}

// ([Logger] interface)
func (self ScopeLogger) Debug(message string, keysAndValues ...any) {
	self.Log(Debug, 1, message, keysAndValues...)
}

// ([Logger] interface)
func (self ScopeLogger) Debugf(format string, args ...any) {
	self.Logf(Debug, 1, format, args...)
}

//
// MockLogger
//

var MOCK_LOGGER MockLogger

// [Logger] that does nothing.
type MockLogger struct{}

// ([Logger] interface)
func (self MockLogger) AllowLevel(level Level) bool {
	return false
}

// ([Logger] interface)
func (self MockLogger) SetMaxLevel(level Level) {
}

// ([Logger] interface)
func (self MockLogger) GetMaxLevel() Level {
	return None
}

// ([Logger] interface)
func (self MockLogger) NewMessage(level Level, depth int, keysAndValues ...any) Message {
	return nil
}

// ([Logger] interface)
func (self MockLogger) Log(level Level, depth int, message string, keysAndValues ...any) {
}

// ([Logger] interface)
func (self MockLogger) Logf(level Level, depth int, format string, args ...any) {
}

// ([Logger] interface)
func (self MockLogger) Critical(message string, keysAndValues ...any) {
}

// ([Logger] interface)
func (self MockLogger) Criticalf(format string, args ...any) {
}

// ([Logger] interface)
func (self MockLogger) Error(message string, keysAndValues ...any) {
}

// ([Logger] interface)
func (self MockLogger) Errorf(format string, args ...any) {
}

// ([Logger] interface)
func (self MockLogger) Warning(message string, keysAndValues ...any) {
}

// ([Logger] interface)
func (self MockLogger) Warningf(format string, args ...any) {
}

// ([Logger] interface)
func (self MockLogger) Notice(message string, keysAndValues ...any) {
}

// ([Logger] interface)
func (self MockLogger) Noticef(format string, args ...any) {
}

// ([Logger] interface)
func (self MockLogger) Info(message string, keysAndValues ...any) {
}

// ([Logger] interface)
func (self MockLogger) Infof(format string, args ...any) {
}

// ([Logger] interface)
func (self MockLogger) Debug(message string, keysAndValues ...any) {
}

// ([Logger] interface)
func (self MockLogger) Debugf(format string, args ...any) {
}
