package converter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDDIToXLSForm_SelectOne(t *testing.T) {
	ddiXML := `<var ID="V1" name="gender" intrvl="discrete">
		<concept>Gender</concept>
		<qstn responseDomainType="category">
			<qstnLit>What is your gender?</qstnLit>
		</qstn>
		<catgry>
			<catValu>1</catValu>
			<labl>Male</labl>
		</catgry>
		<catgry>
			<catValu>2</catValu>
			<labl>Female</labl>
		</catgry>
		<varFormat type="numeric" schema="other"/>
	</var>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(form.Survey) != 1 {
		t.Fatalf("Expected 1 survey row, got %d", len(form.Survey))
	}
	if form.Survey[0].Type != "select_one gender" {
		t.Errorf("Expected type 'select_one gender', got %s", form.Survey[0].Type)
	}
	if form.Survey[0].Name != "gender" {
		t.Errorf("Expected name gender, got %s", form.Survey[0].Name)
	}
	if form.Survey[0].Label != "What is your gender?" {
		t.Errorf("Expected label 'What is your gender?', got %s", form.Survey[0].Label)
	}
	if len(form.Choices) != 2 {
		t.Fatalf("Expected 2 choices, got %d", len(form.Choices))
	}
	if form.Choices[0].ListName != "gender" {
		t.Errorf("Expected list_name 'gender', got %s", form.Choices[0].ListName)
	}
	if form.Choices[0].Name != "1" || form.Choices[0].Label != "Male" {
		t.Errorf("First choice incorrect: %+v", form.Choices[0])
	}
	if form.Choices[1].Name != "2" || form.Choices[1].Label != "Female" {
		t.Errorf("Second choice incorrect: %+v", form.Choices[1])
	}
}

func TestDDIToXLSForm_Integer(t *testing.T) {
	ddiXML := `<var ID="V2" name="age" intrvl="contin">
		<concept>Age of respondent</concept>
		<qstn responseDomainType="numeric">
			<qstnLit>What is your age?</qstnLit>
		</qstn>
		<varFormat type="numeric" schema="other"/>
	</var>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(form.Survey) != 1 {
		t.Fatalf("Expected 1 survey row, got %d", len(form.Survey))
	}
	if form.Survey[0].Type != "integer" {
		t.Errorf("Expected type integer, got %s", form.Survey[0].Type)
	}
	if form.Survey[0].Name != "age" {
		t.Errorf("Expected name age, got %s", form.Survey[0].Name)
	}
	if form.Survey[0].Label != "What is your age?" {
		t.Errorf("Expected label 'What is your age?', got %s", form.Survey[0].Label)
	}
	if len(form.Choices) != 0 {
		t.Errorf("Expected 0 choices, got %d", len(form.Choices))
	}
}

func TestDDIToXLSForm_Text(t *testing.T) {
	ddiXML := `<var ID="V3" name="comments" intrvl="discrete">
		<concept>Additional comments</concept>
		<qstn responseDomainType="text">
			<qstnLit>Please provide any additional comments</qstnLit>
		</qstn>
		<varFormat type="character" schema="other"/>
	</var>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if form.Survey[0].Type != "text" {
		t.Errorf("Expected type text, got %s", form.Survey[0].Type)
	}
}

func TestDDIToXLSForm_SelectMultiple(t *testing.T) {
	// Correct DDI format: varGrp type="multipleResp" + binary vars wrapped in <dataDscr>
	ddiXML := `<dataDscr>
		<varGrp ID="VG1" name="hobbies" type="multipleResp" var="V1 V2 V3">
			<txt>What are your hobbies? (Select all that apply)</txt>
			<concept>What are your hobbies? (Select all that apply)</concept>
		</varGrp>
		<var ID="V1" name="hobbies_reading" intrvl="discrete">
			<qstn responseDomainType="multiple">
				<preQTxt>What are your hobbies? (Select all that apply)</preQTxt>
				<qstnLit>Reading</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<concept>Hobbies: Reading</concept>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V2" name="hobbies_sports" intrvl="discrete">
			<qstn responseDomainType="multiple">
				<preQTxt>What are your hobbies? (Select all that apply)</preQTxt>
				<qstnLit>Sports</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<concept>Hobbies: Sports</concept>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V3" name="hobbies_music" intrvl="discrete">
			<qstn responseDomainType="multiple">
				<preQTxt>What are your hobbies? (Select all that apply)</preQTxt>
				<qstnLit>Music</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<concept>Hobbies: Music</concept>
			<varFormat type="numeric" schema="other"/>
		</var>
	</dataDscr>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(form.Survey) != 1 {
		t.Fatalf("Expected 1 survey row, got %d", len(form.Survey))
	}
	if form.Survey[0].Type != "select_multiple hobbies" {
		t.Errorf("Expected type 'select_multiple hobbies', got %s", form.Survey[0].Type)
	}
	if form.Survey[0].Label != "What are your hobbies? (Select all that apply)" {
		t.Errorf("Expected label from varGrp txt, got %s", form.Survey[0].Label)
	}
	if len(form.Choices) != 3 {
		t.Fatalf("Expected 3 choices, got %d", len(form.Choices))
	}
	expectedChoices := []struct{ name, label string }{
		{"reading", "Reading"},
		{"sports", "Sports"},
		{"music", "Music"},
	}
	for i, exp := range expectedChoices {
		if form.Choices[i].ListName != "hobbies" {
			t.Errorf("Choice %d: expected list_name 'hobbies', got %s", i, form.Choices[i].ListName)
		}
		if form.Choices[i].Name != exp.name {
			t.Errorf("Choice %d: expected name %s, got %s", i, exp.name, form.Choices[i].Name)
		}
		if form.Choices[i].Label != exp.label {
			t.Errorf("Choice %d: expected label %s, got %s", i, exp.label, form.Choices[i].Label)
		}
	}
}

