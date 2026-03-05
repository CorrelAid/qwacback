package converter

import (
	"encoding/json"
	"encoding/xml"
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
	ddiXML := `<var ID="V2" name="age" intrvl="discrete">
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
	ddiXML := `<var ID="V3" name="comments" intrvl="contin">
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
			<catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
			<catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
			<concept>Hobbies: Reading</concept>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V2" name="hobbies_sports" intrvl="discrete">
			<qstn responseDomainType="multiple">
				<preQTxt>What are your hobbies? (Select all that apply)</preQTxt>
				<qstnLit>Sports</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
			<catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
			<concept>Hobbies: Sports</concept>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V3" name="hobbies_music" intrvl="discrete">
			<qstn responseDomainType="multiple">
				<preQTxt>What are your hobbies? (Select all that apply)</preQTxt>
				<qstnLit>Music</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
			<catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
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

func TestXLSFormToDDI_SelectOne(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "select_one gender", "name": "gender", "label": "What is your gender?"}
		],
		"choices": [
			{"list_name": "gender", "name": "1", "label": "Male"},
			{"list_name": "gender", "name": "2", "label": "Female"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	var v DDIVar
	if err := xml.Unmarshal(ddiXML, &v); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if v.Name != "gender" {
		t.Errorf("Expected name gender, got %s", v.Name)
	}
	if v.Qstn == nil {
		t.Fatal("Qstn is nil")
	}
	if v.Qstn.ResponseDomainType != "category" {
		t.Errorf("Expected responseDomainType category, got %s", v.Qstn.ResponseDomainType)
	}
	if v.Intrvl != "discrete" {
		t.Errorf("Expected intrvl discrete, got %s", v.Intrvl)
	}
	if len(v.Catgry) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(v.Catgry))
	}
	if v.VarFormat == nil || v.VarFormat.Type != "numeric" {
		t.Errorf("VarFormat incorrect: %+v", v.VarFormat)
	}
}

func TestXLSFormToDDI_Integer(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "integer", "name": "age", "label": "What is your age?"}
		],
		"choices": [],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	var v DDIVar
	if err := xml.Unmarshal(ddiXML, &v); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if v.Name != "age" {
		t.Errorf("Expected name age, got %s", v.Name)
	}
	if v.Qstn.ResponseDomainType != "numeric" {
		t.Errorf("Expected responseDomainType numeric, got %s", v.Qstn.ResponseDomainType)
	}
	if v.Intrvl != "discrete" {
		t.Errorf("Expected intrvl discrete, got %s", v.Intrvl)
	}
	if v.VarFormat == nil || v.VarFormat.Type != "numeric" {
		t.Errorf("VarFormat incorrect: %+v", v.VarFormat)
	}
}

func TestXLSFormToDDI_Text(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "text", "name": "comments", "label": "Please provide any additional comments", "hint": "Be specific"}
		],
		"choices": [],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	var v DDIVar
	if err := xml.Unmarshal(ddiXML, &v); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if v.Name != "comments" {
		t.Errorf("Expected name comments, got %s", v.Name)
	}
	if v.Qstn.ResponseDomainType != "text" {
		t.Errorf("Expected responseDomainType text, got %s", v.Qstn.ResponseDomainType)
	}
	if v.Qstn.PreQTxt != "Be specific" {
		t.Errorf("Expected PreQTxt 'Be specific', got %s", v.Qstn.PreQTxt)
	}
	if v.Intrvl != "contin" {
		t.Errorf("Expected intrvl contin, got %s", v.Intrvl)
	}
	if v.VarFormat == nil || v.VarFormat.Type != "character" {
		t.Errorf("VarFormat incorrect: %+v", v.VarFormat)
	}
}

