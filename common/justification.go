package common

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type Justification struct {
	// all these are 0-indexed
	StartLine   uint32
	StartColumn uint32
	EndLine     uint32
	EndColumn   uint32

	Tag     string
	Message string
}

func findJustification(n *sitter.Node, content []byte, tag string) *Justification {
	for {
		n = n.PrevNamedSibling()
		if n == nil {
			break
		}
		if n.Type() != "comment" {
			break
		}

		text := n.Content(content)
		startLine := n.StartPoint().Row
		startColumn := n.StartPoint().Column
		justifications := ExtractJustifications(text, startLine, startColumn)
		for _, j := range justifications {
			if j.Tag == tag {
				return &j
			}
		}
	}
	return nil
}

func ExtractJustifications(text string, startLine uint32, startColumn uint32) []Justification {
	justifications := []Justification{}

	for len(text) > 0 {
		if strings.HasPrefix(text, "JUSTIFY(") {
			beginLen := len(text)
			text = strings.TrimPrefix(text, "JUSTIFY(")
			idx := strings.Index(text, ")")
			if idx >= 0 {
				tags := text[:idx]
				text := text[idx:]
				text = strings.TrimPrefix(text, ")")
				text = strings.TrimPrefix(text, ":")
				idx = strings.Index(text, "\n")
				var msg string
				if idx >= 0 {
					msg = strings.TrimSpace(text[:idx])
					text = text[idx:]
				} else {
					msg = strings.TrimSpace(text)
					text = ""
				}
				if len(msg) > 0 {
					parts := strings.Split(tags, ",")
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if len(p) > 0 {
							j := Justification{
								StartLine:   startLine,
								StartColumn: startColumn,
								EndLine:     startLine,
								EndColumn:   startColumn + uint32(beginLen-len(text)),
								Tag:         p,
								Message:     msg,
							}
							justifications = append(justifications, j)
						}
					}
				}
			} else {
				startColumn += 8 // len("JUSTIFY(")
			}
		}

		if len(text) == 0 {
			return justifications
		}
		char := text[0]
		text = text[1:]
		if char == '\n' {
			startLine += 1
			startColumn = 0
		} else {
			startColumn += 1
		}
	}

	return justifications
}
