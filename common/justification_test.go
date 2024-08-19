package common

import (
	"fmt"
	"testing"
)

func assertJustification(t *testing.T, exp Justification, act Justification) {
	if exp.StartLine != act.StartLine {
		t.Fail()
	}
	if exp.StartColumn != act.StartColumn {
		t.Fail()
	}
	if exp.EndLine != act.EndLine {
		t.Fail()
	}
	if exp.EndColumn != act.EndColumn {
		t.Fail()
	}
	if exp.Tag != act.Tag {
		t.Fail()
	}
	if exp.Message != act.Message {
		t.Fail()
	}
	if t.Failed() {
		fmt.Printf("exp: %#v\n", exp)
		fmt.Printf("act: %#v\n", act)
		t.FailNow()
	}
}

func checkForJustifications(t *testing.T, code string, tag string, exp *Justification) {
	content := []byte(code)
	root, err := parseFileContent(content, "go")
	if err != nil {
		t.Fail()
	}
	nodes := FindNamedNodes(root, "function_declaration")
	if len(nodes) != 1 {
		t.Fail()
	}
	act := findJustification(nodes[0], content, tag)
	if exp == nil && act == nil {
		// ok
	} else if exp != nil && act != nil {
		assertJustification(t, *exp, *act)
	} else {
		fmt.Printf("exp: %v\n", exp)
		fmt.Printf("act: %v\n", act)
		t.FailNow()
	}
}

func TestFindJustification(t *testing.T) {
	{
		code := "func main() {}"
		checkForJustifications(t, code, "test", nil)
	}
	{
		code := "// JUSTIFY(test): text\nfunc main() {}"
		checkForJustifications(t, code, "test", &Justification{0, 3, 0, 22, "test", "text"})
	}
	{
		code := "// JUSTIFY(test): text\n\nfunc main() {}"
		checkForJustifications(t, code, "test", &Justification{0, 3, 0, 22, "test", "text"})
	}
	{
		code := "// JUSTIFY(test): text\n// This is the main function\nfunc main() {}"
		checkForJustifications(t, code, "test", &Justification{0, 3, 0, 22, "test", "text"})
	}
	{
		code := "// This is the main function\n// JUSTIFY(test): text\nfunc main() {}"
		checkForJustifications(t, code, "test", &Justification{1, 3, 1, 22, "test", "text"})
	}
	{
		code := "// JUSTIFY(foo): hello\n// JUSTIFY(test): text\nfunc main() {}"
		checkForJustifications(t, code, "test", &Justification{1, 3, 1, 22, "test", "text"})
	}
	{
		code := "// JUSTIFY(test): text\n// JUSTIFY(foo): hello\nfunc main() {}"
		checkForJustifications(t, code, "test", &Justification{0, 3, 0, 22, "test", "text"})
	}
	{
		code := "// JUSTIFY(test): text\n\nconst foo = 42\n\nfunc main() {}"
		checkForJustifications(t, code, "test", nil)
	}
}

func TestExtractJustification(t *testing.T) {
	{
		j := ExtractJustifications("", 0, 0)
		if len(j) != 0 {
			t.Fail()
		}
	}
	{
		j := ExtractJustifications("JUSTIFY(", 0, 0)
		if len(j) != 0 {
			t.Fail()
		}
	}
	{
		j := ExtractJustifications("JUSTIFY()", 0, 0)
		if len(j) != 0 {
			t.Fail()
		}
	}
	{
		j := ExtractJustifications("JUSTIFY(foo", 0, 0)
		if len(j) != 0 {
			t.Fail()
		}
	}
	{
		j := ExtractJustifications("JUSTIFY(foo)", 0, 0)
		if len(j) != 0 {
			t.Fail()
		}
	}
	{
		j := ExtractJustifications("JUSTIFY(foo) message", 0, 0)
		if len(j) != 1 {
			t.Fail()
		}
		assertJustification(t, Justification{0, 0, 0, 20, "foo", "message"}, j[0])
	}
	{
		j := ExtractJustifications("JUSTIFY(foo): message", 0, 0)
		if len(j) != 1 {
			t.Fail()
		}
		assertJustification(t, Justification{0, 0, 0, 21, "foo", "message"}, j[0])
	}
	{
		j := ExtractJustifications("// JUSTIFY(foo): message", 0, 0)
		if len(j) != 1 {
			t.Fail()
		}
		assertJustification(t, Justification{0, 3, 0, 24, "foo", "message"}, j[0])
	}
	{
		j := ExtractJustifications("/* JUSTIFY(foo): message */", 0, 0)
		if len(j) != 1 {
			t.Fail()
		}
		assertJustification(t, Justification{0, 3, 0, 27, "foo", "message */"}, j[0])
	}
	{
		j := ExtractJustifications("/*\n * This is my function that does things\n *\n * JUSTIFY(foo): message\n */", 0, 0)
		if len(j) != 1 {
			t.Fail()
		}
		assertJustification(t, Justification{3, 3, 3, 24, "foo", "message"}, j[0])
	}
	{
		j := ExtractJustifications("JUSTIFY(foo,bar): message", 0, 0)
		if len(j) != 2 {
			t.Fail()
		}
		assertJustification(t, Justification{0, 0, 0, 25, "foo", "message"}, j[0])
		assertJustification(t, Justification{0, 0, 0, 25, "bar", "message"}, j[1])
	}
	{
		j := ExtractJustifications("JUSTIFY(foo, bar): message", 0, 0)
		if len(j) != 2 {
			t.Fail()
		}
		assertJustification(t, Justification{0, 0, 0, 26, "foo", "message"}, j[0])
		assertJustification(t, Justification{0, 0, 0, 26, "bar", "message"}, j[1])
	}
}
