package terrors

import "errors"

// Into finds the first error in err's chain that matches target type T, and if so, returns it.
//
// Into is type-safe alternative to As.
func Into[T error](err error) (val T, ok bool) {
	ok = errors.As(err, &val)
	return val, ok
}
