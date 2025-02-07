//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	fmtcmd "github.com/walteh/retab/v2/cmd/retab-wasm/fmt"
)

func main() {
	// Initialize the retab object
	retab := map[string]interface{}{
		"fmt": wrapResult(fmtcmd.Fmt),
	}
	js.Global().Set("retab", js.ValueOf(retab))

	// Set ready flag to indicate initialization is complete
	js.Global().Set("retab_initialized", js.ValueOf(true))

	fmt.Println("retab-wasm initialized")
	<-make(chan bool)
}

func wrapResult[T any](fn func(this js.Value, args []js.Value) (T, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		result, err := fn(this, args)
		return map[string]any{
			"result": result,
			"error":  err,
		}
	})
}
