package common

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func findJustification(n *sitter.Node, content []byte, name string) string {
	for {
		n = n.PrevNamedSibling()
		if n == nil {
			break
		}
		if n.Type() != "comment" {
			break
		}
		text := n.Content(content)
		idx := strings.Index(text, "JUSTIFY(")
		for idx >= 0 {
			text = text[idx:]
			text = strings.TrimPrefix(text, "JUSTIFY(")
			endIdx := strings.Index(text, ")")
			if endIdx < 0 {
				return ""
			}
			names := strings.Split(text[:endIdx], ",")
			for _, n := range names {
				if name == strings.TrimSpace(n) {
					text = text[endIdx:]
					text = strings.TrimPrefix(text, ")")
					text = strings.TrimPrefix(text, ":")
					lineIdx := strings.Index(text, "\n")
					if lineIdx < 0 {
						return strings.TrimSpace(text)
					}
					return strings.TrimSpace(text[:lineIdx])
				}
			}
			idx = strings.Index(text, "JUSTIFY(")
		}
	}
	return ""
}
