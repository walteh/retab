package main

import (
	"context"
	"fmt"
	"os"

	"github.com/walteh/retab/cmd/root"
	"github.com/walteh/snake/scobra"
)

func main() {

	ctx := context.Background()

	scob, cmd, err := root.NewCommand(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd.SilenceErrors = true

	scobra.ExecuteHandlingError(ctx, scob)
}
