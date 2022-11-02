package templater

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

func (t *Table) InferFields(iter cue.Iterator, unpackPaths ...string) error {
	for iter.Next() {
		// if any, iterate through our raw VARIANTs and unpack them
		for _, unpackPath := range unpackPaths {
			unpackable := unpackJSON(iter.Value(), unpackPath)
			unpackable.Walk(continueUnpacking, func(c cue.Value) { Unpack(t, c, prefix) })
		}

		iter.Value().Walk(
			func(c cue.Value) bool {
				return true
			},
			func(c cue.Value) {
				Unpack(t, c)
			})
		if len(t.Fields) == 0 {
			return errors.New("empty JSON")
		}

		// if any remove any of the raw VARIANT originals
		for _, unpackPath := range unpackPaths {
			delete(t.Fields, unpackPath)
		}

	}

	return nil
}

func unpackJSON(item cue.Value, path string) cue.Value {
	unpackable := item.Value().LookupPath(cue.ParsePath(path))
	if unpackable.Exists() {
		byt, err := unpackable.Bytes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "path %s exists but failed to represent as bytes", path)
			panic(err)
		}
		e, err := json.Extract("", byt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "path %s exists but failed to convert from JSON to CUE", path)
			panic(err)
		}
		unpackable = unpackable.Context().BuildExpr(e)
	}
	return unpackable
}

func Unpack(t *Table, c cue.Value, opts ...NameOption) {
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

	cueType := c.IncompleteKind().String()
	inferredType := SnowflakeTypes[cueType]

	// If we've found an object, no need keep track of it. We'll walk into the member objects instead
	if inferredType == "OBJECT" {
		return
	}

	field := Field{
		Node:         node,
		Path:         EscapePath(path),
		InferredType: inferredType,
	}

	if _, ok := t.Fields[path]; !ok {
		t.Fields[path] = field
		return
	}
	existingField := t.Fields[path]
	// If we couldn't get a type example yet, we'll update
	if existingField.InferredType == "VARCHAR" {
		t.Fields[path] = field
	}
}

var arrayAtLineStart = regexp.MustCompile(`^[[0-9]*].`)
var arrayInLine = regexp.MustCompile(`[\[[0-9]]`)
var validCharacters = regexp.MustCompile(`[A-Z0-9._ ]*`)
var camelCase = regexp.MustCompile(`([a-z])(A?)([A-Z])`)

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

func EscapePath(s string) string {
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, `:`, `":"`)
	s = strings.ReplaceAll(s, `.`, `"."`)
	s = fmt.Sprintf(`"%s"`, s)
	return s
}

func NormaliseKey(s string) string {
	s = camelCase.ReplaceAllString(s, `$1 $2 $3`)
	s = strings.ToUpper(s)
	s = strings.Join(validCharacters.FindAllString(s, -1), " ")
	s = strings.Join(strings.Fields(s), "_")
	s = strings.Trim(s, ` `)
	s = strings.ReplaceAll(s, `.`, `__`)
	s = strings.ReplaceAll(s, ` `, `_`)
	return s
}

type NameOption func(string) string

func CleanTableName(path string) string {
	tableName := filepath.Base(path)
	tableName = strings.ToUpper(tableName)
	tableName = strings.ReplaceAll(tableName, ".CSV", "")
	tableName = strings.Join(validCharacters.FindAllString(tableName, -1), "")
	return tableName
}

func prefix(s string) string {
	return "V:" + s
}
