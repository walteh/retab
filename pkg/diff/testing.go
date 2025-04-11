// Package diff - Testing Utilities
// This file contains testing utilities for comparing values and generating diff reports
package diff

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
	"gitlab.com/tozd/go/errors"
)

// ValueComparison provides methods for comparing different types of values
type ValueComparison struct {
	t *testing.T
}

// NewValueComparison creates a new value comparison helper for testing
func NewValueComparison(t *testing.T) *ValueComparison {
	return &ValueComparison{t: t}
}

// Equal compares two values of the same type and returns true if they are equal
func Equal[T any](t *testing.T, want, got T, opts ...OptTestingOptsSetter) bool {
	t.Helper()
	return knownTypeEqual(t, want, got, opts...)
}

// RequireEqual compares two values and fails the test if they are not equal
func RequireEqual[T any](t *testing.T, want, got T, opts ...OptTestingOptsSetter) {
	t.Helper()
	if !knownTypeEqual(t, want, got, opts...) {
		require.Fail(t, "value mismatch")

	}
}

// TypeEqual compares two reflect.Type values and returns true if they are equal
func TypeEqual(t *testing.T, want, got reflect.Type, opts ...OptTestingOptsSetter) bool {
	t.Helper()
	return unknownTypeEqual(t, want, got, opts...)
}

// RequireTypeEqual compares two reflect.Type values and fails the test if they are not equal
func RequireTypeEqual(t *testing.T, want, got reflect.Type, opts ...OptTestingOptsSetter) {
	t.Helper()
	if !unknownTypeEqual(t, want, got, opts...) {
		require.Fail(t, "type mismatch")
	}
}

// ValueEqual compares two reflect.Value values and returns true if they are equal
func ValueEqual(t *testing.T, want, got reflect.Value) bool {
	t.Helper()
	return unknownValueEqualAsJSON(t, want, got)
}

// RequireValueEqual compares two reflect.Value values and fails the test if they are not equal
func RequireValueEqual(t *testing.T, want, got reflect.Value) {
	t.Helper()
	if !unknownValueEqualAsJSON(t, want, got) {
		require.Fail(t, "value mismatch")
	}
}

// Legacy functions - maintained for backward compatibility
// These functions directly call into the new non-method functions

// RequireUnknownTypeEqual compares two reflect.Type values and fails the test if they are not equal
// This is maintained for backward compatibility
func RequireUnknownTypeEqual(t *testing.T, want, got reflect.Type, opts ...OptTestingOptsSetter) {
	t.Helper()
	RequireTypeEqual(t, want, got, opts...)
}

// RequireUnknownValueEqualAsJSON compares two reflect.Value values and fails the test if they are not equal
// This is maintained for backward compatibility
func RequireUnknownValueEqualAsJSON(t *testing.T, want, got reflect.Value) {
	t.Helper()
	RequireValueEqual(t, want, got)
}

// RequireKnownValueEqual compares two values of the same type and fails the test if they are not equal
// This is maintained for backward compatibility
func RequireKnownValueEqual[T any](t *testing.T, want, got T, opts ...OptTestingOptsSetter) {
	t.Helper()
	RequireEqual(t, want, got, opts...)
}

// Core implementation functions

// unknownValueEqualAsJSON compares two reflect.Value values and returns true if they are equal
// It uses typed diff functionality with JSON formatting for the comparison
func unknownValueEqualAsJSON(t *testing.T, want, got reflect.Value) bool {
	t.Helper()
	td := TypedDiff(want, got)
	if td != "" {
		color.NoColor = false
		str := buildDiffReport(t, "VALUE COMPARISON",
			fmt.Sprintf("want type: %s\n", color.YellowString(want.Type().String())),
			fmt.Sprintf("got type:  %s", color.YellowString(got.Type().String())),
			td)
		t.Log("value comparison report:\n" + str)
		return false
	}
	return true
}

// unknownTypeEqual compares two reflect.Type values and returns true if they are equal
// It uses typed diff functionality for the comparison
func unknownTypeEqual(t *testing.T, want, got reflect.Type, opts ...OptTestingOptsSetter) bool {
	t.Helper()
	td := TypedDiff(want, got, opts...)
	if td != "" {
		color.NoColor = false
		str := buildDiffReport(t, "VALUE COMPARISON",
			fmt.Sprintf("want type: %s\n", color.YellowString(want.String())),
			fmt.Sprintf("got type:  %s", color.YellowString(got.String())),
			td)
		t.Log("value comparison report:\n" + str)

		return false
	}
	return true
}

