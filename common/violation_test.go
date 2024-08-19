package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestViolationFormatters(t *testing.T) {
	{
		v := Violation{
			PluginName: "unwanted-imports",

			FilePath:    "test/test.go",
			StartLine:   4,
			StartColumn: 1,
			EndLine:     4,
			EndColumn:   12,

			ErrorCode: "E001",
			Message:   "contains unwanted import: io/ioutil",

			RelevantContentStartLine: 3,
			RelContent:               []string{"\t\"fmt\"\n", "\t\"io/ioutil\"\n", ")\n"},
		}
		exp := "violation(unwanted-imports/E001): contains unwanted import: io/ioutil\n  --> test/test.go:5:2\n   |\n 4 | \t\"fmt\"\n 5 | \t\"io/ioutil\"\n   | \t^~~~~~~~~~~\n 6 | )\n"
		if exp != v.String() {
			fmt.Printf("exp: %v\n", exp)
			fmt.Printf("v.String(): %v\n", v.String())
			t.Fail()
		}

		csvExp := "unwanted-imports,test/test.go,4,1,4,12,E001,contains unwanted import: io/ioutil,\n"
		var buf bytes.Buffer
		r := Report{violations: []Violation{v}}
		err := r.WriteCsv(&buf)
		if err != nil {
			t.Fail()
		}
		if csvExp != buf.String() {
			fmt.Printf("csvExp: %v\n", csvExp)
			fmt.Printf("buf.String(): %v\n", buf.String())
			t.Fail()
		}

		jsonExp := `{"test/test.go":[{"code":"E001","message":"contains unwanted import: io/ioutil","range":{"end":{"character":12,"line":4},"start":{"character":1,"line":4}},"severity":1,"source":"unwanted-imports"}]}`
		data, err := json.Marshal(r)
		if err != nil {
			t.Fail()
		}
		if jsonExp != string(data) {
			fmt.Printf("jsonExp: %v\n", jsonExp)
			fmt.Printf("string(data): %v\n", string(data))
			t.Fail()
		}
	}
	{
		v := Violation{
			PluginName: "unwanted-imports",

			FilePath:    "test/test.go",
			StartLine:   4,
			StartColumn: 4,
			EndLine:     4,
			EndColumn:   15,

			ErrorCode: "E001",
			Message:   "contains unwanted import: io/ioutil",

			Justification: &Justification{3, 7, 3, 67, "unwanted-imports/E001", "it's okay this time, I swear"},

			RelevantContentStartLine: 3,
			RelContent:               []string{"    // JUSTIFY(unwanted-imports/E001): it's okay this time, I swear\n", "    \"io/ioutil\"\n", ")\n"},
		}
		exp := "justified(unwanted-imports/E001): contains unwanted import: io/ioutil\n  --> test/test.go:5:5\n   |\n 4 |     // JUSTIFY(unwanted-imports/E001): it's okay this time, I swear\n 5 |     \"io/ioutil\"\n   |     ^~~~~~~~~~~\n 6 | )\n   = justification: it's okay this time, I swear\n"
		if exp != v.String() {
			fmt.Printf("exp: %v\n", exp)
			fmt.Printf("v.String(): %v\n", v.String())
			t.Fail()
		}

		csvExp := "unwanted-imports,test/test.go,4,4,4,15,E001,contains unwanted import: io/ioutil,\"it's okay this time, I swear\"\n"
		var buf bytes.Buffer
		r := Report{violations: []Violation{v}}
		err := r.WriteCsv(&buf)
		if err != nil {
			t.Fail()
		}
		if csvExp != buf.String() {
			fmt.Printf("csvExp: %v\n", csvExp)
			fmt.Printf("buf.String(): %v\n", buf.String())
			t.Fail()
		}

		jsonExp := `{"test/test.go":[{"code":"E001","message":"contains unwanted import: io/ioutil","range":{"end":{"character":15,"line":4},"start":{"character":4,"line":4}},"severity":3,"source":"unwanted-imports"}]}`
		data, err := json.Marshal(r)
		if err != nil {
			t.Fail()
		}
		if jsonExp != string(data) {
			fmt.Printf("jsonExp: %v\n", jsonExp)
			fmt.Printf("string(data): %v\n", string(data))
			t.Fail()
		}
	}
	{
		v := Violation{
			PluginName: "unwanted-imports",

			FilePath:    "",
			StartLine:   0,
			StartColumn: 0,
			EndLine:     0,
			EndColumn:   0,

			ErrorCode: "E001",
			Message:   "global catastrophe",

			Justification: nil,

			RelevantContentStartLine: 0,
			RelContent:               []string{},
		}
		exp := "violation(unwanted-imports/E001): global catastrophe\n"
		if exp != v.String() {
			fmt.Printf("exp: %v\n", exp)
			fmt.Printf("v.String(): %v\n", v.String())
			t.Fail()
		}

		csvExp := "unwanted-imports,,0,0,0,0,E001,global catastrophe,\n"
		var buf bytes.Buffer
		r := Report{violations: []Violation{v}}
		err := r.WriteCsv(&buf)
		if err != nil {
			t.Fail()
		}
		if csvExp != buf.String() {
			fmt.Printf("csvExp: %v\n", csvExp)
			fmt.Printf("buf.String(): %v\n", buf.String())
			t.Fail()
		}

		jsonExp := `{"":[{"code":"E001","message":"global catastrophe","range":{"end":{"character":0,"line":0},"start":{"character":0,"line":0}},"severity":1,"source":"unwanted-imports"}]}`
		data, err := json.Marshal(r)
		if err != nil {
			t.Fail()
		}
		if jsonExp != string(data) {
			fmt.Printf("jsonExp: %v\n", jsonExp)
			fmt.Printf("string(data): %v\n", string(data))
			t.Fail()
		}
	}
}

