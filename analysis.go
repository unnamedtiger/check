package main

import (
	"fmt"

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

	violations []Violation
}

func (a *Analysis) Reportf(n *sitter.Node, format string, args ...any) {
	vio := Violation{
		Message: fmt.Sprintf(format, args...),
	}
	a.violations = append(a.violations, vio)
}

type Violation struct {
	Message string
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
