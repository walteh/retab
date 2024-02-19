package terrors

import "slices"

func GetChain(err error) []error {
	errs := []error{}
	for err != nil {
		errs = append(errs, err)
		if we, ok := err.(*wrapError); ok {
			err = we.err
		} else {
			break
		}
	}

	return errs
}

func GetDeepest(err error) error {
	errs := GetChain(err)
	if len(errs) == 0 {
		return nil
	}

	return errs[len(errs)-1]
}

func GetDeepestTerror(err error) *wrapError {
	errs := GetChain(err)

	if len(errs) == 0 {
		return nil
	}

	slices.Reverse(errs)

	for _, e := range errs {
		if we, ok := e.(*wrapError); ok {
			return we
		}
	}

	return nil
}
