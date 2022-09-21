package templater

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
)

func CsvToJson(contents []byte) ([]byte, error) {
	rows, err := csv.NewReader(bytes.NewReader(contents)).Read()
	if err != nil {
		return nil, err
	}

	m := map[string]any{}
	for _, k := range rows {
		m[k] = ""
	}

	jsonString, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return jsonString, nil
}
