package templater

import (
	"errors"
	"regexp"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/json"
)

func MakeTable(v cue.Value, tableName, projectName string) (Table, error) {
	table := Table{
		Name:        tableName,
		Fields:      make(map[string]Field),
		Project:     projectName,
		SQLTemplate: SQLTemplate{},
	}

	item, err := v.List()
	if err != nil {
		return Table{}, errors.New("empty JSON")
	}
	for item.Next() {
		v1 := item.Value().LookupPath(cue.ParsePath("V"))
		if v1.Exists() {
			byt, err := v1.Bytes()
			if err != nil {
				panic(err)
			}
			e, err := json.Extract("", byt)
			if err != nil {
				panic(err)
			}
			v2 := v.Context().BuildExpr(e)
			v2.Walk(stopCondition, func(c cue.Value) { unpack(&table, c, prefix) })
		}

		item.Value().Walk(func(c cue.Value) bool { return true }, func(c cue.Value) { unpack(&table, c, func(s string) string { return strings.ReplaceAll(s, `"`, ``) }) })
		if len(table.Fields) == 0 {
			return Table{}, errors.New("empty JSON")
		}

	}
	delete(table.Fields, "V")
	return table, nil
}

func unpack(t *Table, c cue.Value, opts ...NameOption) {
	exp, err := regexp.Compile(`^[[0-9]*].`)
	if err != nil {
		panic(err)
	}
	path := c.Path().String()
	path = exp.ReplaceAllString(path, "")
	node := path
	node = formatKey(node)

	for _, opt := range opts {
		path = opt(path)
	}
	path = strings.ReplaceAll(path, `"`, "")
	path = strings.ReplaceAll(path, `:`, `":"`)
	path = strings.ReplaceAll(path, `.`, `"."`)

	snowflakeType := SnowflakeTypes[c.IncompleteKind().String()]
	if strings.Contains(path, `[`) {
		return
	}
	if snowflakeType == "OBJECT" {
		return
	}

	field := Field{
		Node:        node,
		Path:        path,
		InferedType: snowflakeType,
	}

	if _, ok := t.Fields[path]; !ok {
		t.Fields[path] = field
		return
	}
	existingField := t.Fields[path]
	if existingField.InferedType == "VARCHAR" || existingField.InferedType == "" {
		existingField.InferedType = snowflakeType
	}
}

func stopCondition(c cue.Value) bool {
	exp := regexp.MustCompile(`^[[0-9]*].`)
	p := c.Path().String()
	p = exp.ReplaceAllString(p, "")
	if strings.Contains(p, `[`) {
		return false
	}
	return true
}
