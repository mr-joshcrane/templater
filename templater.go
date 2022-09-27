package templater

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/json"
)

var (
	//go:embed template.gohtml
	fs embed.FS
)

var SnowflakeTypes = map[string]string{
	"string": "STRING",
	"int":    "INTEGER",
	"float":  "FLOAT",
	"struct": "OBJECT",
	"list":   "ARRAY",
	"null":   "VARCHAR",
	"bool":   "BOOLEAN",
}

type Fields struct {
	Name string
	Type string
}

type Table struct {
	TableName string
	Project   string
	Fields    []Fields
	TypeMap   map[string]string
}
type Metadata struct {
	Tables map[string]Table
}

func GenerateTemplate(filePaths []string) error {
	c := cuecontext.New()
	_, err := os.Stat("output")
	if err != nil {
		err := os.Mkdir("output", 0777)
		if err != nil {
			return err
		}
	}

	tables := make(map[string]Table)
	metadata := Metadata{
		Tables: tables,
	}

	funcMap := template.FuncMap{
		"ToUpper":      strings.ToUpper,
		"Strip":        strip,
		"TrimBrackets": func(s string) string { return strings.ReplaceAll(s, `(`, " ") },
		"TrimLeft":     func(s string) string { return strings.TrimLeft(s, ` `) },
		"SpaceReplace": func(s string) string { return strings.ReplaceAll(s, ` `, `_`) },
	}

	tpl, err := template.New("template.gohtml").Funcs(funcMap).ParseFS(fs, "template.gohtml")
	if err != nil {
		return err
	}

	projectName := filepath.Dir(filePaths[0])
	projectName = filepath.Base(projectName)

	for _, path := range filePaths {
		contents, err := os.ReadFile(path)
		if err == nil {
			contents, err = CsvToJson(contents)
			if err != nil {
				return err
			}
		}
		tableName := filepath.Base(path)
		tableName = strings.ToUpper(tableName)
		tableName = strings.ReplaceAll(tableName, ".CSV", "")

		expr, err := json.Extract("", contents)
		if err != nil {
			fmt.Println(err)
			return errors.New("unable to convert json to cue")
		}

		v := c.BuildExpr(expr)
		iter, err := v.List()
		if err != nil {
			fmt.Println(err)
			return errors.New("unable to iterate through cue fields")
		}

		metadata.Tables[tableName] = Table{
			TableName: tableName,
			Fields:    []Fields{},
			TypeMap:   make(map[string]string),
			Project:   projectName,
		}
		table := metadata.Tables[tableName]
		empty := cue.Value{}
		for iter.Next() {
			unified := iter.Value().Unify(empty)
			if unified.Err() != nil {
				continue
			}

			iter2, _ := unified.Fields()
			for iter2.Next() {
				k := iter2.Selector().String()
				t := SnowflakeTypes[iter2.Value().IncompleteKind().String()]
				if _, ok := table.TypeMap[k]; !ok {
					table.TypeMap[k] = t
					continue
				}
				prevType := table.TypeMap[k]
				if prevType == "VARCHAR" || prevType == "" {
					table.TypeMap[k] = t
				}
			}
		}

		if len(table.TypeMap) == 0 {
			return errors.New("empty JSON")
		}

		for k, v := range table.TypeMap {
			table.Fields = append(table.Fields, Fields{
				Name: k,
				Type: v,
			})
		}

		var body bytes.Buffer
		err = tpl.Execute(&body, table)
		if err != nil {
			return err
		}
		filename := fmt.Sprintf("output/%s.sql", table.TableName)
		err = os.WriteFile(filename, body.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	transformModel, err := generateTransform(c, metadata)
	if err != nil {
		return err
	}

	err = os.WriteFile("output/transform_schema.yml", []byte(transformModel), 0644)
	if err != nil {
		return err
	}
	publicModel, err := generatePublic(c, metadata)
	if err != nil {
		return err
	}
	err = os.WriteFile("output/public_schema.yml", []byte(publicModel), 0644)
	if err != nil {
		return err
	}
	sourceModel, err := generateSources(c, metadata.Tables, projectName)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("output/source_schema.yml", []byte(sourceModel), 0644)
	if err != nil {
		return err
	}
	return nil
}

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == '_' ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func formatKey(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, `(`, " ")
	s = strip(s)
	s = strings.TrimLeft(s, ` `)
	s = strings.ReplaceAll(s, ` `, `_`)
	return s
}
