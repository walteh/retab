package hclwrite

// func trimNewlines(lines []formatLine) {
// 	if len(lines) == 0 {
// 		return nil
// 	}
// 	var start, end int
// 	for start = 0; start < len(lines); start++ {
// 		if tokens[start].Type != hclsyntax.TokenNewline {
// 			break
// 		}
// 	}
// 	for end = len(tokens); end > 0; end-- {
// 		if tokens[end-1].Type != hclsyntax.TokenNewline {
// 			break
// 		}
// 	}
// 	return tokens[start:end]
// }
