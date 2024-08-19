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

func (a *Analysis) Report(n *sitter.Node, msg string) {
	a.ReportCode(n, "", msg)
}

func (a *Analysis) ReportCode(n *sitter.Node, errorCode string, msg string) {
	v := newViolation(a.pluginName, a.filePath, n, a.Content, errorCode, msg)
	a.violations = append(a.violations, v)
}

func (a *Analysis) ReportCodef(n *sitter.Node, errorCode string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.ReportCode(n, errorCode, msg)
}

func (a *Analysis) Reportf(n *sitter.Node, format string, args ...any) {
	a.ReportCodef(n, "", format, args...)
}
