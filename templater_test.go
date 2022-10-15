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
	testscript.Run(t, testscript.Params{Dir: "testdata/script"})
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
	fields := []templater.Field{
		{
			Name: "Team",
			Type: "STRING",
		},
		{
			Name: "Payroll(millions)",
			Type: "FLOAT",
		},
		{
			Name: "Wins",
			Type: "INTEGER",
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
		t.Fatalf(cmp.Diff(want, got))
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
