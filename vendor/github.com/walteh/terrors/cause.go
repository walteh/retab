package terrors

import (
	"errors"
)

type Framer interface {
	error
	Root() error
	Frame() Frame
	Detail() string
	Simple() string
}

func Cause2(err error) (f Framer, r bool) {
	for {
		we, ok := err.(Framer)
		if !ok {
			return
		}

		r = r || ok
		f = we

		err = we.Root()
		if err == nil {
			return
		}
	}
}

func ListCause(err error) ([]Framer, bool) {
	var frames []Framer

	for {
		we, ok := err.(Framer)
		if !ok {
			return frames, ok
		}

		frames = append(frames, we)

		err = we.Root()
		if err == nil {
			return frames, ok
		}
	}
}

func FirstCause(err error) (Framer, bool) {
	for {
		if err == nil {
			return nil, false
		}
		frm, ok := err.(Framer)
		if !ok {
			err = errors.Unwrap(err)
		} else {
			return frm, true
		}
	}
}
