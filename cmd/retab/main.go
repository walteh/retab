package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	fmtcmd "github.com/walteh/retab/v2/cmd/retab/fmt"
)

func main() {
	ctx := context.Background()

	cmd := &cobra.Command{
		Use: "retab",
	}

	cmd.AddCommand(fmtcmd.NewFmtCommand())

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
