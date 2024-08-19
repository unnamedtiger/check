package common

import sitter "github.com/smacker/go-tree-sitter"

func FindNamedNodes(n *sitter.Node, name string) []*sitter.Node {
	results := []*sitter.Node{}
	for i := uint32(0); i < n.NamedChildCount(); i++ {
		child := n.NamedChild(int(i))
		if child.Type() == name {
			results = append(results, child)
		}
		results = append(results, FindNamedNodes(n.NamedChild(int(i)), name)...)
	}
	return results
}
