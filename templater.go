package templater

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/json"
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

func GenerateTemplate(filePath string, project string, table string) (string, error) {
	var template strings.Builder
	project = strings.ToUpper(project)
	table = strings.ToUpper(table)
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return "", errors.New("unable to read file")
	}
	expr, err := json.Extract("", fileContents)
	if err != nil {
		fmt.Println(err)
		return "", errors.New("unable to convert json to cue")
	}
	c := cuecontext.New()
	v := c.BuildExpr(expr)

	iter, err := v.Fields()
	if err != nil {
		return "", errors.New("unable to iterate through cue fields")
	}

	for iter.Next() {
		k := iter.Selector().String()
		t := SnowflakeTypes[iter.Value().IncompleteKind().String()]
		kFormat := strings.ToUpper(k)
		line := fmt.Sprintf("\t\"V\":%s::%s AS %s,\n", k, t, kFormat)
		template.WriteString(line)
	}
	t := template.String()
	if t == "" {
		return "", errors.New("empty JSON")
	}
	template.Reset()
	commaIdx := strings.LastIndex(t, ",")
	t = t[0:commaIdx] + t[commaIdx+1:]
	template.WriteString(fmt.Sprintf("{{ config(tags=['%s', '%s']) }}\n\n", project, table))
	template.WriteString("SELECT\n")
	template.WriteString(t)
	template.WriteString(fmt.Sprintf("FROM\n\t{{ source('%s', '%s') }}\n", project, table))
	return template.String(), nil
}
