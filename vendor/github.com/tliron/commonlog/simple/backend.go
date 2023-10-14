package simple

import (
	"io"
	"os"

	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

const LOG_FILE_WRITE_PERMISSIONS = 0600

const DEFAULT_BUFFER_SIZE = 1_000

func init() {
	backend := NewBackend()
	backend.Configure(0, nil)
	commonlog.SetBackend(backend)
}

//
// Backend
//

type Backend struct {
	Writer     io.Writer
	Format     FormatFunc
	BufferSize int
	Buffered   bool

	colorize      bool
	nameHierarchy *commonlog.NameHierarchy
}

func NewBackend() *Backend {
	return &Backend{
		Format:        DefaultFormat,
		BufferSize:    DEFAULT_BUFFER_SIZE,
		Buffered:      true,
		nameHierarchy: commonlog.NewNameHierarchy(),
	}
}

// ([commonlog.Backend] interface)
func (self *Backend) Configure(verbosity int, path *string) {
	maxLevel := commonlog.VerbosityToMaxLevel(verbosity)

	if maxLevel == commonlog.None {
		self.Writer = io.Discard
		self.nameHierarchy.SetMaxLevel(commonlog.None)
	} else {
		if path != nil {
			if file, err := os.OpenFile(*path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, LOG_FILE_WRITE_PERMISSIONS); err == nil {
				util.OnExitError(file.Close)
				if self.Buffered {
					writer := util.NewBufferedWriter(file, self.BufferSize)
					util.OnExitError(writer.Close)
					self.Writer = writer
				} else {
					self.Writer = util.NewSyncedWriter(file)
				}
			} else {
				util.Failf("log file error: %s", err.Error())
			}
		} else {
			self.colorize = terminal.Colorize
			if self.Buffered {
				writer := util.NewBufferedWriter(os.Stderr, self.BufferSize)
				util.OnExitError(writer.Close)
				self.Writer = writer
			} else {
				self.Writer = util.NewSyncedWriter(os.Stderr)
			}
		}

		self.nameHierarchy.SetMaxLevel(maxLevel)
	}
}

// ([commonlog.Backend] interface)
func (self *Backend) GetWriter() io.Writer {
	return self.Writer
}

// ([commonlog.Backend] interface)
func (self *Backend) NewMessage(level commonlog.Level, depth int, name ...string) commonlog.Message {
	if self.AllowLevel(level, name...) {
		return commonlog.NewUnstructuredMessage(func(message string) {
			message = self.Format(message, name, level, self.colorize)
			io.WriteString(self.Writer, message+"\n")
		})
	} else {
		return nil
	}
}

// ([commonlog.Backend] interface)
func (self *Backend) AllowLevel(level commonlog.Level, name ...string) bool {
	return self.nameHierarchy.AllowLevel(level, name...)
}

// ([commonlog.Backend] interface)
func (self *Backend) SetMaxLevel(level commonlog.Level, name ...string) {
	self.nameHierarchy.SetMaxLevel(level, name...)
}

// ([commonlog.Backend] interface)
func (self *Backend) GetMaxLevel(name ...string) commonlog.Level {
	return self.nameHierarchy.GetMaxLevel(name...)
}
