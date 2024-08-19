module github.com/unnamedtiger/check/wrapper

go 1.22.4

require (
	github.com/smacker/go-tree-sitter v0.0.0-20240625050157-a31a98a7c0f6
	github.com/unnamedtiger/check/common v0.0.0
	github.com/unnamedtiger/check/plugins/unwanted_imports v0.0.0
)

replace (
	github.com/unnamedtiger/check/common => ../common
	github.com/unnamedtiger/check/plugins/unwanted_imports => ../plugins/unwanted_imports
)
