// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package terrors

import (
	"runtime"
	"strings"
)

// A Frame contains part of a call stack.
type Frame struct {
	// Make room for three PCs: the one we were asked for, what it called,
	// and possibly a PC for skipPleaseUseCallersFrames. See:
	// https://go.googlesource.com/go/+/032678e0fb/src/runtime/extern.go#169
	frames [3]uintptr
}

// Caller returns a Frame that describes a frame on the caller's stack.
// The argument skip is the number of frames to skip over.
// Caller(0) returns the frame for the caller of Caller.
func Caller(skip int) Frame {
	var s Frame
	runtime.Callers(skip+1, s.frames[:])
	return s
}

// Location reports the file, line, and function of a frame.
//
// The returned function may be "" even if file and line are not.
func (f Frame) Location() (pkg, function, file string, line int) {
	frames := runtime.CallersFrames(f.frames[:])
	if _, ok := frames.Next(); !ok {
		return "", "", "", 0
	}
	fr, ok := frames.Next()
	if !ok {
		return "", "", "", 0
	}
	// get the name of the package

	pkg, function = GetPackageAndFuncFromFuncName(fr.Function)

	return pkg, function, FileNameOfPath(fr.File), fr.Line
}

func GetPackageAndFuncFromFuncName(pc string) (pkg, function string) {
	// funcName := runtime.FuncForPC(pc).Name()
	funcName := pc
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	lastDot := strings.LastIndexByte(funcName[lastSlash:], '.') + lastSlash

	pkg = funcName[:lastDot]
	fname := funcName[lastDot+1:]

	if strings.Contains(pkg, ".(") {
		splt := strings.Split(pkg, ".(")
		pkg = splt[0]
		fname = "(" + splt[1] + "." + fname
	}

	pkg = strings.TrimPrefix(pkg, "github.com/")

	return pkg, fname
}
