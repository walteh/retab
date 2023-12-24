package scobra

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/walteh/terrors"
)

type ErrHandledByPrintingToConsole struct {
	ref error
}

func IsHandledByPrintingToConsole(err error) bool {
	_, ok := terrors.Into[*ErrHandledByPrintingToConsole](err)
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
	cmd.Println(FormatCommandError(cmd, err))
	return &ErrHandledByPrintingToConsole{err}
}

func FormatCommandError(cmd *cobra.Command, err error) string {

	name := color.New(color.FgHiRed).Sprint(cmd.Name())
	cmd.VisitParents(func(cmd *cobra.Command) {
		if cmd.Name() != "" {
			name = cmd.Name() + " " + name
		}
	})
	// dets := terrors.ExtractErrorDetail(err)

	// TODO: get details from error, don't just print it.
	caller := terrors.FormatErrorCaller(err, name, true)

	// fmt.Println("DETAILS:", dets)

	return fmt.Sprintf("%s - %s", color.New(color.FgRed, color.Bold).Sprint("ERROR"), caller)
}
