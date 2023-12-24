package hclwrite

// import (
// 	"github.com/hashicorp/hcl/v2"
// 	"github.com/hashicorp/hcl/v2/hclsyntax"
// 	"github.com/hashicorp/hcl/v2/hclwrite"
// 	"github.com/walteh/retab/pkg/configuration"
// )

// // lexConfig uses the hclsyntax scanner to get a token stream and then
// // rewrites it into this package's token model.
// //
// // Any errors produced during scanning are ignored, so the results of this
// // function should be used with care.
// func lexConfig(src []byte, cfg configuration.Configuration) Tokens {
// 	mainTokens, _ := hclsyntax.LexConfig(src, "", hcl.Pos{Byte: 0, Line: 1, Column: 1})
// 	return writerTokens(mainTokens, cfg)
// }

// // writerTokens takes a sequence of tokens as produced by the main hclsyntax
// // package and transforms it into an equivalent sequence of tokens using
// // this package's own token model.
// //
// // The resulting list contains the same number of tokens and uses the same
// // indices as the input, allowing the two sets of tokens to be correlated
// // by index.
// func writerTokens(nativeTokens hclsyntax.Tokens, cfg configuration.Configuration) Tokens {
// 	tmap := map[int]*hclsyntax.Token{}

// 	if cfg.OneBracketPerLine() {

// 		prev := hclsyntax.Token{}
// 		for i := range nativeTokens {
// 			ex := func() {
// 				tmap[i] = &hclsyntax.Token{
// 					Type:  hclsyntax.TokenNewline,
// 					Bytes: []byte("\n"),
// 					Range: nativeTokens[i].Range,
// 				}
// 				nativeTokens[i] = hclsyntax.Token{
// 					Type:  hclsyntax.TokenNewline,
// 					Bytes: nativeTokens[i].Bytes,
// 					Range: hcl.Range{
// 						Filename: nativeTokens[i].Range.Filename,
// 						Start: hcl.Pos{
// 							Line:   nativeTokens[i].Range.Start.Line + 1,
// 							Byte:   nativeTokens[i].Range.Start.Byte,
// 							Column: nativeTokens[i].Range.Start.Column,
// 						},
// 						End: hcl.Pos{
// 							Line:   nativeTokens[i].Range.Start.Line + 1,
// 							Byte:   nativeTokens[i].Range.Start.Byte,
// 							Column: nativeTokens[i].Range.Start.Column,
// 						},
// 					},
// 				}
// 			}
// 			switch {
// 			case prev.Type != hclsyntax.TokenNewline && (nativeTokens[i].Type == hclsyntax.TokenCBrack || nativeTokens[i].Type == hclsyntax.TokenCBrace):
// 				{
// 					ex()
// 				}
// 			case (prev.Type == hclsyntax.TokenCBrack || prev.Type == hclsyntax.TokenCBrace || prev.Type == hclsyntax.TokenOBrace || prev.Type == hclsyntax.TokenOBrack) && nativeTokens[i].Type != hclsyntax.TokenNewline:
// 				{
// 					ex()
// 				}
// 			}

// 			prev = nativeTokens[i]
// 		}
// 	}

// 	// Ultimately we want a slice of token _pointers_, but since we can
// 	// predict how much memory we're going to devote to tokens we'll allocate
// 	// it all as a single flat buffer and thus give the GC less work to do.
// 	tokBuf := make([]Token, len(nativeTokens)+len(tmap))
// 	var lastByteOffset int
// 	var numNewlines int
// 	var mainToken hclsyntax.Token
// 	for i := 0; i < len(nativeTokens)+len(tmap); i++ {
// 		if tmap[i] != nil {
// 			mainToken = *tmap[i]
// 			numNewlines++
// 		} else {
// 			mainToken = &nativeTokens[i-numNewlines]
// 		}
// 		// Create a copy of the bytes so that we can mutate without
// 		// corrupting the original token stream.
// 		bytes := make([]byte, len(mainToken.Bytes))
// 		copy(bytes, mainToken.Bytes)

// 		tokBuf[i] = Token{
// 			Token: hclwrite.Token{
// 				Type:  mainToken.Type,
// 				Bytes: bytes,

// 				// We assume here that spaces are always ASCII spaces, since
// 				// that's what the scanner also assumes, and thus the number
// 				// of bytes skipped is also the number of space characters.
// 				SpacesBefore: mainToken.Range.Start.Byte - lastByteOffset,
// 			},
// 			TabsBefore: 0,
// 		}

// 		lastByteOffset = mainToken.Range.End.Byte
// 	}

// 	// Now make a slice of pointers into the previous slice.
// 	ret := make(Tokens, len(tokBuf))
// 	for i := range ret {
// 		ret[i] = &tokBuf[i]
// 	}

// 	return ret
// }