func TestDDIToXLSForm_VarGrp(t *testing.T) {
	ddiXML := `<varGrp ID="VG1" name="satisfaction_group" type="grid" var="V1 V2 V3">
		<concept>Satisfaction questions</concept>
		<txt>Please rate your satisfaction with the following</txt>
	</varGrp>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(form.Survey) != 2 {
		t.Fatalf("Expected 2 survey rows (begin_group + end_group), got %d", len(form.Survey))
	}
	if form.Survey[0].Type != "begin_group" {
		t.Errorf("Expected type begin_group, got %s", form.Survey[0].Type)
	}
	if form.Survey[0].Name != "satisfaction_group" {
		t.Errorf("Expected name satisfaction_group, got %s", form.Survey[0].Name)
	}
	if form.Survey[0].Label != "Satisfaction questions" {
		t.Errorf("Expected label 'Satisfaction questions', got %s", form.Survey[0].Label)
	}
	if form.Survey[1].Type != "end_group" {
		t.Errorf("Expected type end_group, got %s", form.Survey[1].Type)
	}
}

func TestDDIToXLSForm_MissingCategories(t *testing.T) {
	ddiXML := `<var ID="V5" name="satisfaction" intrvl="discrete">
		<concept>Satisfaction</concept>
		<qstn responseDomainType="category">
			<qstnLit>How satisfied are you?</qstnLit>
		</qstn>
		<catgry>
			<catValu>1</catValu>
			<labl>Satisfied</labl>
		</catgry>
		<catgry missing="Y">
			<catValu>-99</catValu>
			<labl>Don't know</labl>
		</catgry>
	</var>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(form.Choices) != 1 {
		t.Errorf("Expected 1 choice (missing categories excluded), got %d", len(form.Choices))
	}
}

