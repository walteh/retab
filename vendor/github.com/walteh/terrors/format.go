package terrors

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func FileNameOfPath(path string) string {
	tot := strings.Split(path, "/")
	if len(tot) > 1 {
		return tot[len(tot)-1]
	}

	return path
}

func FormatCallerFromFrame(frm Frame) string {
	pkg, _, filestr, linestr := frm.Location()
	return FormatCaller(pkg, filestr, linestr)
}
func FormatCaller(pkg, path string, number int) string {
	pkgd := ColorBrackets("pkg", color.New(color.FgHiGreen).Sprint(pkg))
	pathd := ColorBrackets("file", fmt.Sprintf("%s:%s", color.New(color.Bold).Sprint(FileNameOfPath(path)), color.New(color.FgHiRed, color.Bold).Sprintf("%d", number)))
	return fmt.Sprintf("%s%s", pkgd, pathd)
}

func ColorBrackets(label string, value string) string {
	closeBracket := color.New(color.Faint, color.FgHiCyan).Sprint("]")
	openBracket := color.New(color.Faint, color.FgHiCyan).Sprint("[")
	return fmt.Sprintf("%s%s%s%s%s", openBracket, color.New(color.Faint, color.FgHiMagenta).Sprint(label), color.New(color.Faint, color.FgBlack).Sprint("="), value, closeBracket)
}

func ColorCode(code int) string {
	openBracket := color.New(color.Faint, color.FgHiRed).Sprint("{")
	closeBracket := color.New(color.Faint, color.FgHiRed).Sprint("}")
	return fmt.Sprintf("%s%s%s%s%s", openBracket, color.New(color.Faint, color.FgHiBlack).Sprint("code"), color.New(color.Faint, color.FgBlack).Sprint("="), color.New(color.FgHiRed, color.Bold).Sprint(code), closeBracket)
}

func ExtractErrorDetail(err error) string {
	if frm, ok := Cause2(err); ok {
		return frm.Detail()
	}

	return "no error detail found"
}

func FormatErrorCaller(err error, name string, verbose bool) string {
	// caller := ""
	dets := ""
	var errstr string
	if frm, ok := Cause2(err); ok {
		if verbose {
			errstr = frm.Simple()
			dets = frm.Detail()
		} else {
			errstr = frm.Error()
		}
	} else {
		errstr = err.Error()
	}

	if verbose {
		if dets != "" {
			dets = fmt.Sprintf("\n\n%s\n", dets)
		}
	} else {
		dets = ""
	}

	if name != "" {
		name = "[" + name + "] - "
	}

	return fmt.Sprintf("%s%s%s", name, color.New(color.FgRed).Sprint(errstr), dets)
}

func InlineChainFormatter(self func() string, kid error) string {

	if kid == nil {
		slf := self()
		if !strings.Contains(slf, "❌") {
			return "❌ " + slf
		}
		return slf
	}

	errd := kid.Error()

	arrow := "👉"

	if !strings.Contains(errd, arrow) && !strings.HasPrefix(errd, "❌") {
		arrow += " ❌"
	}

	return fmt.Sprintf("%s %s %s", self(), arrow, errd)
}

func FullChainFormatter(kid error) string {

	chain := GetChain(kid)

	wrk := "\n\n"

	for i, err := range chain {
		arrow := "👇"
		if len(chain)-1 == i {
			arrow = "❌"
		}
		wrk += arrow + " "
		switch v := err.(type) {
		case *wrapError:
			wrk += v.DetailedSelf()
		default:
			wrk += fmt.Sprintf("%s\n\n", v.Error())
		}
	}

	wrk += "\n\n"

	return wrk

}
