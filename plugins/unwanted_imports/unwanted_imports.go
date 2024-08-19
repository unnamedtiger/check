package unwanted_imports

import (
	"strings"

	"github.com/unnamedtiger/check/common"
)

// NOTE: If any of these packages are imported it's an instant error
var unwanted_imports = []string{
	"io/ioutil", // As of Go 1.16 this package is deprecated. https://pkg.go.dev/io/ioutil
}

var Plugin = &common.Plugin{
	Name:       "unwanted-imports",
	Doc:        "reports imports of unwanted packages",
	Extensions: []string{"go"},
	Run:        run,
}

func run(a *common.Analysis) error {
	nodes := common.FindNamedNodes(a.Root, "import_spec")
	for _, importSpecNode := range nodes {
		pathNode := importSpecNode.Child(0)
		content := pathNode.Content(a.Content)
		content = strings.Trim(content, "\"")
		for _, unwanted := range unwanted_imports {
			if content == unwanted {
				a.Reportf(importSpecNode, "contains unwanted import: %s", unwanted)
				break
			}
		}
	}
	return nil
}
