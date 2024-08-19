package common

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
)

type Report struct {
	violations []Violation
}

func (r Report) MarshalJSON() ([]byte, error) {
	filePathMap := map[string][]Violation{}
	for _, vio := range r.violations {
		_, ok := filePathMap[vio.FilePath]
		if !ok {
			filePathMap[vio.FilePath] = []Violation{}
		}
		filePathMap[vio.FilePath] = append(filePathMap[vio.FilePath], vio)
	}
	return json.Marshal(filePathMap)
}

func (r Report) WriteCsv(parentWriter io.Writer) error {
	w := csv.NewWriter(parentWriter)
	for _, vio := range r.violations {
		records := []string{vio.PluginName, vio.FilePath, fmt.Sprintf("%d", vio.StartLine), fmt.Sprintf("%d", vio.StartColumn), fmt.Sprintf("%d", vio.EndLine), fmt.Sprintf("%d", vio.EndColumn), vio.ErrorCode, vio.Message, vio.Justification}
		err := w.Write(records)
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}
