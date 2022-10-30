package templater_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rogpeppe/go-internal/testscript"

	"github.com/mr-joshcrane/templater"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"main": templater.Main,
	}))
}

func TestScript(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{Dir: "./testdata/script"})
}

func TestTagStatementGeneratesCorrectly(t *testing.T) {
	t.Parallel()
	PROJECT := "A_ProjectName"
	TABLE := "A_TableName"
	got := templater.GenerateTagsSQL(PROJECT, TABLE)
	want := "{{ config(tags=['A_PROJECTNAME', 'A_TABLENAME']) }}"
	if want != got {
		t.Fatalf(cmp.Diff(want, got))
	}
}

func TestColumnStatementGeneratesCorrectly(t *testing.T) {
	t.Parallel()
	fields := map[string]templater.Field{
		"Team": {
			Path:         "Team",
			Node:         "Team",
			InferredType: "STRING",
		},
		"Payroll(millions)": {
			Path:         "Payroll(millions)",
			Node:         "Payroll(millions)",
			InferredType: "FLOAT",
		},
		"Wins": {
			Path:         "Wins",
			Node:         "Wins",
			InferredType: "INTEGER",
		},
	}
	got := templater.GenerateColumnsSQL(fields)
	want := `  "Payroll(millions)"::FLOAT AS PAYROLL_MILLIONS
  ,"Team"::STRING AS TEAM
  ,"Wins"::INTEGER AS WINS`
	if want != got {
		t.Fatalf(cmp.Diff(want, got))
	}
}

func TestSourceStatementGeneratesCorrectly(t *testing.T) {
	t.Parallel()
	PROJECT := "A_ProjectName"
	TABLE := "A_TableName"
	got := templater.GenerateSourceSQL(PROJECT, TABLE)
	want := "  {{ source('A_PROJECTNAME', 'A_TABLENAME') }}"
	if want != got {
		t.Fatalf("wanted %s, got %s", want, got)
	}
}

func TestGenerateModelTransformationFromTable(t *testing.T) {
	t.Parallel()
	tables := []*templater.Table{
		{
			Name: "BASEBALL",
			Fields: map[string]templater.Field{
				"PAYROLL_MILLIONS": {
					Node: "PAYROLL_MILLIONS",
				},
				"TEAM": {
					Node: "TEAM",
				},
				"WINS": {
					Node: "WINS",
				},
			},
		},
		{
			Name: "FREQUENCY",
			Fields: map[string]templater.Field{
				"FREQUENCY": {
					Node: "FREQUENCY",
				},
				"LETTER": {
					Node: "LETTER",
				},
				"PERCENTAGE": {
					Node: "PERCENTAGE",
				},
			},
		},
	}

	want := []templater.Model{
		{
			Name: "BASEBALL",
			Columns: []templater.Column{
				{
					Name: "PAYROLL_MILLIONS",
				},
				{
					Name: "TEAM",
				},
				{
					Name: "WINS",
				},
			},
		},
		{

			Columns: []templater.Column{
				{
					Name: "PAYROLL_MILLIONS",
				},
				{
					Name: "TEAM",
				},
				{
					Name: "WINS",
				},
			},
		},
	}

	got := templater.GenerateModel(tables)
	if cmp.Equal(want, got) {
		t.Fatalf(cmp.Diff(want, got))
	}
}

func TestContainsArray_IsTrueWhenPathContainsArray(t *testing.T) {
	t.Parallel()
	path := "meta.mass_edit_custom_type_ids[123]"
	if !templater.ContainsArray(path) {
		t.Fatal(path)
	}
}

func TestContainsArray_IsFalseWhenPathDoesNotContainArray(t *testing.T) {
	t.Parallel()
	path := "meta.mass_edit_custom_type_ids"
	if templater.ContainsArray(path) {
		t.Fatal(path)
	}
}

func TestContainsArray_IsFalseWhenContainsOnlyLeadingArray(t *testing.T) {
	t.Parallel()
	path := "[123]meta.mass_edit_custom_type_ids"
	if templater.ContainsArray(path) {
		t.Fatal(path)
	}
}

func TestNormaliseKey_NormalisesAKey(t *testing.T) {
	t.Parallel()
	tc := []struct {
		Description string
		Key         string
		Want        string
	}{
		{
			Description: "Should be uppercased",
			Key:         "thisisakey",
			Want:        "THISISAKEY",
		},
		{
			Description: "Spaces should be converted to underscores",
			Key:         "this is a key",
			Want:        "THIS_IS_A_KEY",
		},
		{
			Description: "Non alphanumeric or underscore characters should be stripped out",
			Key:         "this%^@is``a()*key",
			Want:        "THIS_IS_A_KEY",
		},
		{
			Description: "JSON payloads separated by .'s should be separated instead by double underscore",
			Key:         "json.payload.and_children",
			Want:        "JSON__PAYLOAD__AND_CHILDREN",
		},
		{
			Description: "Keys have leading and trailing spaces stripped",
			Key:         "       THISISAKEY          ",
			Want:        "THISISAKEY",
		},
		{
			Description: "Parenthesised words are considered word boundaries",
			Key:         "(THIS)IS(A)KEY",
			Want:        "THIS_IS_A_KEY",
		},
		{
			Description: "Camel Case is interpreted as a word boundary",
			Key:         "thisIsAKey",
			Want:        "THIS_IS_A_KEY",
		},
	}
	for _, c := range tc {
		got := templater.NormaliseKey(c.Key)
		if c.Want != got {
			t.Errorf("%s: wanted %s, got %s", c.Description, c.Want, got)
		}
	}
}

func TestCleanTableName_DerivesATableNameFromItsPath(t *testing.T) {
	t.Parallel()
	got := templater.CleanTableName("some/file/path/table_NamE@.csv")
	want := "TABLE_NAME"
	if want != got {
		t.Errorf("wanted %s, got %s", want, got)
	}
}

func TestEscapePath(t *testing.T) {
	t.Parallel()
	got := templater.EscapePath(`V:attributes."available_in"`)
	want := `"V":"attributes"."available_in"`
	if want != got {
		t.Errorf("wanted %s, got %s", want, got)
	}
}

// TestScript -> Setup of files -> Should produce some outcome
// Generate template:
// creates a new context, then creates an output folder if one does not exist
// func map for string functions
// run the functions over a template, gohtml return a TEMPLATE
// determine the project name
// for each of the files:
//		read the files contents
//		attempt to convert it to json
//		determine the tablename (from the file name)
//		convert json to cue
//		create an iterator
//      create state to store data (?)
//     	for type  in iterator
//			unify it with the empty type (which will give us the type) (?)
//			iterate through the fields again
//			store the type if new
//			supercede the type if insufficiently complex
//
//		if the length of the type map is 0, return error because empty json
// 		append the key and the type to a table fields entry
//		given table information, execute the template
//		write out the template
//		given the metadata, generate the transform
//		write to disk
//		given the metadata, generate the public
//		write to disk
//		given the metadata, generate the transform
//		write to disk
