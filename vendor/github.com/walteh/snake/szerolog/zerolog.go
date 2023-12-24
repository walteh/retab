package szerolog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/k0kubun/pp/v3"
	"github.com/rs/zerolog"
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

func NewConsoleLoggerContext(ctx context.Context, level zerolog.Level, writer io.Writer) context.Context {
	verboseLogger := NewVerboseConsoleLogger(writer).With().Logger().Level(level)
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

func NewVerboseConsoleLogger(out io.Writer) *zerolog.Logger {

	consoleOutput := zerolog.ConsoleWriter{Out: out, TimeFormat: time.StampMicro, NoColor: false}

	pretty := pp.New()

	pretty.SetColorScheme(pp.ColorScheme{})

	prettyerr := pp.New()
	prettyerr.SetExportedOnly(false)

	consoleOutput.FormatFieldValue = func(i any) string {

		switch t := i.(type) {
		case error:
			return t.Error()
		case []byte:
			var g any
			err := json.Unmarshal(t, &g)
			if err != nil {
				return pretty.Sprint(string(t))
			} else {
				return pretty.Sprint(g)
			}
		}

		switch reflect.TypeOf(i).Kind() {
		case reflect.Struct:
			return pretty.Sprint(i)
		case reflect.Map:
			return pretty.Sprint(i)
		case reflect.Slice:
			return pretty.Sprint(i)
		case reflect.Array:
			return pretty.Sprint(i)
		case reflect.Ptr:
			return pretty.Sprint(i)
		case reflect.Interface:
			return pretty.Sprint(i)
		case reflect.Func:
			return pretty.Sprint(i)
		case reflect.Chan:
			return pretty.Sprint(i)
		case reflect.UnsafePointer:
			return pretty.Sprint(i)
		case reflect.Uintptr:
			return pretty.Sprint(i)
		}

		return fmt.Sprintf("%v", i)
	}

	consoleOutput.FormatTimestamp = func(i any) string {
		return time.Now().Format("[15:04:05.000000]")
	}

	// consoleOutput.FormatCaller = func(i any) string {
	// 	if i == nil {
	// 		return ""
	// 	}
	// 	s := fmt.Sprintf("%s", i)
	// 	tot := strings.Split(s, ":")
	// 	if len(tot) != 2 {
	// 		return terrors.FormatCaller(tot[0], 0)
	// 	}

	// 	in, err := strconv.Atoi(tot[1])
	// 	if err != nil {
	// 		return terrors.FormatCaller(tot[0], 0)
	// 	}

	// 	return terrors.FormatCaller(tot[0], in)
	// }

	consoleOutput.PartsOrder = []string{"level", "time", "caller", "message"}

	consoleOutput.FieldsExclude = []string{"handler", "tags"}

	consoleOutput.FormatErrFieldValue = func(i any) string {
		return fmt.Sprintf("%v", i)
	}

	l := zerolog.New(consoleOutput).With().Caller().Timestamp().Logger()

	return &l
}
