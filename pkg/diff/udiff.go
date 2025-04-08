// Package diff - UnifiedDiff implementation
// This file contains the functionality for parsing and formatting unified diffs
package diff

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/go-diff/diff"
)

// UnifiedDiff represents a parsed unified diff with methods to process and format it
type UnifiedDiff struct {
	FileDiff *diff.FileDiff
}

// DiffLine represents a line in the unified diff with metadata
type DiffLine struct {
	Content string
	Applied string
	Type    DiffLineType
	Changes []DiffChange // Character-level changes within the line
}

// DiffLineType indicates whether a line is context, addition, or removal
type DiffLineType string

const (
	// DiffLineContext represents an unchanged line shown for context
	DiffLineContext DiffLineType = "context"
	// DiffLineAdded represents a line that was added
	DiffLineAdded DiffLineType = "added"
	// DiffLineRemoved represents a line that was removed
	DiffLineRemoved DiffLineType = "removed"
)

// DiffChange represents a character-level change in a line
type DiffChange struct {
	Text string
	Type DiffChangeType
}

// DiffChangeType indicates whether a change is an addition, removal, or unchanged text
type DiffChangeType string

const (
	// DiffChangeUnchanged represents text that is the same in both versions
	DiffChangeUnchanged DiffChangeType = "unchanged"
	// DiffChangeAdded represents text that was added
	DiffChangeAdded DiffChangeType = "added"
	// DiffChangeRemoved represents text that was removed
	DiffChangeRemoved DiffChangeType = "removed"
)

// DiffHunk represents a hunk in a unified diff (a group of changes)
type DiffHunk struct {
	Header string
	Lines  []DiffLine
}

// ProcessedDiff represents a fully processed diff with all metadata
type ProcessedDiff struct {
	OrigFile string
	NewFile  string
	Hunks    []DiffHunk
}

// lineGroup is used internally for grouping related changes in a diff
type lineGroup struct {
	contextLines []string
	oldLines     []string
	newLines     []string
}

// GenerateUnifiedDiff creates a unified diff between two strings.
// It formats the output as a standard unified diff with context.
func ConvertToRawUnifiedDiffString(want string, got string) string {
	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(want),
		B:        difflib.SplitLines(got),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  5,
	})

	return diff
}

// ParseUnifiedDiff parses a unified diff string into a structured format
// It normalizes the input and uses go-diff to parse the unified diff format
func ParseUnifiedDiff(diffStr string) (*UnifiedDiff, error) {
	if diffStr == "" {
		return nil, errors.New("empty diff string")
	}

	// Normalize whitespace
	diffStr = normalizeWhitespace(diffStr)

	// Parse the unified diff using sourcegraph/go-diff
	fileDiff, err := diff.ParseFileDiff([]byte(diffStr))
	if err != nil {
		// If parsing fails, fall back to our original implementation
		return nil, err
	}

	return &UnifiedDiff{
		FileDiff: fileDiff,
	}, nil
}

// normalizeWhitespace removes common leading whitespace from all lines
func normalizeWhitespace(diffStr string) string {
	lines := strings.Split(diffStr, "\n")

	// Find common whitespace prefix
	commonWhitespace := ""
	for _, line := range lines[0] {
		if line == ' ' || line == '\t' {
			commonWhitespace += string(line)
		} else {
			break
		}
	}

	// Remove common whitespace from all lines
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, commonWhitespace)
	}

	return strings.Join(lines, "\n")
}

// ProcessDiff converts a UnifiedDiff into a ProcessedDiff
// This is the core logic that can be tested separately from color formatting
func (ud *UnifiedDiff) ProcessDiff() (*ProcessedDiff, error) {
	if ud == nil || ud.FileDiff == nil {
		return nil, errors.New("no diff data available")
	}

	result := &ProcessedDiff{
		OrigFile: ud.FileDiff.OrigName,
		NewFile:  ud.FileDiff.NewName,
	}

	// Process each hunk
	for _, hunk := range ud.FileDiff.Hunks {
		processedHunk := DiffHunk{
			Header: fmt.Sprintf("@@ -%d,%d +%d,%d @@%s",
				hunk.OrigStartLine, hunk.OrigLines,
				hunk.NewStartLine, hunk.NewLines,
				hunk.Section),
		}

		// Group related changes for better highlighting
		lineGroups := groupRelatedChanges(strings.Split(string(hunk.Body), "\n"))

		// Process each group of related changes
		for _, group := range lineGroups {
			processedHunk.Lines = append(processedHunk.Lines,
				processLineGroup(group)...,
			)
		}

		result.Hunks = append(result.Hunks, processedHunk)
	}

	return result, nil
}

