package common

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// The public members of this struct are only set during Run, not during Finalize
type Analysis struct {
	Content   []byte
	Root      *sitter.Node
	FilePath  string
	Extension string

	pluginName string
	violations []Violation
}

func (a *Analysis) Report(n *sitter.Node, msg string) {
	a.ReportCode(n, "", msg)
}

func (a *Analysis) ReportCode(n *sitter.Node, errorCode string, msg string) {
	v := newViolation(a.pluginName, a.FilePath, n, a.Content, errorCode, msg)
	a.violations = append(a.violations, v)
}

func (a *Analysis) ReportCodef(n *sitter.Node, errorCode string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.ReportCode(n, errorCode, msg)
}

func (a *Analysis) Reportf(n *sitter.Node, format string, args ...any) {
	a.ReportCodef(n, "", format, args...)
}

func (a *Analysis) ReportFile(file string, msg string) {
	a.ReportFileCode(file, "", msg)
}

func (a *Analysis) ReportFileCode(file string, errorCode string, msg string) {
	v := newViolation(a.pluginName, file, nil, a.Content, errorCode, msg)
	a.violations = append(a.violations, v)
}

func (a *Analysis) ReportFileCodef(file string, errorCode string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.ReportFileCode(file, errorCode, msg)
}

func (a *Analysis) ReportFilef(file string, format string, args ...any) {
	a.ReportFileCodef(file, "", format, args...)
}
