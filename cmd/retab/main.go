//go:build !js

package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	fmtcmd "github.com/walteh/retab/v2/cmd/retab/fmt"
)

func main() {
	ctx := context.Background()

	cmd := &cobra.Command{
		Use: "retab",
	}

	cmd.AddCommand(fmtcmd.NewFmtCommand())

	info, ok := debug.ReadBuildInfo()
	if !ok {
		cmd.Version = "unknown"
	} else {
		cmd.Version = info.Main.Version
	}

	cmdVersion := &cobra.Command{
		Use: "raw-version",
		Run: func(cmdz *cobra.Command, args []string) {
			cmdz.Println(cmd.Version)
		},
		Hidden: true,
	}

	cmd.AddCommand(cmdVersion)

	cmd.InitDefaultVersionFlag()

	// cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
