package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type Plugin struct {
	Name       string
	Doc        string
	Extensions []string
	Run        func(analysis *Analysis) error
}

func (p *Plugin) handlesExt(ext string) bool {
	for _, e := range p.Extensions {
		if e == ext {
			return true
		}
	}
	return false
}

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

type Violation struct {
	PluginName string

	FilePath string

	// all these are 0-indexed
	StartLine   uint32
	StartColumn uint32
	EndLine     uint32
	EndColumn   uint32

	ErrorCode string
	Message   string

	Justification string

	// starts at StartLine:0 and ends at end of EndLine
	relevantContent string
}

// NOTE: this follows https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#diagnostic
func (v Violation) MarshalJSON() ([]byte, error) {
	start := map[string]uint32{}
	start["line"] = v.StartLine
	start["character"] = v.StartColumn
	end := map[string]uint32{}
	end["line"] = v.EndLine
	end["character"] = v.EndColumn
	pos := map[string]interface{}{}
	pos["start"] = start
	pos["end"] = end
	m := map[string]interface{}{}
	m["range"] = pos
	if v.ErrorCode != "" {
		m["code"] = v.ErrorCode
	}
	m["source"] = v.PluginName
	m["message"] = v.Message
	if v.Justification != "" {
		m["severity"] = 3 // informational
	} else {
		m["severity"] = 1 // error
	}
	return json.Marshal(m)
}

func (v Violation) String() string {
	return v.StringPretty(false)
}

func (v Violation) StringPretty(color bool) string {
	escReset := ""
	escBold := ""
	escRed := ""
	escBlue := ""
	escCyan := ""
	if color {
		escReset = "\x1b[0m"
		escBold = "\x1b[1m"
		escRed = "\x1b[91m"
		escBlue = "\x1b[94m"
		escCyan = "\x1b[96m"
	}

	t := v.PluginName
	if v.ErrorCode != "" {
		t += "/" + v.ErrorCode
	}
	result := escBold
	if v.Justification == "" {
		result += escRed + "violation"
	} else {
		result += escCyan + "justified"
	}
	result += fmt.Sprintf("(%s)"+escReset+escBold+": %s"+escReset+"\n", t, v.Message)
	result += fmt.Sprintf(escBlue+"  -->"+escReset+" %s:%d:%d\n", v.FilePath, v.StartLine+1, v.StartColumn+1)
	lineNumberWidth := len(fmt.Sprintf("%d", v.EndLine+1))
	if lineNumberWidth < 2 {
		lineNumberWidth = 2
	}
	result += fmt.Sprintf(escBlue+"%*s |"+escReset+"\n", lineNumberWidth, "")

	lines := strings.Split(v.relevantContent, "\n")
	lineNumber := v.StartLine + 1
	for _, line := range lines {
		if v.StartLine+1 <= lineNumber && lineNumber <= v.EndLine+1 && len(line) > 0 {
			startChar := uint32(0)
			endChar := uint32(len(line) - 1)
			if v.StartLine+1 == lineNumber {
				startChar = v.StartColumn
			}
			if v.EndLine+1 == lineNumber {
				endChar = v.EndColumn
			}

			l := fmt.Sprintf(escBlue+"%*d | "+escReset+"", lineNumberWidth, lineNumber)
			l += line[0:startChar]
			if v.Justification == "" {
				l += escRed
			} else {
				l += escCyan
			}
			l += line[startChar:endChar]
			l += escReset
			l += line[endChar:]
			l += "\n"
			result += l

			underline := strings.Repeat("~", int(endChar-startChar))
			if v.StartLine+1 == lineNumber && len(underline) > 0 {
				underline = "^" + underline[1:]
			}
			l = fmt.Sprintf(escBlue+"%*s | "+escReset, lineNumberWidth, "")
			for i := 0; i < int(startChar); i++ {
				c := line[i]
				if c == '\t' {
					l += "\t"
				} else {
					l += " "
				}
			}
			if v.Justification == "" {
				l += escRed
			} else {
				l += escCyan
			}
			l += underline + escReset + "\n"
			result += l
		} else {
			result += fmt.Sprintf(escBlue+"%*d | "+escReset+"%s\n", lineNumberWidth, lineNumber, line)
		}
		lineNumber++
	}
	if v.Justification != "" {
		result += fmt.Sprintf(escBlue+"%*s = "+escReset+escBold+"justification:"+escReset+" %s\n", lineNumberWidth, "", v.Justification)
	}
	return result
}

type Report struct {
	Violations []Violation
}

func (r Report) MarshalJSON() ([]byte, error) {
	filePathMap := map[string][]Violation{}
	for _, vio := range r.Violations {
		_, ok := filePathMap[vio.FilePath]
		if !ok {
			filePathMap[vio.FilePath] = []Violation{}
		}
		filePathMap[vio.FilePath] = append(filePathMap[vio.FilePath], vio)
	}
	return json.Marshal(filePathMap)
}

func (r Report) WriteCsv(parentWriter io.Writer) error {
	w := csv.NewWriter(parentWriter)
	for _, vio := range r.Violations {
		records := []string{vio.PluginName, vio.FilePath, fmt.Sprintf("%d", vio.StartLine), fmt.Sprintf("%d", vio.StartColumn), fmt.Sprintf("%d", vio.EndLine), fmt.Sprintf("%d", vio.EndColumn), vio.ErrorCode, vio.Message, vio.Justification}
		err := w.Write(records)
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}

func findNamedNodes(n *sitter.Node, name string) []*sitter.Node {
	results := []*sitter.Node{}
	for i := uint32(0); i < n.NamedChildCount(); i++ {
		child := n.NamedChild(int(i))
		if child.Type() == name {
			results = append(results, child)
		}
		results = append(results, findNamedNodes(n.NamedChild(int(i)), name)...)
	}
	return results
}
