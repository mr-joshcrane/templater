package templater

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/json"
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

func GenerateTemplateFiles(filePaths []string) error {
	c := cuecontext.New()
	projectName := filepath.Dir(filePaths[0])
	projectName = filepath.Base(projectName)

	tables := []*Table{}

	for _, path := range filePaths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		j := bytes.NewBuffer([]byte{})
		reader := bytes.NewReader(contents)
		df := dataframe.ReadCSV(reader, dataframe.WithLazyQuotes(true))
		err = df.WriteJSON(j)
		if err != nil {
			return err
		}
		tableName := cleanTableName(path)

		expr, err := json.Extract("", j.Bytes())
		if err != nil {
			fmt.Println(err)
			return errors.New("unable to convert json to cue")
		}

		v := c.BuildExpr(expr)

		table, err := MakeTable(v, tableName, projectName, "V")
		if err != nil {
			return err
		}

		tables = append(tables, &table)

		filename := fmt.Sprintf("output/%s.sql", table.Name)
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		err = GenerateSQLModel(table, file)
		if err != nil {
			return err
		}
	}

	models := GenerateModel(tables)
	sources := generateSources(tables, projectName)

	return WriteProperties(c, models, sources)
}

func Main() int {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "takes no arguments, run in the PROJECT folder and make sure CSV files are present")
		return 1
	}

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
			return 1
		}
	}

	for _, file := range dir {
		if strings.HasSuffix(file.Name(), ".csv") {
			p := path.Join(workingDir, file.Name())
			files = append(files, p)
		}
	}
	err = GenerateTemplateFiles(files)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
