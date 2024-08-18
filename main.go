package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
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

	directories := flag.Args()
	for _, dir := range directories {
		fullDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Printf("Unable to get absolute path for %s\n", dir)
			os.Exit(1)
		}
		fmt.Printf("- %s\n", fullDir)
	}
}
