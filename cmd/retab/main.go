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

	scob, _, err := root.NewCommand(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	scobra.ExecuteHandlingError(ctx, scob)
}
