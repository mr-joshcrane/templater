package templater

import (
	"errors"
	"regexp"
	"strings"

	"cuelang.org/go/cue"
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

func MakeTable(v cue.Value, tableName, projectName string, unpackPaths ...string) (Table, error) {
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
		// if any, iterate through our raw VARIANTs and unpack them
		for _, unpackPath := range unpackPaths {
			unpackable := unpackJSON(item.Value(), unpackPath)
			unpackable.Walk(continueUnpacking, func(c cue.Value) { unpack(&table, c, prefix) })
		}

		item.Value().Walk(
			func(c cue.Value) bool {
				return true
			},
			func(c cue.Value) {
				unpack(&table, c, func(s string) string {
					return strings.ReplaceAll(s, `"`, ``)
				})
			})
		if len(table.Fields) == 0 {
			return Table{}, errors.New("empty JSON")
		}

		// if any remove any of the raw VARIANT originals
		for _, unpackPath := range unpackPaths {
			delete(table.Fields, unpackPath)
		}

	}

	return table, nil
}

func unpack(t *Table, c cue.Value, opts ...NameOption) {
	path := c.Path().String()
	path = arrayAtLineStart.ReplaceAllString(path, "")
	// If theres an array in this path, no need to unpack it
	if arrayInLine.MatchString(path) {
		return
	}
	node := NormaliseKey(path)

	for _, opt := range opts {
		path = opt(path)
	}
	path = stripAndEscapeQuotes(path)

	cueType := c.IncompleteKind().String()
	inferredType := SnowflakeTypes[cueType]

	// If we've found an object, no need keep track of it. We'll walk into the member objects instead
	if inferredType == "OBJECT" {
		return
	}

	field := Field{
		Node:        node,
		Path:        path,
		InferedType: inferredType,
	}

	if _, ok := t.Fields[path]; !ok {
		t.Fields[path] = field
		return
	}
	existingField := t.Fields[path]
	// If we couldn't get a type example yet, we'll update
	if existingField.InferedType == "VARCHAR" {
		t.Fields[path] = field
	}
}

var arrayAtLineStart = regexp.MustCompile(`^[[0-9]*].`)
var arrayInLine = regexp.MustCompile(`[\[[0-9]]`)

func ContainsArray(path string) bool {
	path = arrayAtLineStart.ReplaceAllString(path, "")
	if arrayInLine.MatchString(path) {
		return true
	}
	return false
}

func continueUnpacking(c cue.Value) bool {
	return !ContainsArray(c.Path().String())
}

func unpackJSON(item cue.Value, path string) cue.Value {
	unpackable := item.Value().LookupPath(cue.ParsePath(path))
	if unpackable.Exists() {
		byt, err := unpackable.Bytes()
		if err != nil {
			panic(err)
		}
		e, err := json.Extract("", byt)
		if err != nil {
			panic(err)
		}
		unpackable = unpackable.Context().BuildExpr(e)
	}
	return unpackable
}

func stripAndEscapeQuotes(s string) string {
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, `:`, `":"`)
	s = strings.ReplaceAll(s, `.`, `"."`)
	return s
}
