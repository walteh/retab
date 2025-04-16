package cmdfmt

import (
	"fmt"
	"strings"

	"github.com/walteh/retab/v2/pkg/format"
)

func NewDartFormatter(cmds ...string) format.Provider {
	cmds = append(cmds, "format", "--output", "show", "--summary", "none", "--fix")

	return NewExecFormatter(&BasicExternalFormatterOpts{
		Indent: "  ",
		// Targets: []string{"*.dart"},
	}, cmds...)
}

func NewTerraformFormatter(cmds ...string) format.Provider {
	cmds = append(cmds, "fmt", "-write=false", "-list=false")

	return NewExecFormatter(&BasicExternalFormatterOpts{
		Indent: "  ",

		// Targets: []string{"*.tf", "*.tfvars"},
	}, cmds...)
}

// we could maybe read this in too, but we really need the indentation
// to be 2 spaces in the config passed to the command no matter what
// 2 spaces is arbitrary, but it's what we are passing in the Indent
// field of the BasicExternalFormatterOpts for swift
var swiftConfig = /* json */ `
{
	"version": 1,
	"lineLength": 140,
	"indentation": {
		"spaces": 2
	},
	"tabWidth": 4,
	"lineBreakBeforeEachArgument": true,
	"indentConditionalCompilationBlocks": true,
	"prioritizeKeepingFunctionOutputTogether": true,
	"multiElementCollectionTrailingCommas": true,
	"rules": {
		"AlwaysUseLowerCamelCase": false,
		"AmbiguousTrailingClosureOverload": false,
		"NoBlockComments": false,
		"OrderedImports": true,
		"UseLetInEveryBoundCaseVariable": false,
		"UseSynthesizedInitializer": false,
		"AllPublicDeclarationsHaveDocumentation": false,
		"AlwaysUseLiteralForEmptyCollectionInit": false,
		"BeginDocumentationCommentWithOneLineSummary": false,
		"DoNotUseSemicolons": true,
		"DontRepeatTypeInStaticProperties": true,
		"FileScopedDeclarationPrivacy": true,
		"FullyIndirectEnum": true,
		"GroupNumericLiterals": true,
		"IdentifiersMustBeASCII": true,
		"NeverForceUnwrap": false,
		"NeverUseForceTry": false,
		"NeverUseImplicitlyUnwrappedOptionals": false,
		"NoAccessLevelOnExtensionDeclaration": true,
		"NoAssignmentInExpressions": true,
		"NoCasesWithOnlyFallthrough": true,
		"NoEmptyTrailingClosureParentheses": true,
		"NoLabelsInCasePatterns": true,
		"NoLeadingUnderscores": false,
		"NoParensAroundConditions": true,
		"NoPlaygroundLiterals": true,
		"NoVoidReturnOnFunctionSignature": true,
		"OmitExplicitReturns": false,
		"OneCasePerLine": true,
		"OneVariableDeclarationPerLine": true,
		"OnlyOneTrailingClosureArgument": true,
		"ReplaceForEachWithForLoop": true,
		"ReturnVoidInsteadOfEmptyTuple": true,
		"TypeNamesShouldBeCapitalized": true,
		"UseEarlyExits": true,
		"UseExplicitNilCheckInConditions": true,
		"UseShorthandTypeNames": true,
		"UseSingleLinePropertyGetter": true,
		"UseTripleSlashForDocumentationComments": true,
		"UseWhereClausesInForLoops": false,
		"ValidateDocumentationComments": false
	}
}
`

func NewSwiftFormatter(cmds ...string) format.Provider {
	scfg := strings.ReplaceAll(swiftConfig, "\n", "")
	scfg = strings.ReplaceAll(scfg, "\t", "")
	scfg = strings.TrimSpace(scfg)
	cmds = append(cmds, "format", fmt.Sprintf("--configuration='%s'", scfg), "-")

	return NewExecFormatter(&BasicExternalFormatterOpts{
		Indent: "  ",
		// TempFiles: map[string]string{
		// 	"swift-config.json": swiftConfig,
		// },
	}, cmds...)
}
