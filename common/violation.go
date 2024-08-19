package common

import (
	"encoding/json"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

const relevantContentBorder = uint32(1)

type Violation struct {
	PluginName string
	FilePath   string

	// all these are 0-indexed
	StartLine   uint32
	StartColumn uint32
	EndLine     uint32
	EndColumn   uint32

	ErrorCode string
	Message   string

	Justification *Justification

	RelevantContentStartLine uint32
	RelContent               []string
}

func newViolation(pluginName string, filePath string, n *sitter.Node, content []byte, errorCode string, message string) Violation {
	if n == nil {
		v := Violation{
			PluginName:               pluginName,
			FilePath:                 filePath,
			StartLine:                0,
			StartColumn:              0,
			EndLine:                  0,
			EndColumn:                0,
			ErrorCode:                errorCode,
			Message:                  message,
			Justification:            nil,
			RelevantContentStartLine: 0,
			RelContent:               []string{},
		}
		return v
	}

	tag := pluginName
	if errorCode != "" {
		tag += "/" + errorCode
	}
	just := findJustification(n, content, tag)

	startLine := n.StartPoint().Row
	if just != nil && startLine > just.StartLine {
		startLine = just.StartLine
	}
	if startLine >= relevantContentBorder {
		startLine -= relevantContentBorder
	} else {
		startLine = 0
	}
	endLine := n.EndPoint().Row
	endLine += relevantContentBorder
	relevantContent := collectContent(n, content, startLine, endLine)

	v := Violation{
		PluginName:               pluginName,
		FilePath:                 filePath,
		StartLine:                n.StartPoint().Row,
		StartColumn:              n.StartPoint().Column,
		EndLine:                  n.EndPoint().Row,
		EndColumn:                n.EndPoint().Column,
		ErrorCode:                errorCode,
		Message:                  message,
		Justification:            just,
		RelevantContentStartLine: startLine,
		RelContent:               relevantContent,
	}
	return v
}

func collectContent(n *sitter.Node, content []byte, startLine uint32, endLine uint32) []string {
	startByte := n.StartByte()
	endByte := n.EndByte()

	n2 := n
	for n2 != nil {
		row := n2.StartPoint().Row
		col := n2.StartPoint().Column
		if row == startLine && col == 0 {
			startByte = n2.StartByte()
			break
		}
		if row < startLine {
			startByte = n2.StartByte()
			for row < startLine {
				if content[startByte] == '\n' {
					row += 1
				}
				startByte += 1
			}
			break
		}
		if n2.PrevSibling() != nil {
			n2 = n2.PrevSibling()
		} else {
			n2 = n2.Parent()
		}
	}

	n2 = n
	for n2 != nil {
		endByte = n2.EndByte()
		row := n2.EndPoint().Row
		if row > endLine {
			endByte--
			for row > endLine {
				if content[endByte] == '\n' {
					row -= 1
				}
				endByte -= 1
			}
			if endByte+1 < uint32(len(content)) {
				endByte += 2
			}
			break
		}
		if n2.NextSibling() != nil {
			n2 = n2.NextSibling()
		} else {
			n2 = n2.Parent()
		}
	}

	full := string(content[startByte:endByte])
	r := strings.SplitAfter(full, "\n")
	return r[0 : len(r)-1]
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
	if v.Justification != nil {
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

	tag := v.PluginName
	if v.ErrorCode != "" {
		tag += "/" + v.ErrorCode
	}
	result := escBold
	if v.Justification == nil {
		result += escRed + "violation"
	} else {
		result += escCyan + "justified"
	}
	result += fmt.Sprintf("(%s)"+escReset+escBold+": %s"+escReset+"\n", tag, v.Message)
	if v.FilePath != "" {
		if v.StartLine == 0 && v.StartColumn == 0 && v.EndLine == 0 && v.EndColumn == 0 {
			result += fmt.Sprintf(escBlue+"  -->"+escReset+" %s\n", v.FilePath)
		} else {
			result += fmt.Sprintf(escBlue+"  -->"+escReset+" %s:%d:%d\n", v.FilePath, v.StartLine+1, v.StartColumn+1)
		}
	}
	lineNumberWidth := 0
	if len(v.RelContent) > 0 {
		lineNumberWidth = len(fmt.Sprintf("%d", v.EndLine+1))
		if lineNumberWidth < 2 {
			lineNumberWidth = 2
		}
		result += fmt.Sprintf(escBlue+"%*s |"+escReset+"\n", lineNumberWidth, "")

		lineNumber := v.RelevantContentStartLine + 1
		for _, line := range v.RelContent {
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
				if v.Justification == nil {
					l += escRed
				} else {
					l += escCyan
				}
				l += line[startChar:endChar]
				l += escReset
				l += line[endChar:]
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
				if v.Justification == nil {
					l += escRed
				} else {
					l += escCyan
				}
				l += underline + escReset + "\n"
				result += l
			} else {
				result += fmt.Sprintf(escBlue+"%*d | "+escReset+"%s", lineNumberWidth, lineNumber, line)
			}
			lineNumber++
		}
	}
	if v.Justification != nil {
		result += fmt.Sprintf(escBlue+"%*s = "+escReset+escBold+"justification:"+escReset+" %s\n", lineNumberWidth, "", v.Justification.Message)
	}
	return result
}
