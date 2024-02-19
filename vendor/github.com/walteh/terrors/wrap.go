package terrors

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"

	"github.com/rs/zerolog"
)

type wrapError struct {
	msg      string
	err      error
	frame    Frame
	event    []func(*zerolog.Event) *zerolog.Event
	code     int
	recovery *Recovery
}

type Recovery struct {
	Suggestion string
	State      []any
}

func (e *wrapError) Root() error {
	return e.err
}

func (e *wrapError) Frame() Frame {
	return e.frame
}

func (e *wrapError) Recovery() *Recovery {
	return e.recovery
}

func (e *wrapError) WithRecovery(r string, state ...any) *wrapError {
	e.recovery = &Recovery{r, state}
	return e
}

func (me *wrapError) WithRecoveryf(format string, a ...any) *wrapError {
	return me.WithRecovery(fmt.Sprintf(format, a...))
}

func (e *wrapError) Info() []any {
	return []any{e.msg}
}

func (e *wrapError) Event(gv func(*zerolog.Event) *zerolog.Event) error {
	if gv != nil {
		e.event = append(e.event, gv)
	}
	return e
}

func (e *wrapError) With(name string, value any) *wrapError {
	e.event = append(e.event, func(ev *zerolog.Event) *zerolog.Event {
		return ev.Interface(name, value)
	})
	return e
}

func (e *wrapError) Error() string {
	return InlineChainFormatter(e.Self, e.err)
}

func (e *wrapError) Code() int {
	return e.code
}

func (e *wrapError) WithCode(code int) *wrapError {
	e.code = code
	return e
}

func (e *wrapError) Simple() string {
	return InlineChainFormatter(e.Message, e.err)
}

func (e *wrapError) Complicated() string {
	return FullChainFormatter(e.err)
}

func (e *wrapError) Message() string {
	if e.code != 0 {
		return fmt.Sprintf("ERROR%s%s", ColorCode(e.code), ColorBrackets("msg", e.msg))
	}
	return fmt.Sprintf("ERROR%s", ColorBrackets("msg", e.msg))
}

func (e *wrapError) Self() string {
	return fmt.Sprintf("%s%s", e.Message(), FormatCallerFromFrame(e.Frame()))
}

func (e *wrapError) DetailedSelf() string {
	self := e.Self()

	dets := e.Detail()

	if dets != "" {
		self += fmt.Sprintf("\n\n%s\n\n", dets)
	}

	return self
}

func (e *wrapError) Unwrap() error {
	return e.err
}

// Wrap error with message and caller.
func Wrap(err error, message string) *wrapError {
	return WrapWithCaller(err, message, 1)
}

// Wrapf wraps error with formatted message and caller.
func Wrapf(err error, format string, a ...interface{}) *wrapError {
	return WrapWithCaller(err, fmt.Sprintf(format, a...), 1)
}

func WrapWithCaller(err error, message string, frm int) *wrapError {
	frme := Caller(frm + 1)

	return &wrapError{msg: message, err: err, frame: frme, event: []func(*zerolog.Event) *zerolog.Event{}}
}

func (c *wrapError) MarshalZerologObject(e *zerolog.Event) (err error) {
	for _, ev := range c.event {
		*e = *ev(e)
	}
	return nil
}