func TestDDIToXLSForm_CodeBook(t *testing.T) {
	// Test that a full <codeBook> element can be converted
	ddiXML := `<codeBook xmlns="ddi:codebook:2_5">
		<stdyDscr>
			<citation><titlStmt><titl>Test Study</titl></titlStmt></citation>
		</stdyDscr>
		<dataDscr>
			<var ID="V1" name="age" intrvl="contin">
				<concept>Age</concept>
				<qstn responseDomainType="numeric">
					<qstnLit>How old are you?</qstnLit>
				</qstn>
				<varFormat type="numeric" schema="other"/>
			</var>
		</dataDscr>
	</codeBook>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed for codeBook input: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(form.Survey) != 1 {
		t.Fatalf("Expected 1 survey row, got %d", len(form.Survey))
	}
	if form.Survey[0].Type != "integer" {
		t.Errorf("Expected type integer, got %s", form.Survey[0].Type)
	}
	if form.Survey[0].Name != "age" {
		t.Errorf("Expected name age, got %s", form.Survey[0].Name)
	}
}

func TestDDIToXLSForm_SelectOneFromFile(t *testing.T) {
	ddiXML := `<var ID="V_geburtsland" name="geburtsland" intrvl="discrete">
		<qstn responseDomainType="category">
			<qstnLit>In welchem Land wurden Sie geboren?</qstnLit>
		</qstn>
		<concept vocab="iso_3166_1">In welchem Land wurden Sie geboren?</concept>
		<varFormat type="numeric" schema="other"/>
	</var>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(form.Survey) != 1 {
		t.Fatalf("Expected 1 survey row, got %d", len(form.Survey))
	}
	if form.Survey[0].Type != "select_one_from_file iso_3166_1.csv" {
		t.Errorf("Expected type 'select_one_from_file iso_3166_1.csv', got %s", form.Survey[0].Type)
	}
	if form.Survey[0].Appearance != "minimal" {
		t.Errorf("Expected appearance 'minimal', got %q", form.Survey[0].Appearance)
	}
	if form.Survey[0].Name != "geburtsland" {
		t.Errorf("Expected name geburtsland, got %s", form.Survey[0].Name)
	}
	if len(form.Choices) != 0 {
		t.Errorf("Expected 0 choices (external file), got %d", len(form.Choices))
	}
}

func TestDDIToXLSForm_SelectMultipleFromFile(t *testing.T) {
	ddiXML := `<var ID="V_herkunftslaender" name="herkunftslaender" intrvl="discrete">
		<qstn responseDomainType="multiple">
			<qstnLit>Aus welchen Ländern stammen die Menschen?</qstnLit>
		</qstn>
		<concept vocab="iso_3166_1">Herkunftsländer</concept>
		<varFormat type="numeric" schema="other"/>
	</var>`

	result, err := DDIToXLSForm([]byte(ddiXML))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(form.Survey) != 1 {
		t.Fatalf("Expected 1 survey row, got %d", len(form.Survey))
	}
	if form.Survey[0].Type != "select_multiple_from_file iso_3166_1.csv" {
		t.Errorf("Expected type 'select_multiple_from_file iso_3166_1.csv', got %s", form.Survey[0].Type)
	}
	if form.Survey[0].Appearance != "minimal" {
		t.Errorf("Expected appearance 'minimal', got %q", form.Survey[0].Appearance)
	}
	if len(form.Choices) != 0 {
		t.Errorf("Expected 0 choices (external file), got %d", len(form.Choices))
	}
}

