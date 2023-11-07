package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/walteh/retab/cmd/root"
	"github.com/walteh/snake"
)

func main() {

	cmd, err := root.NewCommand()
	if err != nil {
		if !snake.IsHandledByPrintingToConsole(err) {
			_, _ = fmt.Print(err)
		}
		os.Exit(1)
	}

	ctx := cmd.Context()

	str, err := snake.DecorateTemplate(ctx, cmd, &snake.DecorateOptions{
		Headings: color.New(color.FgCyan, color.Bold),
		ExecName: color.New(color.FgHiGreen, color.Bold),
		Commands: color.New(color.FgHiRed, color.Faint),
	})
	if err != nil {
		if !snake.IsHandledByPrintingToConsole(err) {
			_, _ = fmt.Print(err)
		}
		os.Exit(1)
	}

	cmd.SetUsageTemplate(str)

	cmd.SilenceErrors = true

	if err := cmd.ExecuteContext(ctx); err != nil {
		if !snake.IsHandledByPrintingToConsole(err) {
			_, _ = fmt.Print(err)
		}
		os.Exit(1)
	}
}
