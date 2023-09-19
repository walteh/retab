package snake

import (
	"fmt"
	"regexp"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/spf13/cobra"
)

type ErrHandledByPrintingToConsole struct {
	ref error
}

func IsHandledByPrintingToConsole(err error) bool {
	_, ok := errors.Into[*ErrHandledByPrintingToConsole](err)
	return ok
}

func (e *ErrHandledByPrintingToConsole) Error() string {
	return e.ref.Error()
}

func (e *ErrHandledByPrintingToConsole) Unwrap() error {
	return e.ref
}

func HandleErrorByPrintingToConsole(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	cmd.Println(FormatError(cmd, err))
	return &ErrHandledByPrintingToConsole{err}
}

func FormatError(cmd *cobra.Command, err error) string {

	n := color.New(color.FgHiRed).Sprint(cmd.Name())
	cmd.VisitParents(func(cmd *cobra.Command) {
		if cmd.Name() != "" {
			n = cmd.Name() + " " + n
		}
	})
	caller := ""
	if frm, ok := errors.Cause(err); ok {
		_, filestr, linestr := frm.Location()
		caller = FormatCaller(filestr, linestr)
		caller = caller + " - "

	}
	str := fmt.Sprintf("%+s", err)
	prev := ""
	// replace any string that contains "*.Err" with a bold red version using regex
	str = regexp.MustCompile(`\S+\.Err\S*`).ReplaceAllStringFunc(str, func(s string) string {
		prev += color.New(color.FgRed, color.Bold).Sprint(s) + " -> "
		return ""
	})

	return fmt.Sprintf("%s - %s - %s%s%s", color.New(color.FgRed, color.Bold).Sprint("ERROR"), n, caller, prev, color.New(color.FgRed).Sprint(str))
}
