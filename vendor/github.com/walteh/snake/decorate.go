//	decorate.go
//
//	created by walteh on 2023-08-18
//	Copyright © 2023 Walter Scott <w@lter.ca>. All rights reserved.
//
// ---------------------------------------------------------------------
//
//	adapted from ivanpirog/coloredcobra
//	Copyright © 2022 Ivan Pirog <ivan.pirog@gmail.com>. MIT license
//
// ---------------------------------------------------------------------
//

package snake

import (
	"context"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type DecorateOptions struct {
	Headings        *color.Color
	Commands        *color.Color
	CmdShortDescr   *color.Color
	ExecName        *color.Color
	Flags           *color.Color
	FlagsDataType   *color.Color
	FlagsDescr      *color.Color
	Aliases         *color.Color
	Example         *color.Color
	NoExtraNewlines bool
	NoBottomNewline bool
}

// Init patches Cobra's usage template with configuration provided.
func DecorateTemplate(ctx context.Context, root *cobra.Command, cfg *DecorateOptions) (string, error) {

	if root == nil || cfg == nil {
		return "", ErrInvalidArguments
	}

	// Get usage template
	tpl := root.UsageTemplate()

	//
	// Add extra line breaks for headings
	//
	if cfg.NoExtraNewlines == false {
		tpl = strings.NewReplacer(
			"Usage:", "\nUsage:\n",
			"Aliases:", "\nAliases:\n",
			"Examples:", "\nExamples:\n",
			"Available Commands:", "\nAvailable Commands:\n",
			"Global Flags:", "\nGlobal Flags:\n",
			"Additional help topics:", "\nAdditional help topics:\n",
			"Use \"", "\nUse \"",
		).Replace(tpl)
		re := regexp.MustCompile(`(?m)^Flags:$`)
		tpl = re.ReplaceAllString(tpl, "\nFlags:\n")
	}

	//
	// Styling headers
	//
	if cfg.Headings != nil {

		// Add template function to style the headers
		cobra.AddTemplateFunc("HeadingStyle", cfg.Headings.SprintFunc())

		// Wrap template headers into a new function
		tpl = strings.NewReplacer(
			"Usage:", `{{HeadingStyle "Usage:"}}`,
			"Aliases:", `{{HeadingStyle "Aliases:"}}`,
			"Examples:", `{{HeadingStyle "Examples:"}}`,
			"Available Commands:", `{{HeadingStyle "Available Commands:"}}`,
			"Global Flags:", `{{HeadingStyle "Global Flags:"}}`,
			"Additional help topics:", `{{HeadingStyle "Additional help topics:"}}`,
		).Replace(tpl)

		re := regexp.MustCompile(`(?m)^(\s*)Flags:(\s*)$`)
		tpl = re.ReplaceAllString(tpl, `$1{{HeadingStyle "Flags:"}}$2`)
	}

	//
	// Styling commands
	//
	if cfg.Commands != nil {
		cc := cfg.Commands

		// Add template function to style commands
		cobra.AddTemplateFunc("CommandStyle", cc.SprintFunc())
		cobra.AddTemplateFunc("sum", func(a, b int) int {
			return a + b
		})

		// Patch usage template
		re := regexp.MustCompile(`(?i){{\s*rpad\s+.Name\s+.NamePadding\s*}}`)
		tpl = re.ReplaceAllLiteralString(tpl, "{{rpad (CommandStyle .Name) (sum .NamePadding 12)}}")

		re = regexp.MustCompile(`(?i){{\s*rpad\s+.CommandPath\s+.CommandPathPadding\s*}}`)
		tpl = re.ReplaceAllLiteralString(tpl, "{{rpad (CommandStyle .CommandPath) (sum .CommandPathPadding 12)}}")
	}

	//
	// Styling a short desription of commands
	//
	if cfg.CmdShortDescr != nil {
		csd := cfg.CmdShortDescr

		cobra.AddTemplateFunc("CmdShortStyle", csd.SprintFunc())

		re := regexp.MustCompile(`(?ism)({{\s*range\s+.Commands\s*}}.*?){{\s*.Short\s*}}`)
		tpl = re.ReplaceAllString(tpl, `$1{{CmdShortStyle .Short}}`)
	}

	//
	// Styling executable file name
	//
	if cfg.ExecName != nil {
		cen := cfg.ExecName

		// Add template functions
		cobra.AddTemplateFunc("ExecStyle", cen.SprintFunc())
		cobra.AddTemplateFunc("UseLineStyle", func(s string) string {
			spl := strings.Split(s, " ")
			spl[0] = cen.Sprint(spl[0])
			return strings.Join(spl, " ")
		})

		// Patch usage template
		re := regexp.MustCompile(`(?i){{\s*.CommandPath\s*}}`)
		tpl = re.ReplaceAllLiteralString(tpl, "{{ExecStyle .CommandPath}}")

		re = regexp.MustCompile(`(?i){{\s*.UseLine\s*}}`)
		tpl = re.ReplaceAllLiteralString(tpl, "{{UseLineStyle .UseLine}}")
	}

	//
	// Styling flags
	//
	var cf, cfd, cfdt *color.Color

	cf = cfg.Flags

	cfd = cfg.FlagsDescr

	cfdt = cfg.FlagsDataType

	if cf != nil || cfd != nil || cfdt != nil {

		cobra.AddTemplateFunc("FlagStyle", func(s string) string {

			// Flags info section is multi-line.
			// Let's split these lines and iterate them.
			lines := strings.Split(s, "\n")
			for k := range lines {

				// Styling short and full flags (-f, --flag)
				if cf != nil {
					re := regexp.MustCompile(`(--?\S+)`)
					for _, flag := range re.FindAllString(lines[k], 2) {
						lines[k] = strings.Replace(lines[k], flag, cf.Sprint(flag), 1)
					}
				}

				// If no styles for flag data types and description - continue
				if cfd == nil && cfdt == nil {
					continue
				}

				// Split line into two parts: flag data type and description
				// Tip: Use debugger to understand the logic
				re := regexp.MustCompile(`\s{2,}`)
				spl := re.Split(lines[k], -1)
				if len(spl) != 3 {
					continue
				}

				// Styling the flag description
				if cfd != nil {
					lines[k] = strings.Replace(lines[k], spl[2], cfd.Sprint(spl[2]), 1)
				}

				// Styling flag data type
				// Tip: Use debugger to understand the logic
				if cfdt != nil {
					re = regexp.MustCompile(`\s+(\w+)$`) // the last word after spaces is the flag data type
					m := re.FindAllStringSubmatch(spl[1], -1)
					if len(m) == 1 && len(m[0]) == 2 {
						lines[k] = strings.Replace(lines[k], m[0][1], cfdt.Sprint(m[0][1]), 1)
					}
				}

			}
			s = strings.Join(lines, "\n")

			return s

		})

		// Patch usage template
		re := regexp.MustCompile(`(?i)(\.(InheritedFlags|LocalFlags)\.FlagUsages)`)
		tpl = re.ReplaceAllString(tpl, "FlagStyle $1")
	}

	//
	// Styling aliases
	//
	if cfg.Aliases != nil {
		ca := cfg.Aliases
		cobra.AddTemplateFunc("AliasStyle", ca.SprintFunc())

		re := regexp.MustCompile(`(?i){{\s*.NameAndAliases\s*}}`)
		tpl = re.ReplaceAllLiteralString(tpl, "{{AliasStyle .NameAndAliases}}")
	}

	//
	// Styling the example text
	//
	if cfg.Example != nil {
		ce := cfg.Example
		cobra.AddTemplateFunc("ExampleStyle", ce.SprintFunc())

		re := regexp.MustCompile(`(?i){{\s*.Example\s*}}`)
		tpl = re.ReplaceAllLiteralString(tpl, "{{ExampleStyle .Example}}")
	}

	// Adding a new line to the end
	if !cfg.NoBottomNewline {
		tpl += "\n"
	}

	return tpl, nil
}