func TestXLSFormToDDI_SelectMultiple(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "select_multiple hobbies", "name": "hobbies", "label": "What are your hobbies?"}
		],
		"choices": [
			{"list_name": "hobbies", "name": "1", "label": "Reading"},
			{"list_name": "hobbies", "name": "2", "label": "Sports"},
			{"list_name": "hobbies", "name": "3", "label": "Music"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	// Should produce <dataDscr> with varGrp + binary vars
	var dd DDIDataDscr
	if err := xml.Unmarshal(ddiXML, &dd); err != nil {
		t.Fatalf("Failed to parse result as dataDscr: %v", err)
	}

	// Check varGrp
	if len(dd.VarGrps) != 1 {
		t.Fatalf("Expected 1 varGrp, got %d", len(dd.VarGrps))
	}
	grp := dd.VarGrps[0]
	if grp.Type != "multipleResp" {
		t.Errorf("Expected varGrp type multipleResp, got %s", grp.Type)
	}
	if grp.Name != "hobbies" {
		t.Errorf("Expected varGrp name hobbies, got %s", grp.Name)
	}
	if grp.Txt != "What are your hobbies?" {
		t.Errorf("Expected varGrp txt, got %s", grp.Txt)
	}

	// Check binary vars
	if len(dd.Vars) != 3 {
		t.Fatalf("Expected 3 binary vars, got %d", len(dd.Vars))
	}

	expectedVars := []struct{ name, qstnLit string }{
		{"hobbies_1", "Reading"},
		{"hobbies_2", "Sports"},
		{"hobbies_3", "Music"},
	}
	for i, exp := range expectedVars {
		v := dd.Vars[i]
		if v.Name != exp.name {
			t.Errorf("Var %d: expected name %s, got %s", i, exp.name, v.Name)
		}
		if v.Qstn == nil {
			t.Fatalf("Var %d: Qstn is nil", i)
		}
		if v.Qstn.ResponseDomainType != "multiple" {
			t.Errorf("Var %d: expected responseDomainType multiple, got %s", i, v.Qstn.ResponseDomainType)
		}
		if v.Qstn.QstnLit != exp.qstnLit {
			t.Errorf("Var %d: expected qstnLit %s, got %s", i, exp.qstnLit, v.Qstn.QstnLit)
		}
		if v.Qstn.PreQTxt != "What are your hobbies?" {
			t.Errorf("Var %d: expected preQTxt to match group txt, got %s", i, v.Qstn.PreQTxt)
		}
		if len(v.Catgry) != 2 {
			t.Errorf("Var %d: expected 2 binary categories, got %d", i, len(v.Catgry))
		}
		if v.Catgry[0].CatValu != "0" || v.Catgry[1].CatValu != "1" {
			t.Errorf("Var %d: expected binary categories 0/1", i)
		}
	}
}

func TestXLSFormToDDI_Bildungsgrad(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "select_one bildungsgrad", "name": "bildungsgrad", "label": "Was ist Ihr höchster Bildungsabschluss?"}
		],
		"choices": [
			{"list_name": "bildungsgrad", "name": "1", "label": "Kein Abschluss"},
			{"list_name": "bildungsgrad", "name": "2", "label": "Haupt- oder Realschulabschluss"},
			{"list_name": "bildungsgrad", "name": "3", "label": "Fachhochschulreife / Abitur"},
			{"list_name": "bildungsgrad", "name": "4", "label": "Abgeschlossene Berufsausbildung"},
			{"list_name": "bildungsgrad", "name": "5", "label": "Hochschulabschluss"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	var v DDIVar
	if err := xml.Unmarshal(ddiXML, &v); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if v.ID != "V_bildungsgrad" {
		t.Errorf("Expected ID V_bildungsgrad, got %s", v.ID)
	}
	if v.Name != "bildungsgrad" {
		t.Errorf("Expected name bildungsgrad, got %s", v.Name)
	}
	if v.Intrvl != "discrete" {
		t.Errorf("Expected intrvl discrete, got %s", v.Intrvl)
	}
	if v.Qstn == nil {
		t.Fatal("Qstn is nil")
	}
	if v.Qstn.ResponseDomainType != "category" {
		t.Errorf("Expected responseDomainType category, got %s", v.Qstn.ResponseDomainType)
	}
	if v.Qstn.QstnLit != "Was ist Ihr höchster Bildungsabschluss?" {
		t.Errorf("Expected QstnLit 'Was ist Ihr höchster Bildungsabschluss?', got %s", v.Qstn.QstnLit)
	}
	if v.Concept != "Was ist Ihr höchster Bildungsabschluss?" {
		t.Errorf("Expected concept to match label, got %s", v.Concept)
	}
	if len(v.Catgry) != 5 {
		t.Fatalf("Expected 5 categories, got %d", len(v.Catgry))
	}

	expectedChoices := []struct{ value, label string }{
		{"1", "Kein Abschluss"},
		{"2", "Haupt- oder Realschulabschluss"},
		{"3", "Fachhochschulreife / Abitur"},
		{"4", "Abgeschlossene Berufsausbildung"},
		{"5", "Hochschulabschluss"},
	}
	for i, exp := range expectedChoices {
		if v.Catgry[i].CatValu != exp.value {
			t.Errorf("Category %d: expected value %s, got %s", i, exp.value, v.Catgry[i].CatValu)
		}
		if v.Catgry[i].Labl != exp.label {
			t.Errorf("Category %d: expected label %s, got %s", i, exp.label, v.Catgry[i].Labl)
		}
	}
	if v.VarFormat == nil || v.VarFormat.Type != "numeric" {
		t.Errorf("VarFormat incorrect: %+v", v.VarFormat)
	}
}

func TestXLSFormToDDI_SelectMultipleGeraete(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "select_multiple geraete", "name": "geraetebesitz", "label": "Welche dieser Geräte besitzen Sie?"}
		],
		"choices": [
			{"list_name": "geraete", "name": "smartphone", "label": "Smartphone"},
			{"list_name": "geraete", "name": "laptop", "label": "Laptop"},
			{"list_name": "geraete", "name": "tablet", "label": "Tablet"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	// Should produce <dataDscr> with varGrp + binary vars
	var dd DDIDataDscr
	if err := xml.Unmarshal(ddiXML, &dd); err != nil {
		t.Fatalf("Failed to parse result as dataDscr: %v", err)
	}

	// Check varGrp
	if len(dd.VarGrps) != 1 {
		t.Fatalf("Expected 1 varGrp, got %d", len(dd.VarGrps))
	}
	grp := dd.VarGrps[0]
	if grp.ID != "VG_geraetebesitz" {
		t.Errorf("Expected varGrp ID VG_geraetebesitz, got %s", grp.ID)
	}
	if grp.Name != "geraetebesitz" {
		t.Errorf("Expected varGrp name geraetebesitz, got %s", grp.Name)
	}
	if grp.Type != "multipleResp" {
		t.Errorf("Expected varGrp type multipleResp, got %s", grp.Type)
	}
	if grp.Txt != "Welche dieser Geräte besitzen Sie?" {
		t.Errorf("Expected varGrp txt, got %s", grp.Txt)
	}
	if grp.Concept != "Welche dieser Geräte besitzen Sie?" {
		t.Errorf("Expected varGrp concept, got %s", grp.Concept)
	}

	// Check binary vars
	if len(dd.Vars) != 3 {
		t.Fatalf("Expected 3 binary vars, got %d", len(dd.Vars))
	}

	expectedVars := []struct{ id, name, qstnLit, concept string }{
		{"V_geraetebesitz_smartphone", "geraetebesitz_smartphone", "Smartphone", "Welche dieser Geräte besitzen Sie?: Smartphone"},
		{"V_geraetebesitz_laptop", "geraetebesitz_laptop", "Laptop", "Welche dieser Geräte besitzen Sie?: Laptop"},
		{"V_geraetebesitz_tablet", "geraetebesitz_tablet", "Tablet", "Welche dieser Geräte besitzen Sie?: Tablet"},
	}
	for i, exp := range expectedVars {
		v := dd.Vars[i]
		if v.ID != exp.id {
			t.Errorf("Var %d: expected ID %s, got %s", i, exp.id, v.ID)
		}
		if v.Name != exp.name {
			t.Errorf("Var %d: expected name %s, got %s", i, exp.name, v.Name)
		}
		if v.Intrvl != "discrete" {
			t.Errorf("Var %d: expected intrvl discrete, got %s", i, v.Intrvl)
		}
		if v.Qstn == nil {
			t.Fatalf("Var %d: Qstn is nil", i)
		}
		if v.Qstn.ResponseDomainType != "multiple" {
			t.Errorf("Var %d: expected responseDomainType multiple, got %s", i, v.Qstn.ResponseDomainType)
		}
		if v.Qstn.PreQTxt != "Welche dieser Geräte besitzen Sie?" {
			t.Errorf("Var %d: expected preQTxt to match group txt, got %s", i, v.Qstn.PreQTxt)
		}
		if v.Qstn.QstnLit != exp.qstnLit {
			t.Errorf("Var %d: expected qstnLit %s, got %s", i, exp.qstnLit, v.Qstn.QstnLit)
		}
		if v.Concept != exp.concept {
			t.Errorf("Var %d: expected concept %s, got %s", i, exp.concept, v.Concept)
		}
		// Binary categories: 0=Not mentioned, 1=Mentioned
		if len(v.Catgry) != 2 {
			t.Fatalf("Var %d: expected 2 binary categories, got %d", i, len(v.Catgry))
		}
		if v.Catgry[0].CatValu != "0" || v.Catgry[0].Labl != "Not mentioned" {
			t.Errorf("Var %d: first category incorrect: %+v", i, v.Catgry[0])
		}
		if v.Catgry[1].CatValu != "1" || v.Catgry[1].Labl != "Mentioned" {
			t.Errorf("Var %d: second category incorrect: %+v", i, v.Catgry[1])
		}
		if v.VarFormat == nil || v.VarFormat.Type != "numeric" {
			t.Errorf("Var %d: VarFormat incorrect: %+v", i, v.VarFormat)
		}
	}

	// Check varGrp var attribute lists all member IDs
	expectedVarAttr := "V_geraetebesitz_smartphone V_geraetebesitz_laptop V_geraetebesitz_tablet"
	if grp.Var != expectedVarAttr {
		t.Errorf("Expected varGrp var attr %q, got %q", expectedVarAttr, grp.Var)
	}
}

func TestXLSFormToDDI_Group(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "begin_group", "name": "satisfaction_group", "label": "Satisfaction questions"},
			{"type": "end_group", "name": ""}
		],
		"choices": [],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	var vg DDIVarGrp
	if err := xml.Unmarshal(ddiXML, &vg); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if vg.Name != "satisfaction_group" {
		t.Errorf("Expected name satisfaction_group, got %s", vg.Name)
	}
	if vg.Concept != "Satisfaction questions" {
		t.Errorf("Expected concept 'Satisfaction questions', got %s", vg.Concept)
	}
	if !strings.HasPrefix(vg.ID, "VG_") {
		t.Errorf("Expected ID to start with VG_, got %s", vg.ID)
	}
}