// processLineGroup handles a group of related changes and converts them to DiffLine objects
func processLineGroup(group lineGroup) []DiffLine {
	var result []DiffLine

	// Add context lines
	for _, line := range group.contextLines {
		result = append(result, DiffLine{
			Content: line,
			Type:    DiffLineContext,
			Changes: []DiffChange{
				{Text: line, Type: DiffChangeUnchanged},
			},
		})
	}

	// Process changes - if we have a 1:1 mapping, do character-level diffing
	if len(group.oldLines) == 1 && len(group.newLines) == 1 {
		oldLine := group.oldLines[0]
		newLine := group.newLines[0]

		// Process the removed line with character-level changes
		removedLine := processLineChanges(oldLine, newLine, DiffLineRemoved)
		result = append(result, removedLine)

		// Process the added line with character-level changes
		addedLine := processLineChanges(oldLine, newLine, DiffLineAdded)
		result = append(result, addedLine)
	} else {
		// Handle non-1:1 mappings
		// Add all removals
		for _, line := range group.oldLines {
			result = append(result, DiffLine{
				Content: line,
				Type:    DiffLineRemoved,
				Changes: []DiffChange{
					{Text: line, Type: DiffChangeRemoved},
				},
			})
		}

		// Add all additions
		for _, line := range group.newLines {
			result = append(result, DiffLine{
				Content: line,
				Type:    DiffLineAdded,
				Changes: []DiffChange{
					{Text: line, Type: DiffChangeAdded},
				},
			})
		}
	}

	return result
}

// PrettyPrint formats the unified diff with colors
func (ud *UnifiedDiff) PrettyPrint() string {
	processed, err := ud.ProcessDiff()
	if err != nil {
		return err.Error()
	}

	return FormatDiff(processed)
}

// FormatDiff applies color formatting to a ProcessedDiff
func FormatDiff(diff *ProcessedDiff) string {
	prevNoColor := color.NoColor
	defer func() {
		color.NoColor = prevNoColor
	}()
	color.NoColor = false

	var result strings.Builder
	result.WriteString("\n")

	// Process each hunk
	for _, hunk := range diff.Hunks {
		// Write hunk header
		result.WriteString(color.New(color.Faint).Sprintf("%s\n", hunk.Header))

		// Process each line
		for _, line := range hunk.Lines {
			result.WriteString(formatLine(line))
			result.WriteString("\n")
		}

		result.WriteString("\n")
	}

	return result.String()
}

// formatLine formats a single diff line with appropriate colors
func formatLine(line DiffLine) string {
	switch line.Type {
	case DiffLineAdded:
		return formatLineWithPrefix("[got]   -", line, color.New(color.FgRed))
	case DiffLineRemoved:
		return formatLineWithPrefix("[want]  +", line, color.New(color.FgBlue))
	case DiffLineContext:
		return formatLineWithPrefix("         ", line, color.New(color.Faint))
	default:
		return line.Content
	}
}

// formatLineWithPrefix formats a line with the given prefix and color
func formatLineWithPrefix(prefix string, line DiffLine, lineColor *color.Color) string {
	// Format the line differently based on whether we have character-level diffs
	// if len(line.Changes) == 1 {
	// 	// Simple line diff
	// 	return formatSimpleLine(prefix, line.Content, lineColor)
	// } else {
	// Line with character-level changes
	return formatLineChanges(line, lineColor)
	// }
}

// formatSimpleLine formats a line without character-level highlighting
func formatSimpleLine(prefix string, content string, lineColor *color.Color) string {
	return fmt.Sprintf("%s%s",
		color.New(color.Bold).Sprint(prefix),
		formatStartingWhitespace(content, lineColor),
	)
}

