package main

import (
	"bytes"
	"encoding/json"
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

		csvExp := "unwanted-imports,test/test.go,4,1,4,12,E001,contains unwanted import: io/ioutil\n"
		var buf bytes.Buffer
		r := Report{Violations: []Violation{v}}
		err := r.WriteCsv(&buf)
		if err != nil {
			t.Fail()
		}
		if csvExp != buf.String() {
			fmt.Printf("csvExp: %v\n", csvExp)
			fmt.Printf("buf.String(): %v\n", buf.String())
			t.Fail()
		}

		jsonExp := `{"test/test.go":[{"code":"E001","message":"contains unwanted import: io/ioutil","range":{"end":{"character":12,"line":4},"start":{"character":1,"line":4}},"source":"unwanted-imports"}]}`
		data, err := json.Marshal(r)
		if err != nil {
			t.Fail()
		}
		if jsonExp != string(data) {
			fmt.Printf("jsonExp: %v\n", jsonExp)
			fmt.Printf("string(data): %v\n", string(data))
			t.Fail()
		}
	}
}