func TestXLSFormToDDI_GuidanceHint(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "integer", "name": "age", "label": "What is your age?", "parameters": "guidance_hint=Ask politely"}
		],
		"choices": [],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	var v DDIVar
	if err := xml.Unmarshal(ddiXML, &v); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if v.Qstn.IvuInstr != "Ask politely" {
		t.Errorf("Expected IvuInstr 'Ask politely', got %s", v.Qstn.IvuInstr)
	}
}

func TestRoundTrip_SelectOne(t *testing.T) {
	// Start with DDI XML
	originalDDI := `<var ID="V1" name="gender" intrvl="discrete">
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

	// Convert to XLSForm
	xlsformJSON, err := DDIToXLSForm([]byte(originalDDI))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	// Convert back to DDI
	newDDI, err := XLSFormToDDI(xlsformJSON)
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	// Parse both DDI documents
	var originalVar DDIVar
	if err := xml.Unmarshal([]byte(originalDDI), &originalVar); err != nil {
		t.Fatalf("Failed to parse original DDI: %v", err)
	}

	var newVar DDIVar
	if err := xml.Unmarshal(newDDI, &newVar); err != nil {
		t.Fatalf("Failed to parse new DDI: %v", err)
	}

	// Compare key fields
	if originalVar.Name != newVar.Name {
		t.Errorf("Name mismatch: %s vs %s", originalVar.Name, newVar.Name)
	}
	if originalVar.Intrvl != newVar.Intrvl {
		t.Errorf("Intrvl mismatch: %s vs %s", originalVar.Intrvl, newVar.Intrvl)
	}
	if originalVar.Qstn.ResponseDomainType != newVar.Qstn.ResponseDomainType {
		t.Errorf("ResponseDomainType mismatch: %s vs %s",
			originalVar.Qstn.ResponseDomainType, newVar.Qstn.ResponseDomainType)
	}
	if len(originalVar.Catgry) != len(newVar.Catgry) {
		t.Errorf("Category count mismatch: %d vs %d",
			len(originalVar.Catgry), len(newVar.Catgry))
	}
}

func TestXLSFormToDDI_SelectOneWithOther(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "select_one geschlecht", "name": "geschlecht", "label": "Welches Geschlecht haben Sie?"},
			{"type": "text", "name": "geschlecht_other", "label": "Sonstiges (bitte angeben)", "relevance": "${geschlecht} = 'sonstiges'"}
		],
		"choices": [
			{"list_name": "geschlecht", "name": "maennlich", "label": "Männlich"},
			{"list_name": "geschlecht", "name": "weiblich", "label": "Weiblich"},
			{"list_name": "geschlecht", "name": "divers", "label": "Divers"},
			{"list_name": "geschlecht", "name": "sonstiges", "label": "Sonstiges"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	// Should produce <dataDscr> with two vars: the select_one + the text other
	var dd DDIDataDscr
	if err := xml.Unmarshal(ddiXML, &dd); err != nil {
		t.Fatalf("Failed to parse result as dataDscr: %v", err)
	}

	if len(dd.Vars) != 2 {
		t.Fatalf("Expected 2 vars, got %d", len(dd.Vars))
	}

	// First var: the categorical select_one
	selectVar := dd.Vars[0]
	if selectVar.Name != "geschlecht" {
		t.Errorf("Expected name geschlecht, got %s", selectVar.Name)
	}
	if selectVar.Qstn == nil {
		t.Fatal("selectVar Qstn is nil")
	}
	if selectVar.Qstn.ResponseDomainType != "category" {
		t.Errorf("Expected responseDomainType category, got %s", selectVar.Qstn.ResponseDomainType)
	}
	if selectVar.Qstn.QstnLit != "Welches Geschlecht haben Sie?" {
		t.Errorf("Expected QstnLit, got %s", selectVar.Qstn.QstnLit)
	}
	if len(selectVar.Catgry) != 4 {
		t.Fatalf("Expected 4 categories, got %d", len(selectVar.Catgry))
	}
	expectedChoices := []struct{ value, label string }{
		{"maennlich", "Männlich"},
		{"weiblich", "Weiblich"},
		{"divers", "Divers"},
		{"sonstiges", "Sonstiges"},
	}
	for i, exp := range expectedChoices {
		if selectVar.Catgry[i].CatValu != exp.value || selectVar.Catgry[i].Labl != exp.label {
			t.Errorf("Category %d: expected %s/%s, got %+v", i, exp.value, exp.label, selectVar.Catgry[i])
		}
	}

	// Second var: the text other specification
	otherVar := dd.Vars[1]
	if otherVar.Name != "geschlecht_other" {
		t.Errorf("Expected name geschlecht_other, got %s", otherVar.Name)
	}
	if otherVar.Qstn == nil {
		t.Fatal("otherVar Qstn is nil")
	}
	if otherVar.Qstn.ResponseDomainType != "text" {
		t.Errorf("Expected responseDomainType text, got %s", otherVar.Qstn.ResponseDomainType)
	}
	if otherVar.Qstn.QstnLit != "Sonstiges (bitte angeben)" {
		t.Errorf("Expected QstnLit 'Sonstiges (bitte angeben)', got %s", otherVar.Qstn.QstnLit)
	}
	if otherVar.Intrvl != "contin" {
		t.Errorf("Expected intrvl contin, got %s", otherVar.Intrvl)
	}
	if otherVar.VarFormat == nil || otherVar.VarFormat.Type != "character" {
		t.Errorf("Expected VarFormat character, got %+v", otherVar.VarFormat)
	}

	t.Logf("DDI output:\n%s", string(ddiXML))
}

func TestXLSFormToDDI_SelectMultipleWithOther(t *testing.T) {
	xlsformJSON := `{
		"survey": [
			{"type": "select_multiple geraete", "name": "geraetebesitz", "label": "Welche dieser Geräte besitzen Sie?"},
			{"type": "text", "name": "geraetebesitz_other", "label": "Sonstiges (bitte angeben)", "relevance": "selected(${geraetebesitz}, 'sonstiges')"}
		],
		"choices": [
			{"list_name": "geraete", "name": "smartphone", "label": "Smartphone"},
			{"list_name": "geraete", "name": "laptop", "label": "Laptop"},
			{"list_name": "geraete", "name": "tablet", "label": "Tablet"},
			{"list_name": "geraete", "name": "sonstiges", "label": "Sonstiges"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	var dd DDIDataDscr
	if err := xml.Unmarshal(ddiXML, &dd); err != nil {
		t.Fatalf("Failed to parse result as dataDscr: %v", err)
	}

	// Should have 1 varGrp (multipleResp) + 4 binary vars + 1 text var = 5 vars
	if len(dd.VarGrps) != 1 {
		t.Fatalf("Expected 1 varGrp, got %d", len(dd.VarGrps))
	}
	if len(dd.Vars) != 5 {
		t.Fatalf("Expected 5 vars (4 binary + 1 text), got %d", len(dd.Vars))
	}

	// Check varGrp
	grp := dd.VarGrps[0]
	if grp.Type != "multipleResp" {
		t.Errorf("Expected varGrp type multipleResp, got %s", grp.Type)
	}
	if grp.Name != "geraetebesitz" {
		t.Errorf("Expected varGrp name geraetebesitz, got %s", grp.Name)
	}

	// Check binary vars (first 4)
	expectedBinary := []struct{ name, qstnLit string }{
		{"geraetebesitz_smartphone", "Smartphone"},
		{"geraetebesitz_laptop", "Laptop"},
		{"geraetebesitz_tablet", "Tablet"},
		{"geraetebesitz_sonstiges", "Sonstiges"},
	}
	for i, exp := range expectedBinary {
		v := dd.Vars[i]
		if v.Name != exp.name {
			t.Errorf("Var %d: expected name %s, got %s", i, exp.name, v.Name)
		}
		if v.Qstn.ResponseDomainType != "multiple" {
			t.Errorf("Var %d: expected responseDomainType multiple, got %s", i, v.Qstn.ResponseDomainType)
		}
		if v.Qstn.QstnLit != exp.qstnLit {
			t.Errorf("Var %d: expected qstnLit %s, got %s", i, exp.qstnLit, v.Qstn.QstnLit)
		}
		if len(v.Catgry) != 2 {
			t.Errorf("Var %d: expected 2 binary categories, got %d", i, len(v.Catgry))
		}
	}

	// Check text other var (last one)
	otherVar := dd.Vars[4]
	if otherVar.Name != "geraetebesitz_other" {
		t.Errorf("Expected name geraetebesitz_other, got %s", otherVar.Name)
	}
	if otherVar.Qstn.ResponseDomainType != "text" {
		t.Errorf("Expected responseDomainType text, got %s", otherVar.Qstn.ResponseDomainType)
	}
	if otherVar.Qstn.QstnLit != "Sonstiges (bitte angeben)" {
		t.Errorf("Expected QstnLit 'Sonstiges (bitte angeben)', got %s", otherVar.Qstn.QstnLit)
	}
	if otherVar.Intrvl != "contin" {
		t.Errorf("Expected intrvl contin, got %s", otherVar.Intrvl)
	}
	if otherVar.VarFormat == nil || otherVar.VarFormat.Type != "character" {
		t.Errorf("Expected VarFormat character, got %+v", otherVar.VarFormat)
	}

	t.Logf("DDI output:\n%s", string(ddiXML))
}

func TestInvalidInput(t *testing.T) {
	// Test invalid DDI XML
	_, err := DDIToXLSForm([]byte("<invalid>xml</invalid>"))
	if err == nil {
		t.Error("Expected error for invalid DDI XML")
	}

	// Test invalid XLSForm JSON
	_, err = XLSFormToDDI([]byte(`{"invalid": "json"}`))
	if err == nil {
		t.Error("Expected error for invalid XLSForm JSON")
	}
}
