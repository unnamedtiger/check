package main

import (
	"fmt"
	"testing"
)

func TestViolationFormatters(t *testing.T) {
	{
		v := Violation{
			PluginName: "unwanted-imports",

			FilePath:    "test/test.go",
			StartLine:   4,
			StartColumn: 1,
			EndLine:     4,
			EndColumn:   12,

			ErrorCode: "E001",
			Message:   "contains unwanted import: io/ioutil",

			relevantContent: "\t\"io/ioutil\"",
		}
		exp := "violation(unwanted-imports/E001): contains unwanted import: io/ioutil\n  --> test/test.go:5:2\n   |\n 5 | 	\"io/ioutil\"\n   | \t^~~~~~~~~~~\n"
		if exp != v.String() {
			fmt.Printf("exp: %v\n", exp)
			fmt.Printf("v.String(): %v\n", v.String())
			t.Fail()
		}
	}
}
