package templater

import (
	"bytes"
	"embed"
	"errors"
	"strings"
	"text/template"

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
type Metadata struct {
	Project string
	Table   string
	Fields  []Fields
}

func GenerateTemplate(contents []byte, project string, table string) (string, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}

	tpl, err := template.New("template.gohtml").Funcs(funcMap).ParseFS(fs, "template.gohtml")
	if err != nil {
		return "", err
	}

	expr, err := json.Extract("", contents)
	if err != nil {
		return "", errors.New("unable to convert json to cue")
	}
	c := cuecontext.New()
	v := c.BuildExpr(expr)

	iter, err := v.Fields()
	if err != nil {
		return "", errors.New("unable to iterate through cue fields")
	}

	metadata := Metadata{
		Project: strings.ToUpper(project),
		Table:   strings.ToUpper(table),
		Fields:  []Fields{},
	}

	for iter.Next() {
		k := iter.Selector().String()
		t := SnowflakeTypes[iter.Value().IncompleteKind().String()]
		metadata.Fields = append(metadata.Fields, Fields{
			Name: k,
			Type: t,
		})
	}
	if len(metadata.Fields) == 0 {
		return "", errors.New("empty JSON")
	}

	var body bytes.Buffer
	err = tpl.Execute(&body, metadata)
	if err != nil {
		return "", err
	}
	return body.String(), nil
}
