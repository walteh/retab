package hclwrite

import (
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
	return writerTokens(mainTokens)
}

// writerTokens takes a sequence of tokens as produced by the main hclsyntax
// package and transforms it into an equivalent sequence of tokens using
// this package's own token model.
//
// The resulting list contains the same number of tokens and uses the same
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
