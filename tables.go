package templater

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"github.com/go-gota/gota/dataframe"
)

func TableIterator(c *cue.Context, r io.Reader) (cue.Iterator, error) {
	buf := bytes.NewBuffer([]byte{})
	df := dataframe.ReadCSV(r, dataframe.WithLazyQuotes(true))
	err := df.WriteJSON(buf)
	if err != nil {
		return cue.Iterator{}, err
	}
	cueValue := c.CompileBytes(buf.Bytes())
	return cueValue.List()
}

func GenerateTables(fsys fs.FS, projectName string, unpackPaths ...string) ([]*Table, error) {
	tables := []*Table{}
	err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {

		if filepath.Ext(path) == ".csv" && !info.IsDir() {

			f, err := fsys.Open(path)
			if err != nil {
				return err
			}
			contents := io.Reader(f)
			table := Table{
				Name:        CleanTableName(info.Name()),
				Project:     projectName,
				Fields:      make(map[string]Field),
				rawContents: contents,
			}
			tables = append(tables, &table)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func GenerateTableFields(table *Table, c *cue.Context, unpackPaths ...string) error {
	iterator, err := TableIterator(c, table.rawContents)
	if err != nil {
		return err
	}
	err = table.InferFields(iterator, unpackPaths...)
	if err != nil {
		return err
	}
	return nil
}

func writeTableModel(table *Table) error {
	transformFile := fmt.Sprintf("output/transform/TRANS01_%s.sql", table.Name)
	file, err := os.Create(transformFile)
	if err != nil {
		return err
	}
	err = WriteTransformSQLModel(*table, file)
	if err != nil {
		return err
	}
	publicFile := fmt.Sprintf("output/public/%s.sql", table.Name)
	file, err = os.Create(publicFile)
	if err != nil {
		return err
	}
	err = WritePublicSQLModel(*table, file)
	if err != nil {
		return err
	}
	return nil
}
