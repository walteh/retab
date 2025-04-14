package format

import "regexp"

var (
	shebangRe = regexp.MustCompile(`^#![ \t]*/(usr/)?bin/(env[ \t]+)?(\S+)(\s|$)`)
	extRe     = regexp.MustCompile(`\.(\S+)$`)
)

// TODO: consider removing HasShebang in favor of Shebang in v4

// HasShebang reports whether bs begins with a valid shell shebang.
// It supports variations with /usr and env.
func HasShebang(bs []byte) bool {
	return Shebang(bs) != ""
}

// Shebang parses a "#!" sequence from the beginning of the input bytes,
// and returns the shell that it points to.
//
// For instance, it returns "sh" for "#!/bin/sh",
// and "bash" for "#!/usr/bin/env bash".
func Shebang(bs []byte) string {
	m := shebangRe.FindSubmatch(bs)
	if m == nil {
		return ""
	}
	return string(m[3])
}
