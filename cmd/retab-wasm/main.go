//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"syscall/js"

	fmtcmd "github.com/walteh/retab/v2/cmd/retab-wasm/fmt"
)

// note about logging: the logging will go to the extenstion host process, so if we want to actually return logs to the extension,
// we need to do some extra work
// - like intercepting the logs and returning them to the extension in the response

func main() {
	ctx := context.Background()

	// Initialize the retab object
	retab := map[string]interface{}{
		"fmt": wrapResult(ctx, fmtcmd.Fmt),
	}
	js.Global().Set("retab", js.ValueOf(retab))

	// Set ready flag to indicate initialization is complete
	js.Global().Set("retab_initialized", js.ValueOf(true))

	fmt.Println("[retab-golang-wasm] initialized")
	<-make(chan bool)
}

func wrapResult[T any](ctx context.Context, fn func(ctx context.Context, this js.Value, args []js.Value) (T, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		result, err := fn(ctx, this, args)
		if err != nil {
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
