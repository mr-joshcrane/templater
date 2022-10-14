package templater_test

import (
	"os"
	"testing"

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
