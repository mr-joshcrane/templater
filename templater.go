package templater

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/go-gota/gota/dataframe"
)

type Field struct {
	Node         string
	Path         string
	InferredType string
}

type Table struct {
	Name        string
	Project     string
	Fields      map[string]Field
	rawContents io.Reader
}

func ConvertToCueValue(c *cue.Context, r io.Reader) (cue.Value, error) {
	buf := bytes.NewBuffer([]byte{})
	df := dataframe.ReadCSV(r, dataframe.WithLazyQuotes(true))
	err := df.WriteJSON(buf)
	if err != nil {
		return cue.Value{}, err
	}
	return c.CompileBytes(buf.Bytes()), nil
}

func GenerateTemplateFiles(fsys fs.FS, projectName string, unpackPaths ...string) error {
	c := cuecontext.New()

	tables := []*Table{}
	err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {

		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".csv" {

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
		return err
	}

	for _, table := range tables {
		cueValue, err := ConvertToCueValue(c, table.rawContents)
		if err != nil {
			return err
		}

		iter, err := cueValue.List()
		if err != nil {
			return err
		}
		err = table.InferFields(iter, unpackPaths...)
		if err != nil {
			return err
		}
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
	}

	models := GenerateModel(tables)
	sources := generateSources(tables, projectName)

	return WriteProperties(c, models, sources)
}

func Main() int {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	fsys := os.DirFS(workingDir)

	_, err = os.Stat("output")
	if err != nil {
		err := os.Mkdir("output", os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	}
	_, err = os.Stat("output/transform")
	if err != nil {
		err := os.Mkdir("output/transform", os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	}
	_, err = os.Stat("output/public")
	if err != nil {
		err := os.Mkdir("output/public", os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	}

	err = GenerateTemplateFiles(fsys, "PROJECT", os.Args[1:]...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
