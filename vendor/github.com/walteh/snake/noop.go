package snake

type noopResolver[A any] struct {
}

func (me *noopResolver[A]) Run() (a A, err error) {
	return a, err
}

func NewNoopMethod[A any]() TypedRunner[*noopResolver[A]] {
	return GenRunResolver_In00_Out02(&noopResolver[A]{})
}

type noopAsker[A any] struct {
}

func (me *noopAsker[A]) Run(a A) (err error) {
	return err
}

func NewNoopAsker[A any]() TypedRunner[*noopAsker[A]] {
	return GenRunResolver_In01_Out01(&noopAsker[A]{})
}
