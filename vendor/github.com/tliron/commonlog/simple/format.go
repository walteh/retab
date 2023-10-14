package simple

import (
	"io"
	"strings"
	"time"

	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/terminal"
)

const TIME_FORMAT = "2006/01/02 15:04:05.000"

type FormatFunc func(message string, id []string, level commonlog.Level, colorize bool) string

// ([FormatFunc] signature)
func DefaultFormat(message string, id []string, level commonlog.Level, colorize bool) string {
	var builder strings.Builder

	FormatTime(&builder)
	FormatLevel(&builder, level, true)
	builder.WriteRune(' ')
	FormatID(&builder, id)
	builder.WriteRune(' ')
	builder.WriteString(message)

	s := builder.String()

	if colorize {
		s = FormatColorize(s, level)
	}

	return s
}

func FormatTime(writer io.StringWriter) {
	writer.WriteString(time.Now().Format(TIME_FORMAT))
}

func FormatID(writer io.StringWriter, id []string) {
	writer.WriteString("[")
	length := len(id)
	switch length {
	case 0:
	case 1:
		writer.WriteString(id[0])
	default:
		last := length - 1
		for _, i := range id[:last] {
			writer.WriteString(i)
			writer.WriteString(".")
		}
		writer.WriteString(id[last])
	}
	writer.WriteString("]")
}

func FormatLevel(writer io.StringWriter, level commonlog.Level, align bool) {
	if align {
		switch level {
		case commonlog.Critical:
			writer.WriteString("  CRIT")
		case commonlog.Error:
			writer.WriteString(" ERROR")
		case commonlog.Warning:
			writer.WriteString("  WARN")
		case commonlog.Notice:
			writer.WriteString("  NOTE")
		case commonlog.Info:
			writer.WriteString("  INFO")
		case commonlog.Debug:
			writer.WriteString(" DEBUG")
		}
	} else {
		switch level {
		case commonlog.Critical:
			writer.WriteString("CRIT")
		case commonlog.Error:
			writer.WriteString("ERROR")
		case commonlog.Warning:
			writer.WriteString("WARN")
		case commonlog.Notice:
			writer.WriteString("NOTE")
		case commonlog.Info:
			writer.WriteString("INFO")
		case commonlog.Debug:
			writer.WriteString("DEBUG")
		}
	}
}

func FormatColorize(s string, level commonlog.Level) string {
	switch level {
	case commonlog.Critical:
		return terminal.ColorRed(s)
	case commonlog.Error:
		return terminal.ColorRed(s)
	case commonlog.Warning:
		return terminal.ColorYellow(s)
	case commonlog.Notice:
		return terminal.ColorCyan(s)
	case commonlog.Info:
		return terminal.ColorBlue(s)
	default:
		return s
	}
}
