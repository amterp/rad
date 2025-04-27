package core

import ts "github.com/tree-sitter/go-tree-sitter"

func ToStringArrayQuoteStr[T any](v []T, quoteStrings bool) []string {
	output := make([]string, len(v))
	for i, val := range v {
		output[i] = ToPrintableQuoteStr(val, quoteStrings)
	}
	return output
}

func GetSrc(src string, node *ts.Node) string {
	return src[node.StartByte():node.EndByte()]
}
