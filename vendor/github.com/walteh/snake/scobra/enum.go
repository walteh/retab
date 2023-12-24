package scobra

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/walteh/snake"
)

var (
	_ pflag.Value = &wrappedEnum{}
)

func NewWrappedEnum(opt snake.Enum) *wrappedEnum {
	strt := &wrappedEnum{internal: opt}
	return strt
}

type wrappedEnum struct {
	internal snake.Enum
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *wrappedEnum) Set(v string) error {
	return e.internal.SetCurrent(v)
}

func (e *wrappedEnum) String() string {
	if e.internal.CurrentPtr() == nil {
		return ""
	}
	return *e.internal.CurrentPtr()
}

// Type is only used in help text
func (e *wrappedEnum) Type() string {
	return "string"
}

func (e *wrappedEnum) CompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return e.internal.Options(), cobra.ShellCompDirectiveDefault
}

func (e *wrappedEnum) Assign(cmd *cobra.Command, key string, descritpion string) error {
	cmd.Flags().Var(e, key, descritpion)
	err := cmd.RegisterFlagCompletionFunc(key, e.CompletionFunc)
	if err != nil {
		return err
	}
	return nil
}
