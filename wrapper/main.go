package main

import (
	"github.com/unnamedtiger/check/common"
	"github.com/unnamedtiger/check/plugins/unwanted_imports"
)

func main() {
	common.Main(unwanted_imports.Plugin)
}
