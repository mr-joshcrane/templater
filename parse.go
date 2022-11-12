package templater

import (
	"errors"
	"fmt"
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

// InferFields takes a cue.Iterator and walks through it, adding fields to the table
// It will also unpack any JSON fields where the column name matches the (optional) unpackPath
func (t *Table) InferFields(iter cue.Iterator, unpackPaths ...string) error {
	for iter.Next() {
		// if any, iterate through our raw VARIANTs and unpack them
		for _, unpackPath := range unpackPaths {
			JSONString, err := lookupCuePath(iter.Value(), unpackPath)
			if err != nil {
				return err
			}

			unpackable, err := UnmarshalJSONFromCUE(JSONString)
			if err != nil {
				return err
			}
			unpackable.Walk(continueUnpacking, func(c cue.Value) { Unpack(t, c, func(s string) string { return fmt.Sprintf("%s:%s", unpackPath, s) }) })
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

// lookupCuePath attenpts to find a child of a cue.Value at a given path
func lookupCuePath(c cue.Value, path string) (cue.Value, error) {
	lookupPath := cue.ParsePath(path)
	if lookupPath.Err() != nil {
		return cue.Value{}, lookupPath.Err()
	}
	return c.LookupPath(lookupPath), nil
}

// UnmarshalJSONFromCUE takes a [cue.Value] that is assumed to be a JSON string
// and attempts to marshal it to JSON, returning an error if unable to do so
func UnmarshalJSONFromCUE(c cue.Value) (cue.Value, error) {
	byt, err := c.Bytes()
	if err != nil {
		return cue.Value{}, err
	}
	e, err := json.Extract("", byt)
	if err != nil {
		return cue.Value{}, err
	}
	c = c.Context().BuildExpr(e)
	return c, nil
}

// Unpack constructs a field from a cue.Value
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

func ContainsNonLeadingArray(path string) bool {
	path = arrayAtLineStart.ReplaceAllString(path, "")
	if arrayInLine.MatchString(path) {
		return true
	}
	return false
}

func continueUnpacking(c cue.Value) bool {
	return !ContainsNonLeadingArray(c.Path().String())
}

type NameOption func(string) string

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
	return s
}

func CleanTableName(path string) string {
	tableName := filepath.Base(path)
	tableName = strings.ToUpper(tableName)
	tableName = strings.ReplaceAll(tableName, ".CSV", "")
	tableName = strings.Join(validCharacters.FindAllString(tableName, -1), "")
	return tableName
}
