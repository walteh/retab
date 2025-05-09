package swiftfmt

import (
	"fmt"
	"regexp"

	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/formatters/cmdfmt"
)

var wsrg = regexp.MustCompile(`\s+`)

func removeAllWhitespaceChars(s string) string {
	return wsrg.ReplaceAllString(s, "")
}

func rawSwiftCmd() []string {
	scfg := removeAllWhitespaceChars(externalSwiftFormatConfig)

	cmds := []string{"-", fmt.Sprintf("--configuration=%s", scfg)}
	return cmds
}

func NewSwiftCmdFormatter(opts ...cmdfmt.OptBasicExternalFormatterOptsSetter) format.Provider {
	cmds := rawSwiftCmd()

	startopts := []cmdfmt.OptBasicExternalFormatterOptsSetter{
		cmdfmt.WithIndent("  "),
		cmdfmt.WithExecutable("swift-format"),
		cmdfmt.WithDockerImageName("swift"),
		cmdfmt.WithDockerImageTag("6.1"),
	}

	return cmdfmt.NewFormatter(cmds, append(startopts, opts...)...)
}

// we could maybe read this in too, but we really need the indentation
// to be 2 spaces in the config passed to the command no matter what
// 2 spaces is arbitrary, but it's what we are passing in the Indent
// field of the BasicExternalFormatterOpts for swift
var externalSwiftFormatConfig = /* json */ `
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