// TestDDIToXLSForm_FlattenedMultipleChoiceOther tests conversion of a
// type="other" group where all member vars (binary + _other text) are direct
// children — no child multipleResp varGrp reference. This is the format
// produced by the exporter after the importer flattens the group hierarchy.
func TestDDIToXLSForm_FlattenedMultipleChoiceOther(t *testing.T) {
	ddi := `<dataDscr>
		<varGrp ID="VG_geschlecht" name="geschlecht" type="other"
			var="V_geschlecht_weiblich V_geschlecht_maennlich V_geschlecht_nicht_binaer V_geschlecht_other">
			<txt>Was ist Ihr Geschlecht?</txt>
			<concept>Geschlecht (Selbstdefinition)</concept>
		</varGrp>
		<var ID="V_geschlecht_weiblich" name="geschlecht_weiblich" intrvl="discrete">
			<qstn responseDomainType="multiple">
				<preQTxt>Was ist Ihr Geschlecht?</preQTxt>
				<qstnLit>weiblich</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<concept>Geschlecht: weiblich</concept>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V_geschlecht_maennlich" name="geschlecht_maennlich" intrvl="discrete">
			<qstn responseDomainType="multiple">
				<preQTxt>Was ist Ihr Geschlecht?</preQTxt>
				<qstnLit>männlich</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<concept>Geschlecht: männlich</concept>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V_geschlecht_nicht_binaer" name="geschlecht_nicht_binaer" intrvl="discrete">
			<qstn responseDomainType="multiple">
				<preQTxt>Was ist Ihr Geschlecht?</preQTxt>
				<qstnLit>nicht-binär</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<concept>Geschlecht: nicht-binär</concept>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V_geschlecht_other" name="geschlecht_other" intrvl="discrete">
			<qstn responseDomainType="text">
				<qstnLit>Geschlecht (eigene Angabe)</qstnLit>
			</qstn>
			<concept>Geschlecht (Freitextangabe)</concept>
			<varFormat type="character" schema="other"/>
		</var>
	</dataDscr>`

	result, err := DDIToXLSForm([]byte(ddi))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	var form XLSForm
	if err := json.Unmarshal(result, &form); err != nil {
		t.Fatalf("Failed to parse XLSForm: %v", err)
	}

	// Should produce: 1 select_multiple + 1 _other text = 2 survey rows
	if len(form.Survey) != 2 {
		t.Fatalf("Expected 2 survey rows, got %d: %+v", len(form.Survey), form.Survey)
	}

	selectRow := form.Survey[0]
	if selectRow.Type != "select_multiple geschlecht" {
		t.Errorf("Expected type 'select_multiple geschlecht', got %s", selectRow.Type)
	}
	if selectRow.Name != "geschlecht" {
		t.Errorf("Expected name 'geschlecht', got %s", selectRow.Name)
	}
	if selectRow.Label != "Was ist Ihr Geschlecht?" {
		t.Errorf("Expected label 'Was ist Ihr Geschlecht?', got %s", selectRow.Label)
	}

	otherRow := form.Survey[1]
	if otherRow.Name != "geschlecht_other" {
		t.Errorf("Expected name 'geschlecht_other', got %s", otherRow.Name)
	}
	if !strings.Contains(otherRow.Relevance, "geschlecht") {
		t.Errorf("Expected relevance referencing geschlecht, got %q", otherRow.Relevance)
	}

	// Choices: 3 binary options + 1 "other" = 4, all sharing list_name "geschlecht"
	if len(form.Choices) != 4 {
		t.Fatalf("Expected 4 choices, got %d: %+v", len(form.Choices), form.Choices)
	}
	for _, c := range form.Choices {
		if c.ListName != "geschlecht" {
			t.Errorf("Expected all choices to have list_name 'geschlecht', got %s for %s", c.ListName, c.Name)
		}
	}
	if form.Choices[3].Name != "other" {
		t.Errorf("Expected last choice name 'other', got %s", form.Choices[3].Name)
	}
}

// TestInvalidInput exercises the converters' malformed-input handling. The
// XLSForm path delegates to the ddi-emitter sidecar, so it skips when the
// sidecar is unreachable (e.g. `go test ./...` without docker-compose).
func TestInvalidInput(t *testing.T) {
	if _, err := DDIToXLSForm([]byte("<invalid>xml</invalid>")); err == nil {
		t.Error("Expected error for invalid DDI XML")
	}

	if _, err := XLSFormToDDI([]byte(`{"survey":[],"choices":[],"settings":{}}`)); err == nil {
		// No sidecar → this would silently succeed? No — our Go validator
		// rejects empty surveys before talking to the sidecar.
		t.Error("Expected error for empty survey sheet")
	}
}

// #22: <ivuInstr> goes to the guidance_hint column. In `parameters` it would
// be "guidance_hint=Bei Unsicherheit nachfragen", which pyxform rejects
// because parameters are space-separated key=value pairs.
func TestDDIToXLSForm_IvuInstrToGuidanceHintColumn(t *testing.T) {
	ddi := `<var ID="V_alter" name="alter">
  <qstn responseDomainType="numeric">
    <preQTxt>In Jahren</preQTxt>
    <qstnLit>Wie alt sind Sie?</qstnLit>
    <ivuInstr>Bei Unsicherheit nachfragen</ivuInstr>
  </qstn>
  <concept>Alter</concept>
</var>`
	out, err := DDIToXLSForm([]byte(ddi))
	if err != nil {
		t.Fatal(err)
	}
	var form XLSForm
	if err := json.Unmarshal(out, &form); err != nil {
		t.Fatal(err)
	}
	if len(form.Survey) != 1 {
		t.Fatalf("expected 1 survey row, got %d", len(form.Survey))
	}
	row := form.Survey[0]
	if row.GuidanceHint != "Bei Unsicherheit nachfragen" {
		t.Errorf("guidance_hint: got %q", row.GuidanceHint)
	}
	if row.Parameters != "" {
		t.Errorf("parameters must stay empty, got %q", row.Parameters)
	}
	if !strings.Contains(string(out), `"guidance_hint"`) {
		t.Errorf("JSON is missing the guidance_hint column: %s", out)
	}
}
