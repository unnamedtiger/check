package test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/unnamedtiger/check/common"
	"github.com/unnamedtiger/check/plugins/unwanted_imports"
)

type testFile struct {
	Encountered   bool
	FoundTags     []string
	FoundMessages []string
}

type testJustification struct {
	Tag     string
	Message string
	Found   bool
}

func TestTool(t *testing.T) {
	plugins := []*common.Plugin{
		unwanted_imports.Plugin,
	}

	directories := []string{"data"}

	files := map[string]testFile{}
	for _, dir := range directories {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			files[path] = testFile{}
			return nil
		})
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}
	}

	violations, err := common.RunChecksForDirectories(plugins, directories)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	for _, vio := range violations {
		tf := files[vio.FilePath]
		tf.Encountered = true

		if vio.Justification == nil {
			fmt.Printf("found unjustified violation: %s\n", vio)
			t.FailNow()
		}

		tf.FoundTags = append(tf.FoundTags, vio.Justification.Tag)
		tf.FoundMessages = append(tf.FoundMessages, vio.Justification.Message)

		files[vio.FilePath] = tf
	}

	for filename, tf := range files {
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}

		justifications := common.ExtractJustifications(string(content), 0, 0)
		testJustifications := []testJustification{}
		for _, j := range justifications {
			testJustifications = append(testJustifications, testJustification{j.Tag, j.Message, false})
		}

		for i := 0; i < len(tf.FoundTags); i++ {
			for j := 0; j < len(testJustifications); j++ {
				if tf.FoundTags[i] == testJustifications[j].Tag && tf.FoundMessages[i] == testJustifications[j].Message {
					testJustifications[j].Found = true
					break
				}
			}
		}

		for _, j := range testJustifications {
			if !j.Found {
				fmt.Printf("the following justification was not detected: file=%s, tag=%s, message=%s\n", filename, j.Tag, j.Message)
				t.FailNow()
			}
		}
	}
}
