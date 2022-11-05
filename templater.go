package templater

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	Name    string
	Project string
	Fields  map[string]Field
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

func GenerateTemplateFiles(filePaths []string, unpackPaths ...string) error {
	c := cuecontext.New()
	projectName := filepath.Dir(filePaths[0])
	projectName = filepath.Base(projectName)

	tables := []*Table{}

	for _, path := range filePaths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		cueValue, err := ConvertToCueValue(c, bytes.NewReader(contents))
		if err != nil {
			return err
		}

		tableName := CleanTableName(path)

		table := Table{
			Name:    tableName,
			Fields:  make(map[string]Field),
			Project: projectName,
		}

		iter, err := cueValue.List()
		if err != nil {
			return err
		}
		err = table.InferFields(iter, unpackPaths...)
		if err != nil {
			return err
		}
		tables = append(tables, &table)

		transformFile := fmt.Sprintf("output/transform/TRANS01_%s.sql", table.Name)
		file, err := os.Create(transformFile)
		if err != nil {
			return err
		}
		err = WriteTransformSQLModel(table, file)
		if err != nil {
			return err
		}
		publicFile := fmt.Sprintf("output/public/%s.sql", table.Name)
		file, err = os.Create(publicFile)
		if err != nil {
			return err
		}
		err = WritePublicSQLModel(table, file)
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

	var files []string
	dir, err := os.ReadDir(workingDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
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

	for _, file := range dir {
		if strings.HasSuffix(file.Name(), ".csv") {
			p := path.Join(workingDir, file.Name())
			files = append(files, p)
		}
	}
	err = GenerateTemplateFiles(files, os.Args[1:]...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
