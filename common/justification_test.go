package common

import (
	"context"
	"fmt"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func checkForJustifications(t *testing.T, code string, exp string) {
	content := []byte(code)
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		t.Fail()
	}
	nodes := FindNamedNodes(tree.RootNode(), "function_declaration")
	if len(nodes) != 1 {
		t.Fail()
	}
	act := findJustification(nodes[0], content, "test")
	if exp != act {
		fmt.Printf("exp: %v\n", exp)
		fmt.Printf("act: %v\n", act)
		t.Fail()
	}
}

func TestJustification(t *testing.T) {
	{
		code := "func main() {}"
		checkForJustifications(t, code, "")
	}
	{
		code := "// JUSTIFY(test): text\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(foo, test): text\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(foo,test): text\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(test, foo): text\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(test,foo): text\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(bar, test, foo): text\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(bar,test,foo): text\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(test): text\n\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(test): text\n// This is the main function\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(foo): hello\n// JUSTIFY(test): text\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(test): text\n// JUSTIFY(foo): hello\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(test): text\n\nconst foo = 42\n\nfunc main() {}"
		checkForJustifications(t, code, "")
	}
	{
		code := "/* JUSTIFY(test): text */\nfunc main() {}"
		checkForJustifications(t, code, "text */")
	}
	{
		code := "/*\n  JUSTIFY(test): text\n */\nfunc main() {}"
		checkForJustifications(t, code, "text")
	}
	{
		code := "// JUSTIFY(\nfunc main() {}"
		checkForJustifications(t, code, "")
	}
	{
		code := "// JUSTIFY(test\nfunc main() {}"
		checkForJustifications(t, code, "")
	}
	{
		code := "// JUSTIFY(test)\nfunc main() {}"
		checkForJustifications(t, code, "")
	}
}
