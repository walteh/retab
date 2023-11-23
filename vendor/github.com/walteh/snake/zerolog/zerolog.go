package snake

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/k0kubun/pp/v3"
	"github.com/rs/zerolog"
	"github.com/walteh/snake"
)

func Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

func NewJsonLogger() *zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	return &logger
}

func NewJsonStdErrLogger() *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
	return &logger
}

func NewVerboseLoggerContext(ctx context.Context) context.Context {
	verboseLogger := NewVerboseConsoleLogger()
	return verboseLogger.WithContext(ctx)
}

func NewJsonLoggerContext(ctx context.Context) context.Context {
	jsonLogger := NewJsonLogger()
	return jsonLogger.WithContext(ctx)
}

func NewJsonStdErrLoggerContext(ctx context.Context) context.Context {
	jsonLogger := NewJsonStdErrLogger()
	return jsonLogger.WithContext(ctx)
}

func NewVerboseConsoleLogger() *zerolog.Logger {

	consoleOutput := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro, NoColor: false}

	pretty := pp.New()

	pretty.SetColorScheme(pp.ColorScheme{})

	prettyerr := pp.New()
	prettyerr.SetExportedOnly(false)

	consoleOutput.FormatFieldValue = func(i interface{}) string {

		switch i := i.(type) {
		case error:
			return prettyerr.Sprint(i)
		case []byte:
			var g interface{}
			err := json.Unmarshal(i, &g)
			if err != nil {
				return pretty.Sprint(string(i))
			} else {
				return pretty.Sprint(g)
			}
		}

		return pretty.Sprint(i)
	}

	consoleOutput.FormatTimestamp = func(i interface{}) string {
		return time.Now().Format("[15:04:05.000000]")
	}

	consoleOutput.FormatCaller = func(i interface{}) string {
		if i == nil {
			return ""
		}
		s := fmt.Sprintf("%s", i)
		tot := strings.Split(s, ":")
		if len(tot) != 2 {
			return snake.FormatCaller(tot[0], 0)
		}

		in, err := strconv.Atoi(tot[1])
		if err != nil {
			return snake.FormatCaller(tot[0], 0)
		}

		return snake.FormatCaller(tot[0], in)
	}

	consoleOutput.PartsOrder = []string{"level", "time", "caller", "message"}

	consoleOutput.FieldsExclude = []string{"handler", "tags"}

	l := zerolog.New(consoleOutput).With().Caller().Timestamp().Logger()

	return &l
}
