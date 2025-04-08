package diff

import (
	"sync"

	"github.com/fatih/color"
)

type ColoredRune struct {
	Rune      rune
	Suffix    string
	Color     *color.Color
	IsSpecial bool
}

func (cr *ColoredRune) MarkIsSpecial() {
	cr.IsSpecial = true
}

func (cr *ColoredRune) Add(c ...color.Attribute) {
	cr.Color = cr.Color.Add(c...)
}

func cr(r rune, c ...color.Attribute) *ColoredRune {
	return &ColoredRune{Rune: r, Color: color.New(c...)}
}

func crs(r string, c ...color.Attribute) []*ColoredRune {
	out := []*ColoredRune{}
	for _, r := range r {
		out = append(out, cr(r, c...))
	}
	return out
}

type ColoredString struct {
	runes     []*ColoredRune
	index     int
	indexLock sync.Mutex
}

func (cs *ColoredString) AppendToEnd(r rune, c ...color.Attribute) {
	cs.indexLock.Lock()
	defer cs.indexLock.Unlock()
	cs.runes = append(cs.runes, &ColoredRune{Rune: r, Color: color.New(c...)})
	cs.index++
}

func (cs *ColoredString) AppendToStart(r rune, c ...color.Attribute) {
	cs.indexLock.Lock()
	defer cs.indexLock.Unlock()
	crd := ColoredRune{Rune: r, Color: color.New(c...)}
	cs.runes = append([]*ColoredRune{&crd}, cs.runes...)
	cs.index++
}

func (cs *ColoredString) MultiAppendToStart(rs ...*ColoredRune) {
	cs.indexLock.Lock()
	defer cs.indexLock.Unlock()
	cs.runes = append(rs, cs.runes...)
	cs.index += len(rs)
}

func (cs *ColoredString) MultiAppendToEnd(rs ...*ColoredRune) {
	cs.indexLock.Lock()
	defer cs.indexLock.Unlock()
	cs.runes = append(cs.runes, rs...)
	cs.index += len(rs)
}
func (s *ColoredString) Annotate(colord *color.Color) {

	for _, r := range s.runes {
		switch r.Rune {
		case ' ':
			r.Rune = '∙'
			if !r.IsSpecial {
				r.Add(color.Faint, color.FgHiYellow)
			}
		case '\t':
			r.Rune = '→'
			r.Suffix = "   "
			if !r.IsSpecial {
				r.Add(color.Faint, color.FgHiYellow)
			}
		default:
			r.Add(color.Italic)

		}
	}
	s.MultiAppendToStart(cr(' '), cr('|', color.Bold), cr(' '))

	// return applyWhitespaceColor(s, colord)
}

func NewColoredString(s string) *ColoredString {
	str := &ColoredString{runes: []*ColoredRune{}}
	for _, r := range s {
		str.AppendToEnd(r)
	}
	return str
}

func (cs *ColoredString) ColoredString() string {
	cs.indexLock.Lock()
	defer cs.indexLock.Unlock()
	out := ""
	for _, r := range cs.runes {
		out += r.Color.Sprint(string(r.Rune))
		if r.Suffix != "" {
			out += r.Color.Sprint(r.Suffix)
		}
	}
	return out
}

// func applyWhitespaceColor(s *ColoredString, colord *color.Color) string {
// 	out := ""
// 	for j, char := range s {
// 		switch char {
// 		case ' ':
// 			out += color.New(color.Faint).Sprint("∙") // ⌷
// 		case '\t':
// 			out += color.New(color.Faint).Sprint("→   ") // → └──▹
// 		default:
// 			wrk := s
// 			trailing := getFormattedTrailingWhitespace(wrk[j:])
// 			fomrmattedTrail := color.New(color.Faint).Sprint(string(trailing))
// 			return out + formatInternalWhitespace(wrk[j:len(wrk)-len(trailing)], colord) + fomrmattedTrail
// 		}
// 	}
// 	return out
// }

// func formatInternalWhitespace(s string, colord *color.Color) string {
// 	out := ""
// 	for _, char := range s {
// 		switch char {
// 		case ' ':
// 			out += color.New(color.Faint).Sprint("∙") // ⌷
// 		case '\t':
// 			out += color.New(color.Faint).Sprint("→   ") // → └──▹
// 		default:
// 			// if colord == nil {
// 			// 	out += string(char)
// 			// } else {
// 			out += colord.Sprint(string(char))
// 			// }
// 		}
// 	}
// 	return out
// }

// func getFormattedTrailingWhitespace(s string) []rune {
// 	out := []rune{}
// 	rstr := []rune(s)
// 	slices.Reverse(rstr)
// 	for _, char := range rstr {
// 		switch char {
// 		case ' ':
// 			out = append(out, '∙')
// 		case '\t':
// 			out = append(out, '→')
// 		// case '\n':
// 		// 	out += color.New(color.Faint, color.FgHiGreen).Sprint("↵") // ↵
// 		default:
// 			return out
// 		}
// 	}
// 	return out
// }
