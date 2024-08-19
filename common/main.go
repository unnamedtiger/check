package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func Main(plugins ...*Plugin) {
	// handling command line flags and parameters
	output := flag.String("o", "terminal", "output format [terminal, csv, json]")
	version := flag.Bool("V", false, "print version and exit")

	flag.Parse()
	directories := flag.Args()

	if version != nil && *version {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Printf("check (unknown)\n")
			os.Exit(0)
		}
		fmt.Printf("check %s %s\n", info.Main.Version, info.GoVersion)
		os.Exit(0)
	}

	if output != nil && *output != "terminal" && *output != "csv" && *output != "json" {
		fmt.Fprintf(os.Stderr, "invalid output format\n")
		os.Exit(2)
	}

	// checking that all plugins are usable in this tool
	for _, plugin := range plugins {
		for _, ext := range plugin.Extensions {
			if ext != "go" {
				fmt.Fprintf(os.Stderr, "unable to use plugin %s: unknown extension %s\n", plugin.Name, ext)
				os.Exit(2)
			}
		}
	}

	// looping over all directories and passing the files to the plugins
	violations := []Violation{}
	for _, dir := range directories {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to walk directory %s: %s\n", path, err)
				os.Exit(2)
			}

			if d.IsDir() {
				return nil
			}

			name := d.Name()
			ext := filepath.Ext(name)
			ext = strings.TrimPrefix(ext, ".")

			for _, plugin := range plugins {
				if plugin.handlesExtension(ext) {
					vios, err := makePluginHandleFile(plugin, path, ext)
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s\n", err)
						os.Exit(2)
					}
					violations = append(violations, vios...)
				}
			}
			return nil
		})

		if err != nil {
			os.Exit(2)
		}
	}

	// building up the report and outputting it
	report := Report{violations: violations}
	if output == nil || *output == "terminal" {
		for _, vio := range report.violations {
			fmt.Println(vio.StringPretty(true))
		}
	} else if *output == "csv" {
		var buf bytes.Buffer
		err := report.WriteCsv(&buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to write violations to csv: %s", err)
			os.Exit(2)
		}
		fmt.Print(buf.String())
	} else if *output == "json" {
		bytes, err := json.Marshal(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to write violations to json: %s", err)
			os.Exit(2)
		}
		fmt.Println(string(bytes))
	}

	// exit with correct code
	for _, vio := range report.violations {
		if vio.Justification == nil {
			os.Exit(1)
		}
	}
	os.Exit(0)
}

func parseFileContent(content []byte, ext string) (*sitter.Node, error) {
	parser := sitter.NewParser()
	if ext == "go" {
		parser.SetLanguage(golang.GetLanguage())
	} else {
		return nil, errors.New("unknown extension")
	}
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	root := tree.RootNode()
	return root, nil
}

func makePluginHandleFile(plugin *Plugin, path string, ext string) ([]Violation, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s: %s", path, err)
	}
	root, err := parseFileContent(content, ext)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file %s: %s", path, err)
	}

	a := &Analysis{
		Content: content,
		Root:    root,

		pluginName: plugin.Name,
		filePath:   path,
	}

	err = plugin.Run(a)
	if err != nil {
		return nil, fmt.Errorf("[%s] unable to check file %s: %s", plugin.Name, path, err)
	}
	return a.violations, nil
}
