package common

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

type Analysis struct {
	Content []byte
	Root    *sitter.Node

	pluginName string
	filePath   string
	violations []Violation
}

func (a *Analysis) Reportf(n *sitter.Node, format string, args ...any) {
	startByte := n.StartByte() - n.StartPoint().Column
	endByte := n.EndByte()
	for len(a.Content) < int(endByte) && a.Content[endByte] != '\n' {
		endByte++
	}
	code := a.Content[startByte:endByte]

	just := findJustification(n, a.Content, a.pluginName)

	vio := Violation{
		PluginName: a.pluginName,

		FilePath: a.filePath,

		StartLine:   n.StartPoint().Row,
		StartColumn: n.StartPoint().Column,
		EndLine:     n.EndPoint().Row,
		EndColumn:   n.EndPoint().Column,

		ErrorCode: "",
		Message:   fmt.Sprintf(format, args...),

		Justification: just,

		relevantContent: string(code),
	}
	a.violations = append(a.violations, vio)
}
