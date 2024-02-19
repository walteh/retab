package terrors

import (
	"errors"
	"slices"
)

// Into finds the first error in err's chain that matches target type T, and if so, returns it.
//
// Into is type-safe alternative to As.
func Into[T error](err error) (val T, ok bool) {
	ok = errors.As(err, &val)
	return val, ok
}

type RecoveryInfo struct {
	DeepestSimpleErrorMessage string
	Suggestion                string
}

func IsRecoverable(err error) (bool, *RecoveryInfo) {
	chain := GetChain(err)

	// we want to get the deepest recoverable error in the chain
	slices.Reverse(chain)

	for _, e := range chain {
		if werr, ok := e.(*wrapError); ok {
			if werr.recovery != nil {
				msg := werr.msg
				if werr.err != nil {
					msg += ": " + werr.err.Error()
				}
				return true, &RecoveryInfo{
					DeepestSimpleErrorMessage: msg,
					Suggestion:                werr.recovery.Suggestion,
				}
			}
		}
	}

	return false, nil
}
