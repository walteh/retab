// Package diff - Diff Output Enrichment
// This file contains functions to enhance and format diff outputs
package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// EnrichCmpDiff enhances a diff produced by cmp.Diff with colors and formatting
// to make it more readable.
func EnrichCmpDiff(diff string) string {
	if diff == "" {
		return ""
	}

	// Save and restore color state
	prevNoColor := color.NoColor
	defer func() {
		color.NoColor = prevNoColor
	}()
	color.NoColor = false

	// Define standard prefixes for expected and actual values
	expectedPrefix := formatWantPrefix()
	actualPrefix := formatGotPrefix()

	var result strings.Builder
	result.WriteString("\n")

	// Process each line
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Format the line based on its content
		switch {
		case strings.HasPrefix(line, "-"):
			content := strings.TrimPrefix(line, "-")
			result.WriteString(actualPrefix)
			result.WriteString(" | ")
			result.WriteString(color.New(color.FgRed).Sprint(content))
			result.WriteString("\n")
		case strings.HasPrefix(line, "+"):
			content := strings.TrimPrefix(line, "+")
			result.WriteString(expectedPrefix)
			result.WriteString(" | ")
			result.WriteString(color.New(color.FgBlue).Sprint(content))
			result.WriteString("\n")
		default:
			result.WriteString(strings.Repeat(" ", 9))
			result.WriteString(" | ")
			result.WriteString(color.New(color.Faint).Sprint(line))
			result.WriteString("\n")
		}
	}

	return result.String()
}

// AltEnrichUnifiedDiff provides an alternative implementation of UnifiedDiff enrichment
// This uses the structured diff parser and renderer
func AltEnrichUnifiedDiff(diff string) string {
	if diff == "" {
		return ""
	}

	ud, err := ParseUnifiedDiff(diff)
	if err != nil {
		// If parsing fails, fall back to the basic enrichment
		return EnrichUnifiedDiff(diff)
	}

	return ud.PrettyPrint()
}

// EnrichUnifiedDiff enhances a unified diff with colors and formatting
// to make it more readable.
func EnrichUnifiedDiff(diff string) string {
	if diff == "" {
		return ""
	}

	// Save and restore color state
	prevNoColor := color.NoColor
	defer func() {
		color.NoColor = prevNoColor
	}()
	color.NoColor = false

	// Define standard prefixes for expected and actual values
	expectedPrefix := formatWantPrefix()
	actualPrefix := formatGotPrefix()

	// Format file headers
	diff = formatFileHeaders(diff)

	// Split the lines by \n and normalize common whitespace
	diff = normalizeWhitespace(diff)

	// Process the diff content by sections
	var result []string
	for i, section := range strings.Split(diff, "\n@@") {
		if i == 0 {
			// First section is the header
			result = append(result, section)
		} else {
			// Process hunk headers and content
			hunkResult := processHunkSection(section, expectedPrefix, actualPrefix)
			result = append(result, hunkResult...)
		}

		// Add blank line between sections
		result = append(result, "")
	}

	return "\n" + strings.Join(result, "\n")
}

// Helper functions

// formatWantPrefix formats the prefix for expected (want) values
func formatWantPrefix() string {
	return fmt.Sprintf("[%s] %s",
		color.New(color.FgBlue, color.Bold).Sprint("want"),
		color.New(color.Faint).Sprint(" +"))
}

// formatGotPrefix formats the prefix for actual (got) values
func formatGotPrefix() string {
	return fmt.Sprintf("[%s] %s",
		color.New(color.Bold, color.FgRed).Sprint("got"),
		color.New(color.Faint).Sprint("  -"))
}

// formatFileHeaders enhances file header lines with colors
func formatFileHeaders(diff string) string {
	diff = strings.ReplaceAll(diff, "--- Expected", fmt.Sprintf("%s %s [%s]",
		color.New(color.Faint).Sprint("---"),
		color.New(color.FgBlue).Sprint("want"),
		color.New(color.FgBlue, color.Bold).Sprint("want")))

	diff = strings.ReplaceAll(diff, "+++ Actual", fmt.Sprintf("%s %s [%s]",
		color.New(color.Faint).Sprint("+++"),
		color.New(color.FgRed).Sprint("got"),
		color.New(color.FgRed, color.Bold).Sprint("got")))

	return diff
}

// processHunkSection formats a hunk within a diff
func processHunkSection(hunkText string, expectedPrefix string, actualPrefix string) []string {
	result := []string{}
	lines := strings.Split(hunkText, "\n")

	// First line of the hunk is the header
	if len(lines) > 0 {
		result = append(result, color.New(color.Faint).Sprint("@@"+lines[0]))
		lines = lines[1:]
	}

	// Process content lines
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Determine line type and format accordingly
		prefix := line[0]
		content := line[1:]

		switch prefix {
		case '-':
			result = append(result, expectedPrefix+
				formatStartingWhitespace(content, color.New(color.FgBlue)))
		case '+':
			result = append(result, actualPrefix+
				formatStartingWhitespace(content, color.New(color.FgRed)))
		case ' ':
			result = append(result, strings.Repeat(" ", 9)+
				formatStartingWhitespace(content, color.New(color.Faint)))
		default:
			if line != "" {
				// Handle any other line type
				result = append(result, strings.Repeat(" ", 9)+
					formatStartingWhitespace(line, color.New(color.Faint)))
			}
		}
	}

	return result
}
