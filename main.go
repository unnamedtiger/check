package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func main() {
	version := flag.Bool("V", false, "print version and exit")

	flag.Parse()

	if version != nil && *version {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Printf("check (unknown)\n")
			os.Exit(0)
		}
		fmt.Printf("check %s %s\n", info.Main.Version, info.GoVersion)
		os.Exit(0)
	}

	plugins := []*Plugin{UnwantedImportsPlugin}
	for _, plugin := range plugins {
		for _, ext := range plugin.Extensions {
			if ext != "go" {
				fmt.Printf("unable to use plugin %s: unknown extension %s\n", plugin.Name, ext)
				os.Exit(1)
			}
		}
	}

	directories := flag.Args()
	for _, dir := range directories {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("unable to walk directory %s: %s\n", path, err)
				os.Exit(1)
			}

			if d.IsDir() {
				return nil
			}
			fmt.Printf("%s:\n", path)

			name := d.Name()
			ext := filepath.Ext(name)
			ext = strings.TrimPrefix(ext, ".")

			for _, plugin := range plugins {
				if plugin.handlesExt(ext) {
					fmt.Printf("- %s\n", plugin.Name)
					content, err := os.ReadFile(path)
					if err != nil {
						fmt.Printf("unable to read file %s: %s\n", path, err)
						os.Exit(1)
					}
					parser := sitter.NewParser()
					if ext == "go" {
						parser.SetLanguage(golang.GetLanguage())
					} else {
						fmt.Printf("unknown extension: %s\n", ext)
						os.Exit(1)
					}
					tree, err := parser.ParseCtx(context.Background(), nil, content)
					if err != nil {
						fmt.Printf("unable to parse file %s: %s\n", path, err)
						os.Exit(1)
					}
					root := tree.RootNode()

					a := &Analysis{
						Content: content,
						Root:    root,
					}

					err = plugin.Run(a)
					if err != nil {
						fmt.Printf("[%s] unable to check file %s: %s\n", plugin.Name, path, err)
						os.Exit(1)
					}

					for _, vio := range a.violations {
						fmt.Println(vio)
					}
				}
			}
			return nil
		})

		if err != nil {
			os.Exit(1)
		}
	}
}
