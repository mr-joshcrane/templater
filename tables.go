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

// tableIterator returns a [cue.Iterator] for a given [io.Reader].
// It will attempt to parse the [io.Reader] as a CSV, transform it into a JSON string
// and finally parse the JSON string into a [cue.Iterator].
// We will use this [cue.Iterator] to walk through the table values and infer the fields types.
func tableIterator(c *cue.Context, r io.Reader) (cue.Iterator, error) {
	buf := bytes.NewBuffer([]byte{})
	df := dataframe.ReadCSV(r, dataframe.WithLazyQuotes(true))
	err := df.WriteJSON(buf)
	if err != nil {
		return cue.Iterator{}, err
	}
	cueValue := c.CompileBytes(buf.Bytes())
	return cueValue.List()
}

// generateTables will walk through the given [inputDir] and generate the [Table]s.
// It will return a map of *[Table]s keyed by the table name.
// Once we have this intermediate reprsentation, we no longer need the tables on disk.
func generateTables(fsys fs.FS, projectName string, unpackPaths ...string) ([]*Table, error) {
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

// generateTableFields will iterate over the CUE representation of the table data and infer the fields types.
func generateTableFields(table *Table, c *cue.Context, unpackPaths ...string) error {
	iterator, err := tableIterator(c, table.rawContents)
	if err != nil {
		return err
	}
	err = table.InferFields(iterator, unpackPaths...)
	if err != nil {
		return err
	}
	return nil
}

// writeTableModel will write a given Table to disk in its transform/public SQL representations.
func writeTableModel(table *Table) error {
	transformFile := fmt.Sprintf("output/transform/TRANS01_%s.sql", table.Name)
	file, err := os.Create(transformFile)
	if err != nil {
		return err
	}
	err = writeTransformSQLModel(*table, file)
	if err != nil {
		return err
	}
	publicFile := fmt.Sprintf("output/public/%s.sql", table.Name)
	file, err = os.Create(publicFile)
	if err != nil {
		return err
	}
	err = writePublicSQLModel(*table, file)
	if err != nil {
		return err
	}
	return nil
}