func checkCollectContent(t *testing.T, pre string, lines []string, post string, start uint32, end uint32) {
	code := pre + strings.Join(lines, "") + post
	content := []byte(code)
	root, err := parseFileContent(content, "go")
	if err != nil {
		t.Fail()
	}
	nodes := FindNamedNodes(root, "var_declaration")
	if len(nodes) != 1 {
		t.Fail()
	}
	result := collectContent(nodes[0], content, start, end)
	for i := 0; i < min(len(lines), len(result)); i++ {
		if lines[i] != result[i] {
			fmt.Printf("lines[%d]: %#v\n", i, lines[i])
			fmt.Printf("result[%d]: %#v\n", i, result[i])
			t.Fail()
		}
	}
	if len(lines) != len(result) {
		for i := min(len(lines), len(result)); i < len(lines); i++ {
			fmt.Printf("lines[%d]: %#v\n", i, lines[i])
		}
		for i := min(len(lines), len(result)); i < len(result); i++ {
			fmt.Printf("result[%d]: %#v\n", i, result[i])
		}
		fmt.Printf("len(lines): %#v\n", len(lines))
		fmt.Printf("len(result): %#v\n", len(result))
		t.Fail()
	}
	if t.Failed() {
		t.FailNow()
	}
}

func TestCollectContent(t *testing.T) {
	checkCollectContent(t, "", []string{"var foo bool\n"}, "", 0, 0)
	checkCollectContent(t, "", []string{"var foo bool\n", "// 2\n"}, "", 0, 1)
	checkCollectContent(t, "", []string{"// 1\n", "var foo bool\n", "// 2\n"}, "", 0, 2)
	checkCollectContent(t, "/* 1 -\n", []string{"- 1 */\n", "var foo bool\n", "/* 2 -\n"}, "- 2 */ /* 3 -\n- 3 */\n", 1, 3)
	checkCollectContent(t, "/* 1 -\n", []string{"- 1 */\n", "var foo bool\n", ")\n"}, "", 1, 3)
	checkCollectContent(t, "", []string{"func main() {\n", "var foo bool\n", "}\n"}, "", 0, 2)
	checkCollectContent(t, "func main() {\n", []string{"// 1\n", "var foo bool\n", "}\n"}, "", 1, 3)
	checkCollectContent(t, "", []string{"func main() {\n", "var foo bool\n", "// 2\n"}, "}\n", 0, 2)
}
