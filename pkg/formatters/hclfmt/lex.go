package hclfmt

import (
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// lexConfig uses the hclsyntax scanner to get a token stream and then
// rewrites it into this package's token model.
//
// Any errors produced during scanning are ignored, so the results of this
// function should be used with care.
func lexConfig(src []byte) Tokens {
	mainTokens, _ := hclsyntax.LexConfig(src, "", hcl.Pos{Byte: 0, Line: 1, Column: 1})

	mainTokens = injectNewlinesAndTrailingCommas(mainTokens)

	return writerTokens(mainTokens)
}

// this puts a newline after EACH bracket

// it also adds trailing commas where needed for arrays
// which wouldn't make much sense if things were on the same line,
// so this is coupled with the newline insertion
func injectNewlinesAndTrailingCommas(nativeTokens hclsyntax.Tokens) hclsyntax.Tokens {
	tnt := make([]hclsyntax.Token, 0)
	myline := []hclsyntax.Token{}
	prev := hclsyntax.Token{}
	lastNonNewline := hclsyntax.Token{
		Type: hclsyntax.TokenNil,
	}
	lastNewlineIdx := 0
	for _, nt := range nativeTokens {
		injectline := func() {
			tnt = append(tnt, hclsyntax.Token{
				Type:  hclsyntax.TokenNewline,
				Bytes: []byte("\n"),
				Range: nt.Range,
			})
			nt.Range.Start.Line++
			nt.Range.End.Line++

			myline = []hclsyntax.Token{}
		}

		injectTrailingComma := func() {

			injector := lastNewlineIdx

			tnt = slices.Insert(tnt, injector+1, hclsyntax.Token{
				Type:  hclsyntax.TokenComma,
				Bytes: []byte(","),
				Range: hcl.Range{
					Start: tnt[injector].Range.End,
					End:   tnt[injector].Range.End,
				},
			})

			lastNonNewline = hclsyntax.Token{
				Type: hclsyntax.TokenComma,
			}
		}

		switch {
		case prev.Type != hclsyntax.TokenNewline && (nt.Type == hclsyntax.TokenCBrack || nt.Type == hclsyntax.TokenCBrace):
			{
				injectline()
			}
		case (prev.Type == hclsyntax.TokenCBrack || prev.Type == hclsyntax.TokenCBrace || prev.Type == hclsyntax.TokenOBrace || prev.Type == hclsyntax.TokenOBrack) && nt.Type != hclsyntax.TokenNewline:
			{
				injectline()
			}
		}

		switch {
		case nt.Type == hclsyntax.TokenCBrack && (lastNonNewline.Type != hclsyntax.TokenComma):
			{
				injectTrailingComma()
			}
		}

		tnt = append(tnt, nt)
		myline = append(myline, nt)
		prev = nt

		if nt.Type != hclsyntax.TokenNewline {
			lastNonNewline = nt
			lastNewlineIdx = len(tnt) - 1
		}
	}

	return tnt
}

// writerTokens takes a sequence of tokens as produced by the main hclsyntax
// package and transforms it into an equivalent sequence of tokens using
// this package's own token model.
//
// The resulting list contain the same number of tokens and uses the same
// indices as the input, allowing the two sets of tokens to be correlated
// by index.
func writerTokens(nativeTokens hclsyntax.Tokens) Tokens {

	// Ultimately we want a slice of token _pointers_, but since we can
	// predict how much memory we're going to devote to tokens we'll allocate
	// it all as a single flat buffer and thus give the GC less work to do.
	tokBuf := make([]Token, len(nativeTokens))
	var lastByteOffset int
	for i, mainToken := range nativeTokens {
		// Create a copy of the bytes so that we can mutate without
		// corrupting the original token stream.
		bytes := make([]byte, len(mainToken.Bytes))
		copy(bytes, mainToken.Bytes)

		tokBuf[i] = Token{
			Token: hclwrite.Token{
				Type:  mainToken.Type,
				Bytes: bytes,

				// We assume here that spaces are always ASCII spaces, since
				// that's what the scanner also assumes, and thus the number
				// of bytes skipped is also the number of space characters.
				SpacesBefore: mainToken.Range.Start.Byte - lastByteOffset,
			},
			TabsBefore: 0,
		}

		lastByteOffset = mainToken.Range.End.Byte
	}

	// Now make a slice of pointers into the previous slice.
	ret := make(Tokens, len(tokBuf))
	for i := range ret {
		ret[i] = &tokBuf[i]
	}

	return ret
}