// knownTypeEqual compares two values of the same type and returns true if they are equal
// It uses typed diff functionality for the comparison
func knownTypeEqual[T any](t *testing.T, want, got T, opts ...OptTestingOptsSetter) bool {
	t.Helper()
	td := TypedDiff(want, got, opts...)
	if td != "" {
		color.NoColor = false
		str := buildDiffReport(t, "TYPE COMPARISON",
			fmt.Sprintf("type: %s", color.YellowString(reflect.TypeOf(want).String())),
			"",
			td)
		t.Log("type comparison report:\n" + str)
		ops := NewTestingOpts(opts...)
		_, isString := any(want).(string)
		if ops.logRawDiffOnFail && isString {
			t.Log(ConvertToRawUnifiedDiffString(any(want).(string), any(got).(string)))
		}
		return false
	}
	return true
}

// buildDiffReport creates a formatted diff report with header and details
func buildDiffReport(t *testing.T, title string, header1 string, header2 string, diffContent string) string {
	var result strings.Builder

	// Add report header
	result.WriteString(color.New(color.FgHiYellow, color.Faint).Sprintf("\n\n============= %s START =============\n\n", title))
	result.WriteString(fmt.Sprintf("%s\n\n", color.YellowString("%s", t.Name())))

	// Add type/value information headers if provided
	if header1 != "" {
		result.WriteString(header1 + "\n")
	}
	if header2 != "" {
		result.WriteString(header2 + "\n\n\n")
	}

	// Add diff content
	result.WriteString(shortenOutputIfNeeded(diffContent) + "\n\n")

	// Add report footer
	result.WriteString(color.New(color.FgHiYellow, color.Faint).Sprintf("============= %s END ===============\n\n", title))

	return result.String()
}

// ANSI escape code regex pattern
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

// Strip removes ANSI color codes from a string
func Strip(str string) string {
	return re.ReplaceAllString(str, "")
}

// shortenOutputIfNeeded truncates long diffs to make them more readable
// It keeps the first and last few lines of each section to provide
// a more concise view of the differences
func shortenOutputIfNeeded(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	var buffer []string

	// Helper to process a batch of lines when we're ready to flush them
	flushBuffer := func(buf []string) {
		if len(buf) >= 10 {
			// Add first 5 lines
			result = append(result, buf[:5]...)
			// Add truncation message
			result = append(result, fmt.Sprintf("...truncated %d lines...", len(buf)-10))
			// Add last 5 lines
			result = append(result, buf[len(buf)-5:]...)
		} else {
			// If less than 10 lines, just add them all
			result = append(result, buf...)
		}
	}

	for _, line := range lines {
		stripped := Strip(line)
		isAct := strings.HasPrefix(stripped, "[act]")
		isExp := strings.HasPrefix(stripped, "[exp]")

		// If this is a continuation of the current buffer type
		if (isAct && len(buffer) > 0 && strings.HasPrefix(Strip(buffer[0]), "[act]")) ||
			(isExp && len(buffer) > 0 && strings.HasPrefix(Strip(buffer[0]), "[exp]")) {
			buffer = append(buffer, line)
			continue
		}

		// If we hit a different type of line, flush the buffer
		if len(buffer) > 0 {
			flushBuffer(buffer)
			buffer = nil
		}

		// Start a new buffer if this is an act/exp line
		if isAct || isExp {
			buffer = []string{line}
		} else {
			// Regular line, just add it
			result = append(result, line)
		}
	}

	// Don't forget to flush any remaining buffer
	if len(buffer) > 0 {
		flushBuffer(buffer)
	}

	return strings.Join(result, "\n")
}

type Req struct {
	want    any
	wantSet bool
	got     any
	gotSet  bool
	t       *testing.T
	opts    []OptTestingOptsSetter
}

func Require(t *testing.T) *Req {
	return &Req{t: t, opts: []OptTestingOptsSetter{}}
}

func (r *Req) Want(want any) *Req {
	r.want = want
	r.wantSet = true
	return r
}

func (r *Req) Got(got any) *Req {
	r.got = got
	r.gotSet = true
	return r
}

func (r *Req) Opts(opts ...OptTestingOptsSetter) *Req {
	r.opts = append(r.opts, opts...)
	return r
}

func (r *Req) Equals() {
	r.t.Helper()

	if !r.wantSet || !r.gotSet {
		r.t.Fatalf("want and got must be set")
	}

	wantStr, err := convertToString(r.want)
	if err != nil {
		r.t.Fatalf("failed to convert want to string: %v", err)
	}
	gotStr, err := convertToString(r.got)
	if err != nil {
		r.t.Fatalf("failed to convert got to string: %v", err)
	}

	RequireEqual(r.t, wantStr, gotStr, r.opts...)
}

func convertToString(v any) (string, error) {
	switch v := v.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case io.Reader:
		bs, err := io.ReadAll(v)
		if err != nil {
			return "", err
		}
		return string(bs), nil
	default:
		return "", errors.Errorf("tried to convert unsupported type to string: %T", v)
	}
}
