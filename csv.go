package templater

import (
	"bytes"

	"github.com/go-gota/gota/dataframe"
)

func CsvToJson(contents []byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	reader := bytes.NewReader(contents)
	df := dataframe.ReadCSV(reader, dataframe.WithLazyQuotes(true))
	err := df.WriteJSON(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
