package templater_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rogpeppe/go-internal/testscript"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
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

func createCueValue(t *testing.T, literal string) cue.Value {
	c := cuecontext.New()
	v := c.CompileString(literal)
	if v.Err() != nil {
		t.Fatal(v.Err())
	}
	return v
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
func ExampleGenerateTagsSQL() {
	PROJECT := "A_ProjectName"
	TABLE := "A_TableName"
	fmt.Println(templater.GenerateTagsSQL(PROJECT, TABLE))
	// Output:
	// {{ config(tags=['A_PROJECTNAME', 'A_TABLENAME']) }}
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

func ExampleGenerateColumnsSQL() {
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
	fmt.Println(templater.GenerateColumnsSQL(fields))
	// Output:
	//   "Payroll(millions)"::FLOAT AS PAYROLL_MILLIONS
	//   ,"Team"::STRING AS TEAM
	//   ,"Wins"::INTEGER AS WINS
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

func ExampleGenerateSourceSQL() {
	PROJECT := "A_ProjectName"
	TABLE := "A_TableName"
	fmt.Println(templater.GenerateSourceSQL(PROJECT, TABLE))
	// Output:
	//   {{ source('A_PROJECTNAME', 'A_TABLENAME') }}
}

func TestGenerateProjectModel_GivenASetOfTablesGeneratesAppropriateModel(t *testing.T) {
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

	got := templater.GenerateProjectModel(tables)
	if cmp.Equal(want, got) {
		t.Fatalf(cmp.Diff(want, got))
	}
}

func TestContainsNonLeadingArray_IsTrueWhenPathContainsArray(t *testing.T) {
	t.Parallel()
	path := "meta.mass_edit_custom_type_ids[123]"
	if !templater.ContainsNonLeadingArray(path) {
		t.Fatal(path)
	}
}

func TestContainsNonLeadingArray_IsFalseWhenPathDoesNotContainArray(t *testing.T) {
	t.Parallel()
	path := "meta.mass_edit_custom_type_ids"
	if templater.ContainsNonLeadingArray(path) {
		t.Fatal(path)
	}
}

func TestContainsNonLeadingArray_IsFalseWhenContainsOnlyLeadingArray(t *testing.T) {
	t.Parallel()
	path := "[123]meta.mass_edit_custom_type_ids"
	if templater.ContainsNonLeadingArray(path) {
		t.Fatal(path)
	}
}

func ExampleContainsNonLeadingArray() {
	path1 := "meta.mass_edit_custom_type_ids[123]"
	path2 := "meta.mass_edit_custom_type_ids"
	path3 := "[123]meta.mass_edit_custom_type_ids"
	fmt.Println(templater.ContainsNonLeadingArray(path1))
	fmt.Println(templater.ContainsNonLeadingArray(path2))
	fmt.Println(templater.ContainsNonLeadingArray(path3))
	// Output:
	// true
	// false
	// false
}

func TestNormaliseKey_Normalises(t *testing.T) {
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
		t.Run(c.Description, func(t *testing.T) {
			got := templater.NormaliseKey(c.Key)
			if c.Want != got {
				t.Errorf("%s: wanted %s, got %s", c.Description, c.Want, got)
			}
		})

	}
}

func ExampleNormaliseKey() {
	rule_uppercased := "thisisakey"
	rule_spaces_to_underscore := "this is a key"
	rule_nonalphanumeric_stripped := "this%^@is``a()*key"
	rule_dot_seperaters_to_double_underscore := "json.payload.and_children"
	rule_leading_and_trailing_space_trimmed := "       THISISAKEY          "
	rule_parenthesised_words_considered_word_boundaries := "(THIS)IS(A)KEY"
	rule_camel_case_considered_seperate_words := "thisIsAKey"

	fmt.Println(templater.NormaliseKey(rule_uppercased))
	fmt.Println(templater.NormaliseKey(rule_spaces_to_underscore))
	fmt.Println(templater.NormaliseKey(rule_nonalphanumeric_stripped))
	fmt.Println(templater.NormaliseKey(rule_dot_seperaters_to_double_underscore))
	fmt.Println(templater.NormaliseKey(rule_leading_and_trailing_space_trimmed))
	fmt.Println(templater.NormaliseKey(rule_parenthesised_words_considered_word_boundaries))
	fmt.Println(templater.NormaliseKey(rule_camel_case_considered_seperate_words))

	// Output:
	// THISISAKEY
	// THIS_IS_A_KEY
	// THIS_IS_A_KEY
	// JSON__PAYLOAD__AND_CHILDREN
	// THISISAKEY
	// THIS_IS_A_KEY
	// THIS_IS_A_KEY
}

func TestCleanTableName_DerivesAValidTableNameFromItsPath(t *testing.T) {
	t.Parallel()
	got := templater.CleanTableName("some/file/path/table_NamE@.csv")
	want := "TABLE_NAME"
	if want != got {
		t.Errorf("wanted %s, got %s", want, got)
	}
}

func TestEscapePath_CorrectlySQLEscapesDatabaseIdentifiers(t *testing.T) {
	t.Parallel()
	got := templater.EscapePath(`V:attributes."available_in"`)
	want := `"V":"attributes"."available_in"`
	if want != got {
		t.Errorf("wanted %s, got %s", want, got)
	}
}

func TestInferFields_ErrorsIfGivenBlankJSON(t *testing.T) {
	t.Parallel()
	table := templater.Table{
		Name:    "TABLE",
		Project: "PROJECT",
		Fields:  make(map[string]templater.Field),
	}
	v := createCueValue(t, "[{}]")
	iter, err := v.List()
	if err != nil {
		t.Fatal(err)
	}
	err = table.InferFields(iter)
	if err == nil {
		t.Fatal("no error thrown when passed an empty JSON")
	}
}

func TestInferFields_GivenRowInfersType(t *testing.T) {
	t.Parallel()
	table := templater.Table{
		Name:    "TABLE",
		Project: "PROJECT",
		Fields:  make(map[string]templater.Field),
	}
	v := createCueValue(t, `[{a: 1, b: "2", c: true, d: 1.1, e: [1, 2, 3], f: {g: 1, h: "2"}}]`)
	iter, err := v.List()
	if err != nil {
		t.Fatal(err)
	}
	err = table.InferFields(iter)
	if err != nil {
		t.Fatal(err)
	}

	if table.Fields["a"].InferredType != "INTEGER" {
		t.Fatalf("expected 'a' to be inferred as INTEGER, got %s", table.Fields["a"].InferredType)
	}
	if table.Fields["b"].InferredType != "STRING" {
		t.Fatalf("expected 'b' to be inferred as STRING, got %s", table.Fields["b"].InferredType)
	}
	if table.Fields["c"].InferredType != "BOOLEAN" {
		t.Fatalf("expected 'c' to be inferred as BOOLEAN, got %s", table.Fields["c"].InferredType)
	}
	if table.Fields["d"].InferredType != "FLOAT" {
		t.Fatalf("expected 'd' to be inferred as FLOAT, got %s", table.Fields["d"].InferredType)
	}
	if table.Fields["e"].InferredType != "ARRAY" {
		t.Fatalf("expected 'e' to be inferred as ARRAY, got %s", table.Fields["e"].InferredType)
	}

}

func TestInferFields_GivenRowEscapesPath(t *testing.T) {
	t.Parallel()
	table := templater.Table{
		Name:    "TABLE",
		Project: "PROJECT",
		Fields:  make(map[string]templater.Field),
	}
	v := createCueValue(t, `[{ PathToEscape: true,}]`)
	iter, err := v.List()
	if err != nil {
		t.Fatal(err)
	}
	err = table.InferFields(iter)
	if err != nil {
		t.Fatal(err)
	}
	if table.Fields["PathToEscape"].Path != "\"PathToEscape\"" {
		t.Fatalf("expected 'PathToEscape' to be escaped as \"PathToEscape\", got %s", table.Fields["PathToEscape"].Path)
	}
}

func TestInferFields_GivenRowNormalisesNode(t *testing.T) {
	t.Parallel()
	table := templater.Table{
		Name:    "TABLE",
		Project: "PROJECT",
		Fields:  make(map[string]templater.Field),
	}
	v := createCueValue(t, `[{ "this is a key": true,}]`)
	iter, err := v.List()
	if err != nil {
		t.Fatal(err)
	}
	err = table.InferFields(iter)
	if err != nil {
		t.Fatal(err)
	}
	if table.Fields[`"this is a key"`].Node != "THIS_IS_A_KEY" {
		t.Fatalf("expected 'this_is_a_key' to be normalised as THIS_IS_A_KEY, got %s", table.Fields["this is a key"].Node)
	}
}

func TestInferFields_UnpacksAndRemovesRawEntry(t *testing.T) {
	t.Parallel()
	table := templater.Table{
		Name:    "TABLE",
		Project: "PROJECT",
		Fields:  make(map[string]templater.Field),
	}
	v := createCueValue(t, `[{ unpackable: '{"field": 1}', "someVal": true,}]`)
	iter, err := v.List()
	if err != nil {
		t.Fatal(err)
	}
	err = table.InferFields(iter, "unpackable")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := table.Fields["unpackable:field"]

	if !ok {
		t.Fatalf("expected unpackable field to be unpacked, but could not find it in %v", table.Fields)
	}
}

func TestUnmarshalJSONFromCUE_TranslatesJSONIntoItsCUERepresentation(t *testing.T) {
	t.Parallel()
	v := createCueValue(t, `'{"a": 1}'`)
	got, err := templater.UnmarshalJSONFromCUE(v)
	if err != nil {
		t.Fatal(err)
	}
	want := createCueValue(t, `{a: 1}`)
	if !want.Equals(got) {
		t.Errorf("wanted %s, got %s", want, got)
	}
}

func TestUnmarshalJSONFromCUE_ErrorsOutOnInvalidJSON(t *testing.T) {
	t.Parallel()
	v := createCueValue(t, `'{INVALID_JSON}'`)
	_, err := templater.UnmarshalJSONFromCUE(v)
	if err == nil {
		t.Fatal()
	}
}

func TestUnpackAttemptsToKeepInferringIfBestCurrentGuessIsVARCHAR(t *testing.T) {
	t.Parallel()
	table := templater.Table{
		Name:    "TABLE",
		Project: "PROJECT",
		Fields:  make(map[string]templater.Field),
	}
	v := createCueValue(t, `[
		{ a: null,},
		{ a: null,},
		{ a: 1,},
		
	]`)
	iter, err := v.List()
	if err != nil {
		t.Fatal(err)
	}
	table.InferFields(iter)
	if table.Fields["a"].InferredType != "INTEGER" {
		t.Fatalf("expected 'a' to be inferred as INTEGER, got %s", table.Fields["a"].InferredType)
	}
}
