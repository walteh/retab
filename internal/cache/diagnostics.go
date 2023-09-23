package cache

import (
	"encoding/json"
	"fmt"
	"go/token"
	"log"
	urlpkg "net/url"

	"github.com/walteh/retab/gen/gopls/frob"
	"github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/gen/gopls/span"
	"github.com/walteh/retab/internal/command"
	"github.com/walteh/retab/internal/source"
	"golang.org/x/tools/go/analysis"
)

// toSourceDiagnostic converts a gobDiagnostic to "source" form.
func toSourceDiagnostic(srcAnalyzer *source.Analyzer, gobDiag *gobDiagnostic) *source.Diagnostic {
	var related []protocol.DiagnosticRelatedInformation
	for _, gobRelated := range gobDiag.Related {
		related = append(related, protocol.DiagnosticRelatedInformation(gobRelated))
	}

	kinds := srcAnalyzer.ActionKind
	if len(srcAnalyzer.ActionKind) == 0 {
		kinds = append(kinds, protocol.QuickFix)
	}
	fixes := suggestedAnalysisFixes(gobDiag, kinds)
	if srcAnalyzer.Fix != "" {
		cmd, err := command.NewApplyFixCommand(gobDiag.Message, command.ApplyFixArgs{
			URI:   gobDiag.Location.URI,
			Range: gobDiag.Location.Range,
			Fix:   srcAnalyzer.Fix,
		})
		if err != nil {
			// JSON marshalling of these argument values cannot fail.
			log.Fatalf("internal error in NewApplyFixCommand: %v", err)
		}
		for _, kind := range kinds {
			fixes = append(fixes, source.SuggestedFixFromCommand(cmd, kind))
		}
	}

	severity := srcAnalyzer.Severity
	if severity == 0 {
		severity = protocol.SeverityWarning
	}

	diag := &source.Diagnostic{
		URI:      gobDiag.Location.URI.SpanURI(),
		Range:    gobDiag.Location.Range,
		Severity: severity,
		Code:     gobDiag.Code,
		CodeHref: gobDiag.CodeHref,
		Source:   source.AnalyzerErrorKind(gobDiag.Source),
		Message:  gobDiag.Message,
		Related:  related,
		Tags:     srcAnalyzer.Tag,
	}
	if srcAnalyzer.FixesDiagnostic(diag) {
		diag.SuggestedFixes = fixes
	}

	// If the fixes only delete code, assume that the diagnostic is reporting dead code.
	if onlyDeletions(fixes) {
		diag.Tags = append(diag.Tags, protocol.Unnecessary)
	}
	return diag
}

// onlyDeletions returns true if all of the suggested fixes are deletions.
func onlyDeletions(fixes []source.SuggestedFix) bool {
	for _, fix := range fixes {
		if fix.Command != nil {
			return false
		}
		for _, edits := range fix.Edits {
			for _, edit := range edits {
				if edit.NewText != "" {
					return false
				}
				if protocol.ComparePosition(edit.Range.Start, edit.Range.End) == 0 {
					return false
				}
			}
		}
	}
	return len(fixes) > 0
}

func suggestedAnalysisFixes(diag *gobDiagnostic, kinds []protocol.CodeActionKind) []source.SuggestedFix {
	var fixes []source.SuggestedFix
	for _, fix := range diag.SuggestedFixes {
		edits := make(map[span.URI][]protocol.TextEdit)
		for _, e := range fix.TextEdits {
			uri := span.URI(e.Location.URI)
			edits[uri] = append(edits[uri], protocol.TextEdit{
				Range:   e.Location.Range,
				NewText: string(e.NewText),
			})
		}
		for _, kind := range kinds {
			fixes = append(fixes, source.SuggestedFix{
				Title:      fix.Message,
				Edits:      edits,
				ActionKind: kind,
			})
		}

	}
	return fixes
}

var diagnosticsCodec = frob.CodecFor[[]gobDiagnostic]()

type gobDiagnostic struct {
	Location       protocol.Location
	Severity       protocol.DiagnosticSeverity
	Code           string
	CodeHref       string
	Source         string
	Message        string
	SuggestedFixes []gobSuggestedFix
	Related        []gobRelatedInformation
	Tags           []protocol.DiagnosticTag
}

type gobRelatedInformation struct {
	Location protocol.Location
	Message  string
}

type gobSuggestedFix struct {
	Message    string
	TextEdits  []gobTextEdit
	Command    *gobCommand
	ActionKind protocol.CodeActionKind
}

type gobCommand struct {
	Title     string
	Command   string
	Arguments []json.RawMessage
}

type gobTextEdit struct {
	Location protocol.Location
	NewText  []byte
}

// toGobDiagnostic converts an analysis.Diagnosic to a serializable gobDiagnostic,
// which requires expanding token.Pos positions into protocol.Location form.
func toGobDiagnostic(posToLocation func(start, end token.Pos) (protocol.Location, error), a *analysis.Analyzer, diag analysis.Diagnostic) (gobDiagnostic, error) {
	var fixes []gobSuggestedFix
	for _, fix := range diag.SuggestedFixes {
		var gobEdits []gobTextEdit
		for _, textEdit := range fix.TextEdits {
			loc, err := posToLocation(textEdit.Pos, textEdit.End)
			if err != nil {
				return gobDiagnostic{}, fmt.Errorf("in SuggestedFixes: %w", err)
			}
			gobEdits = append(gobEdits, gobTextEdit{
				Location: loc,
				NewText:  textEdit.NewText,
			})
		}
		fixes = append(fixes, gobSuggestedFix{
			Message:   fix.Message,
			TextEdits: gobEdits,
		})
	}

	var related []gobRelatedInformation
	for _, r := range diag.Related {
		loc, err := posToLocation(r.Pos, r.End)
		if err != nil {
			return gobDiagnostic{}, fmt.Errorf("in Related: %w", err)
		}
		related = append(related, gobRelatedInformation{
			Location: loc,
			Message:  r.Message,
		})
	}

	loc, err := posToLocation(diag.Pos, diag.End)
	if err != nil {
		return gobDiagnostic{}, err
	}

	// The Code column of VSCode's Problems table renders this
	// information as "Source(Code)" where code is a link to CodeHref.
	// (The code field must be nonempty for anything to appear.)
	diagURL := effectiveURL(a, diag)
	code := "default"
	if diag.Category != "" {
		code = diag.Category
	}

	return gobDiagnostic{
		Location: loc,
		// Severity for analysis diagnostics is dynamic,
		// based on user configuration per analyzer.
		Code:           code,
		CodeHref:       diagURL,
		Source:         a.Name,
		Message:        diag.Message,
		SuggestedFixes: fixes,
		Related:        related,
		// Analysis diagnostics do not contain tags.
	}, nil
}

// effectiveURL computes the effective URL of diag,
// using the algorithm specified at Diagnostic.URL.
func effectiveURL(a *analysis.Analyzer, diag analysis.Diagnostic) string {
	u := diag.URL
	if u == "" && diag.Category != "" {
		u = "#" + diag.Category
	}
	if base, err := urlpkg.Parse(a.URL); err == nil {
		if rel, err := urlpkg.Parse(u); err == nil {
			u = base.ResolveReference(rel).String()
		}
	}
	return u
}
