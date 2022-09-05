package templater

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/json"
)

var SnowflakeTypes = map[string]string {
	"string": "STRING",
	"int": "INTEGER",
	"float": "FLOAT",	
	"struct": "OBJECT",
	"list": "ARRAY",
	"null": "VARCHAR",
	"bool": "BOOLEAN",
}

func GenerateTemplate(filePath string, project string, table string) string {	
	var template strings.Builder
	project = strings.ToUpper(project)
	table = strings.ToUpper(table)
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	expr, err := json.Extract("", fileContents)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c := cuecontext.New()
	v := c.BuildExpr(expr)

	iter, err := v.Fields()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	template.WriteString(fmt.Sprintf("{{ config(tags=['%s', '%s']) }}\n\n", project, table))
	template.WriteString("SELECT\n")
	for iter.Next() {
		k := iter.Selector().String()
		t := SnowflakeTypes[iter.Value().IncompleteKind().String()]
		kFormat := strings.ToUpper(k)
		line := fmt.Sprintf("\t\"V\":%s::%s AS %s,\n", k, t, kFormat)
		template.WriteString(line)
	}
	template.WriteString(fmt.Sprintf("FROM\n\t{{ source('%s', '%s') }}\n", project, table))
	return template.String()
}