// formatLineChanges formats a line with character-level highlighting
func formatLineChanges(line DiffLine, lineColor *color.Color) string {
	prefix := NewColoredString("")
	shouldBold := false
	switch line.Type {
	case DiffLineAdded:
		shouldBold = true
		r := crs("[", color.Bold)
		r = append(r, crs("got", color.Bold, color.FgRed)...)
		r = append(r, crs("]", color.Bold)...)
		r = append(r, crs("    ")...)
		prefix.MultiAppendToStart(r...)
	case DiffLineRemoved:
		shouldBold = true
		r := crs("[", color.Bold)
		r = append(r, crs("want", color.Bold, color.FgBlue)...)
		r = append(r, crs("]", color.Bold)...)
		r = append(r, crs("   ")...)
		prefix.MultiAppendToStart(r...)
	default:
		prefix.MultiAppendToStart(crs(strings.Repeat(" ", 9))...)
	}

	working := NewColoredString("")
	// Render character changes with appropriate colors
	for _, change := range line.Changes {
		switch change.Type {
		case DiffChangeAdded:
			crz := crs(change.Text, color.FgHiGreen, color.Bold, color.Underline)
			for _, cr := range crz {
				cr.MarkIsSpecial()
			}
			working.MultiAppendToEnd(crz...)
		case DiffChangeRemoved:
			crz := crs(change.Text, color.FgHiRed, color.Bold, color.CrossedOut)
			for _, cr := range crz {
				cr.MarkIsSpecial()
			}
			working.MultiAppendToEnd(crz...)
		case DiffChangeUnchanged:
			crz := crs(change.Text, color.Faint)
			if shouldBold {
				crz = crs(change.Text)
			}
			working.MultiAppendToEnd(crz...)
		}
	}

	var result strings.Builder

	working.Annotate(color.New(color.Faint))

	result.WriteString(prefix.ColoredString())
	// result.WriteString(" | ")
	result.WriteString(working.ColoredString())

	return result.String()
}

// groupRelatedChanges groups lines into related changes for better diffing
func groupRelatedChanges(lines []string) []lineGroup {
	var groups []lineGroup
	var currentGroup lineGroup

	// Skip the last line if it's empty
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		prefix := line[0]
		content := line[1:]

		switch prefix {
		case ' ': // Context line
			// If we have pending changes and hit a context line,
			// commit the current group and start a new one
			if len(currentGroup.oldLines) > 0 || len(currentGroup.newLines) > 0 {
				groups = append(groups, currentGroup)
				currentGroup = lineGroup{}
			}
			currentGroup.contextLines = append(currentGroup.contextLines, content)
		case '-': // Removed line
			currentGroup.oldLines = append(currentGroup.oldLines, content)
		case '+': // Added line
			currentGroup.newLines = append(currentGroup.newLines, content)
		}
	}

	// Don't forget to add the last group
	if len(currentGroup.oldLines) > 0 || len(currentGroup.newLines) > 0 || len(currentGroup.contextLines) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

// processLineChanges performs character-level diffing between two lines
func processLineChanges(oldLine, newLine string, lineType DiffLineType) DiffLine {
	// Use diffmatchpatch for character-level diffing
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldLine, newLine, false)

	result := DiffLine{
		Content: oldLine,
		Type:    lineType,
	}

	// Convert dmp diffs to our own DiffChange format
	for _, d := range diffs {
		changeType := DiffChangeUnchanged

		if lineType == DiffLineRemoved {
			switch d.Type {
			case diffmatchpatch.DiffDelete:
				changeType = DiffChangeRemoved
			case diffmatchpatch.DiffInsert:
				// Skip insertions when showing removed lines
				continue
			}
		} else if lineType == DiffLineAdded {
			switch d.Type {
			case diffmatchpatch.DiffInsert:
				changeType = DiffChangeAdded
			case diffmatchpatch.DiffDelete:
				// Skip deletions when showing added lines
				continue
			}
		}

		result.Changes = append(result.Changes, DiffChange{
			Text: d.Text,
			Type: changeType,
		})
	}

	return result
}
