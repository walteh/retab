// Package diff provides utilities for generating, comparing, and visualizing differences
// between expected and actual values in various formats.
//
// The package offers multiple ways to compare values:
// - Generic type comparisons using go-cmp
// - String-based diffing with unified diff format
// - Character-level diffing for detailed text comparison
//
// It's designed to be used in testing scenarios but can be used in any context
// where difference visualization is needed.
package diff

import (
	"reflect"
	"slices"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// DiffResult represents the output of a diff operation
type DiffResult struct {
	// Content is the formatted diff content
	Content string
	// IsEqual is true if there is no difference
	IsEqual bool
}

// Differ defines the interface for types that can generate diffs
type Differ interface {
	// Diff generates a diff between two values
	Diff(want, got interface{}) DiffResult
}

// DiffFormatter defines the interface for formatters that can render diffs
type DiffFormatter interface {
	// Format formats a diff string with visual enhancements
	Format(diff string) string
}

// formatStartingWhitespace formats leading whitespace characters to be visible while maintaining proper spacing
// Example:
//
//	Input:  "    \t  hello"
//	Output: "····→···hello"
//
// Where:
//
//	· represents a space (Middle Dot U+00B7)
//	→ represents a tab (Rightwards Arrow U+2192)

func formatStartingWhitespace(s string, colord *color.Color) string {
	out := color.New(color.Bold).Sprint(" | ")

	return out + applyWhitespaceColor(s, colord)
}

func applyWhitespaceColor(s string, colord *color.Color) string {
	out := ""
	for j, char := range s {
		switch char {
		case ' ':
			out += color.New(color.Faint).Sprint("∙") // ⌷
		case '\t':
			out += color.New(color.Faint).Sprint("→   ") // → └──▹
		default:
			wrk := s
			trailing := getFormattedTrailingWhitespace(wrk[j:])
			fomrmattedTrail := color.New(color.Faint).Sprint(string(trailing))
			return out + formatInternalWhitespace(wrk[j:len(wrk)-len(trailing)], colord) + fomrmattedTrail
		}
	}
	return out
}

func formatInternalWhitespace(s string, colord *color.Color) string {
	out := ""
	for _, char := range s {
		switch char {
		case ' ':
			out += color.New(color.Faint).Sprint("∙") // ⌷
		case '\t':
			out += color.New(color.Faint).Sprint("→   ") // → └──▹
		default:
			// if colord == nil {
			// 	out += string(char)
			// } else {
			out += colord.Sprint(string(char))
			// }
		}
	}
	return out
}

func getFormattedTrailingWhitespace(s string) []rune {
	out := []rune{}
	rstr := []rune(s)
	slices.Reverse(rstr)
	for _, char := range rstr {
		switch char {
		case ' ':
			out = append(out, '∙')
		case '\t':
			out = append(out, '→')
		// case '\n':
		// 	out += color.New(color.Faint, color.FgHiGreen).Sprint("↵") // ↵
		default:
			return out
		}
	}
	return out
}

// func colorizeWhiteSpace(s string, defaultColor *color.Color) string {

// TypedDiff performs a diff operation between two values of the same type,
// considering all fields (exported and unexported).
// It supports various types including reflect.Type, reflect.Value, string, and others.

func TypedDiff[T any](want T, got T, opts ...OptTestingOptsSetter) string {

	switch any(want).(type) {
	case reflect.Type:
		// Handle reflect.Type values by converting them to string representation
		wantType := ConvolutedFormatReflectType(any(want).(reflect.Type))
		gotType := ConvolutedFormatReflectType(any(got).(reflect.Type))
		return TypedDiff(wantType, gotType, opts...)
	case reflect.Value:
		// Handle reflect.Value by formatting their content
		w := any(want).(reflect.Value)
		g := any(got).(reflect.Value)
		wantValue := ConvolutedFormatReflectValueAsJSON(w)
		gotValue := ConvolutedFormatReflectValueAsJSON(g)
		return TypedDiff(wantValue, gotValue, opts...)
	case string:

		// Handle string values with unified diff
		unified := ConvertToRawUnifiedDiffString(any(want).(string), any(got).(string))
		ud, err := ParseUnifiedDiff(unified)
		if err != nil {
			// Fall back to cmp.Diff for string comparison if unified diff fails
			testOpts := NewTestingOpts(opts...)
			return EnrichCmpDiff(cmp.Diff(got, want, testOpts.cmpOpts...))
		}
		return ud.PrettyPrint()
	default:
		// For all other types, use cmp.Diff
		testOpts := NewTestingOpts(opts...)
		cmpDiff := cmp.Diff(got, want, testOpts.cmpOpts...)
		return EnrichCmpDiff(cmpDiff)
	}
}

// SingleLineStringDiff performs character-level diffing between two strings.
// It highlights specific characters that differ, which is useful for single-line string comparisons.
func SingleLineStringDiff(want string, got string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(want, got, false)
	return dmp.DiffPrettyText(diffs)
}
