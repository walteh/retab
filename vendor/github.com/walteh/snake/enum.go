package snake

import (
	"strings"

	"github.com/go-faster/errors"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

func NewWrappedEnum[I ~string](values ...I) *wrappedEnum[I] {
	return &wrappedEnum[I]{values: values}
}

type wrappedEnum[I ~string] struct {
	current *string
	values  []I
}

// String is used both by fmt.Print and by Cobra in help text
func (e *wrappedEnum[I]) String() string {
	return string(*e.current)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *wrappedEnum[I]) Set(v string) error {
	if slices.Contains(e.values, I(v)) {
		e.current = &v
		return nil
	}
	return errors.Errorf("invalid value %q, expected one of %s", v, strings.Join(e.ValuesStringSlice(), ", "))
}

// Type is only used in help text
func (e *wrappedEnum[I]) Type() string {
	return "myEnum"
}

func (e *wrappedEnum[I]) ValuesStringSlice() []string {
	wrk := make([]string, len(e.values))
	for i, v := range e.values {
		wrk[i] = string(v)
	}
	return wrk
}

// func     myCmd.RegisterFlagCompletionFunc("myenum", myEnumCompletion)

// myEnumCompletion should probably live next to the myEnum definition
func (e *wrappedEnum[I]) CompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return e.ValuesStringSlice(), cobra.ShellCompDirectiveDefault
}

func (e *wrappedEnum[I]) Assign(cmd *cobra.Command, key string, descritpion string) error {
	cmd.Flags().Var(e, key, "descritpion")
	err := cmd.RegisterFlagCompletionFunc(key, e.CompletionFunc)
	if err != nil {
		return err
	}
	return nil
}
