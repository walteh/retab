//go:build js && wasm

package main

import (
	"context"
	"strings"
	"syscall/js"

	"github.com/rs/zerolog"
	fmtcmd "github.com/walteh/retab/v2/cmd/retab/fmt"
)

type ZeroLogConsoleWriter struct {
}

func (w *ZeroLogConsoleWriter) Write(p []byte) (n int, err error) {
	js.Global().Call("wasm_log", strings.TrimSpace(string(p)))
	return len(p), nil
}

// note about logging: the logging will go to the extenstion host process, so if we want to actually return logs to the extension,
// we need to do some extra work
// - like intercepting the logs and returning them to the extension in the response

func main() {
	ctx := context.Background()

	ctx = zerolog.New(&ZeroLogConsoleWriter{}).Level(zerolog.DebugLevel).With().Timestamp().Caller().Logger().WithContext(ctx)

	// Initialize the retab object
	retab := map[string]interface{}{
		"fmt": wrapResult(ctx, fmtcmd.Fmt),
	}

	// Set the retab object first
	js.Global().Set("retab", js.ValueOf(retab))

	// Log initialization
	js.Global().Call("wasm_log", "[retab-golang-wasm] initialized")

	// Set ready flag to indicate initialization is complete
	js.Global().Set("retab_initialized", js.ValueOf(true))

	// Keep the program running
	select {}
}

func wrapResult[T any](ctx context.Context, fn func(ctx context.Context, this js.Value, args []js.Value) (T, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		result, err := fn(ctx, this, args)
		if err != nil {
			// Log errors to console for debugging
			js.Global().Get("console").Call("error", "[retab-golang-wasm]", err.Error())
			return map[string]any{
				"result": nil,
				"error":  err.Error(),
			}
		}
		return map[string]any{
			"result": result,
			"error":  nil,
		}
	})
}
