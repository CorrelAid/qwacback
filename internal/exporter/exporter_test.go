package exporter

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestExportStudyToXML(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_export")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDataDir)

	testApp, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Cleanup()

	// 1. Create a dummy study record
	studyCollection, err := testApp.FindCollectionByNameOrId("studies")
	if err != nil {
		t.Fatal(err)
	}
	study := core.NewRecord(studyCollection)
	study.Set("title", "Test Study")
	study.Set("abstract", "This is a test abstract")
	if err := testApp.Save(study); err != nil {
		t.Fatal(err)
	}

	// 2. Create a dummy variable
	varCollection, err := testApp.FindCollectionByNameOrId("variables")
	if err != nil {
		t.Fatal(err)
	}
	variable := core.NewRecord(varCollection)
	variable.Set("study", study.Id)
	variable.Set("name", "v1")
	variable.Set("concept", "Variable 1")
	if err := testApp.Save(variable); err != nil {
		t.Fatal(err)
	}

	// 3. Export to XML
	xmlBytes, err := ExportStudyToXML(testApp, study)
	if err != nil {
		t.Fatalf("Failed to export: %v", err)
	}

	// 4. Check it's well-formed XML
	var doc interface{}
	if err := xml.Unmarshal(xmlBytes, &doc); err != nil {
		t.Errorf("Generated XML is not well-formed: %v\n%s", err, string(xmlBytes))
	}
}

// roundTripSetup imports an XML seed file into a fresh test app and returns the app and first study.
func roundTripSetup(t *testing.T, seedPath string) (*tests.TestApp, *core.Record, CodeBook) {
	t.Helper()

	testDataDir, err := os.MkdirTemp("", "pb_test_roundtrip")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(testDataDir) })

	testApp, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(testApp.Cleanup)

	xmlData, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	mv, err := mxj.NewMapXml(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if err := importer.ImportCodebookData(testApp, mv, xmlData); err != nil {
		t.Fatal(err)
	}

	studies, err := testApp.FindRecordsByFilter("studies", "", "", 0, 0)
	if err != nil || len(studies) == 0 {
		t.Fatal("No study found after import")
	}
	study := studies[0]

	exported, err := ExportStudyToXML(testApp, study)
	if err != nil {
		t.Fatalf("Export failed: %v\n", err)
	}

	var cb CodeBook
	if err := xml.Unmarshal(exported, &cb); err != nil {
		t.Fatalf("Failed to parse exported XML: %v\n%s", err, string(exported))
	}

	return testApp, study, cb
}

// TestRoundTripProveIt imports prove_it.xml and verifies the full round-trip.
func TestRoundTripProveIt(t *testing.T) {
	_, _, cb := roundTripSetup(t, "../../seed_data/prove_it.xml")

	varByName := make(map[string]Var)
	for _, v := range cb.DataDscr.Vars {
		varByName[v.Name] = v
	}
	grpByID := make(map[string]VarGrp)
	for _, g := range cb.DataDscr.VarGrp {
		grpByID[g.ID] = g
	}

	t.Run("study_metadata", func(t *testing.T) {
		if cb.StdyDscr.Citation.TitlStmt.Titl != "Prove It! Toolkit Questionnaire" {
			t.Errorf("Title: got %q", cb.StdyDscr.Citation.TitlStmt.Titl)
		}
		if cb.StdyDscr.Citation.RspStmt == nil {
			t.Fatal("RspStmt is nil")
		}
		if cb.StdyDscr.Citation.RspStmt.AuthEnty.Value != "New Economics Foundation" {
			t.Errorf("Author: got %q", cb.StdyDscr.Citation.RspStmt.AuthEnty.Value)
		}
		if cb.StdyDscr.Citation.RspStmt.AuthEnty.Affiliation != "New Economics Foundation (NEF)" {
			t.Errorf("Author affiliation: got %q", cb.StdyDscr.Citation.RspStmt.AuthEnty.Affiliation)
		}
		if cb.StdyDscr.StdyInfo.SumDscr.AnlyUnit != "Individuals" {
			t.Errorf("AnlyUnit: got %q", cb.StdyDscr.StdyInfo.SumDscr.AnlyUnit)
		}
		if cb.StdyDscr.StdyInfo.SumDscr.DataKind != "Survey Data" {
			t.Errorf("DataKind: got %q", cb.StdyDscr.StdyInfo.SumDscr.DataKind)
		}
		if cb.StdyDscr.StdyInfo.Abstract.Content == "" {
			t.Error("Abstract is empty")
		}
	})

	t.Run("subject_keywords_and_topcClas", func(t *testing.T) {
		if cb.StdyDscr.StdyInfo.Subject == nil {
			t.Fatal("Subject is nil")
		}
		// prove_it.xml has 11 keywords
		if len(cb.StdyDscr.StdyInfo.Subject.Keywords) != 11 {
			t.Errorf("Expected 11 keywords, got %d: %v", len(cb.StdyDscr.StdyInfo.Subject.Keywords), cb.StdyDscr.StdyInfo.Subject.Keywords)
		}
		if cb.StdyDscr.StdyInfo.Subject.Keywords[0] != "social capital" {
			t.Errorf("First keyword: got %q", cb.StdyDscr.StdyInfo.Subject.Keywords[0])
		}
		// prove_it.xml has 2 topcClas: "Impact Assessment" and "Template"
		if len(cb.StdyDscr.StdyInfo.Subject.TopcClas) != 2 {
			t.Errorf("Expected 2 topic classifications, got %d", len(cb.StdyDscr.StdyInfo.Subject.TopcClas))
		}
	})

	t.Run("variable_counts", func(t *testing.T) {
		if len(cb.DataDscr.Vars) != 28 {
			t.Errorf("Expected 28 variables, got %d", len(cb.DataDscr.Vars))
		}
		if len(cb.DataDscr.VarGrp) != 2 {
			t.Errorf("Expected 2 variable groups, got %d", len(cb.DataDscr.VarGrp))
		}
	})

	t.Run("standalone_variable_neighbour_trust", func(t *testing.T) {
		v, ok := varByName["neighbour_trust"]
		if !ok {
			t.Fatal("Variable neighbour_trust not found")
		}
		if v.ID != "V7" {
			t.Errorf("DDI ID: got %q", v.ID)
		}
		if v.Intrvl != "discrete" {
			t.Errorf("Interval: got %q", v.Intrvl)
		}
		if v.Qstn == nil {
			t.Fatal("Qstn is nil")
		}
		if v.Qstn.QstnLit != "Do you think that your neighbours act in your best interests?" {
			t.Errorf("QstnLit: got %q", v.Qstn.QstnLit)
		}
		if v.Concept != "Interpersonal trust" {
			t.Errorf("Concept: got %q", v.Concept)
		}
		if len(v.Catgry) != 3 {
			t.Errorf("Expected 3 categories, got %d", len(v.Catgry))
		} else {
			if v.Catgry[0].CatValu != "1" || v.Catgry[0].Labl != "Yes" {
				t.Errorf("First category: got %+v", v.Catgry[0])
			}
			if v.Catgry[2].CatValu != "3" || v.Catgry[2].Labl != "Don't Know" {
				t.Errorf("Last category: got %+v", v.Catgry[2])
			}
		}
		if v.VarFormat == nil || v.VarFormat.Type != "numeric" {
			t.Errorf("VarFormat: got %+v", v.VarFormat)
		}
	})

	// --- XHTML qstnLit mixed content: text interleaved with child elements ---
	t.Run("xhtml_qstnlit_mixed_content", func(t *testing.T) {
		v, ok := varByName["area_attractiveness"]
		if !ok {
			t.Fatal("Variable area_attractiveness not found")
		}
		if v.Qstn == nil {
			t.Fatal("Qstn is nil")
		}
		// prove_it.xml: "I think that my <PROJECT AREA> is more attractive than it was <TIME PERIOD> ago"
		// The literal angle brackets come from decoded &lt; / &gt; entities inside xhtml:em.
		q := v.Qstn.QstnLit
		for _, want := range []string{"I think that my", "is more attractive than it was", "ago"} {
			if !strings.Contains(q, want) {
				t.Errorf("QstnLit missing %q: got %q", want, q)
			}
		}
	})

	t.Run("grid_variable_with_preQTxt", func(t *testing.T) {
		v, ok := varByName["contact_community_groups"]
		if !ok {
			t.Fatal("Variable contact_community_groups not found")
		}
		if v.Qstn == nil {
			t.Fatal("Qstn is nil")
		}
		if v.Qstn.PreQTxt != "If you did want to change things around here, do you know who to contact to help you in the following groups…?" {
			t.Errorf("PreQTxt: got %q", v.Qstn.PreQTxt)
		}
		if v.Qstn.QstnLit != "Local Community Groups" {
			t.Errorf("QstnLit: got %q", v.Qstn.QstnLit)
		}
		if len(v.Catgry) != 3 {
			t.Errorf("Expected 3 categories, got %d", len(v.Catgry))
		}
	})

	t.Run("variable_groups", func(t *testing.T) {
		vg1, ok := grpByID["VG1"]
		if !ok {
			t.Fatal("VarGrp VG1 not found")
		}
		if vg1.Type != "grid" {
			t.Errorf("VG1 type: got %q", vg1.Type)
		}
		if vg1.Concept != "Civic network awareness" {
			t.Errorf("VG1 concept: got %q", vg1.Concept)
		}
		if vg1.Txt != "If you did want to change things around here, do you know who to contact to help you in the following groups…?" {
			t.Errorf("VG1 txt: got %q", vg1.Txt)
		}
	})
}

// TestRoundTripDemo imports demo.xml (no variable groups, German content) and verifies the round-trip.
func TestRoundTripDemo(t *testing.T) {
	_, _, cb := roundTripSetup(t, "../../seed_data/demo.xml")

	varByName := make(map[string]Var)
	for _, v := range cb.DataDscr.Vars {
		varByName[v.Name] = v
	}

	t.Run("study_metadata", func(t *testing.T) {
		if cb.StdyDscr.Citation.TitlStmt.Titl != "Demographische Standards: Ausgabe 2024" {
			t.Errorf("Title: got %q", cb.StdyDscr.Citation.TitlStmt.Titl)
		}
		if cb.StdyDscr.Citation.TitlStmt.IDNo != "10.21241/ssoar.94099" {
			t.Errorf("IDNo: got %q", cb.StdyDscr.Citation.TitlStmt.IDNo)
		}
		if cb.StdyDscr.StdyInfo.Abstract.Content == "" {
			t.Error("Abstract is empty")
		}
	})

	t.Run("sumDscr_ordering", func(t *testing.T) {
		// These three fields test the corrected SumDscr field ordering:
		// timePrd → nation → (anlyUnit) → universe → dataKind
		if cb.StdyDscr.StdyInfo.SumDscr.TimePrd != "2024" {
			t.Errorf("TimePrd: got %q", cb.StdyDscr.StdyInfo.SumDscr.TimePrd)
		}
		if cb.StdyDscr.StdyInfo.SumDscr.Nation != "Deutschland" {
			t.Errorf("Nation: got %q", cb.StdyDscr.StdyInfo.SumDscr.Nation)
		}
		if cb.StdyDscr.StdyInfo.SumDscr.Universe != "Bevölkerung in Deutschland" {
			t.Errorf("Universe: got %q", cb.StdyDscr.StdyInfo.SumDscr.Universe)
		}
	})

	t.Run("subject_keywords_and_topcClas", func(t *testing.T) {
		if cb.StdyDscr.StdyInfo.Subject == nil {
			t.Fatal("Subject is nil")
		}
		// demo.xml has 11 keywords
		if len(cb.StdyDscr.StdyInfo.Subject.Keywords) != 11 {
			t.Errorf("Expected 11 keywords, got %d: %v", len(cb.StdyDscr.StdyInfo.Subject.Keywords), cb.StdyDscr.StdyInfo.Subject.Keywords)
		}
		if cb.StdyDscr.StdyInfo.Subject.Keywords[0] != "Demografie" {
			t.Errorf("First keyword: got %q", cb.StdyDscr.StdyInfo.Subject.Keywords[0])
		}
		// demo.xml has 1 topcClas: "Template"
		if len(cb.StdyDscr.StdyInfo.Subject.TopcClas) != 1 || cb.StdyDscr.StdyInfo.Subject.TopcClas[0] != "Template" {
			t.Errorf("TopcClas: got %v", cb.StdyDscr.StdyInfo.Subject.TopcClas)
		}
	})

	t.Run("variable_counts", func(t *testing.T) {
		if len(cb.DataDscr.Vars) != 44 {
			t.Errorf("Expected 44 variables, got %d", len(cb.DataDscr.Vars))
		}
		if len(cb.DataDscr.VarGrp) != 0 {
			t.Errorf("Expected 0 variable groups, got %d", len(cb.DataDscr.VarGrp))
		}
	})

	t.Run("spot_check_geschlecht", func(t *testing.T) {
		v, ok := varByName["geschlecht"]
		if !ok {
			t.Fatal("Variable geschlecht not found")
		}
		if v.ID != "V1" {
			t.Errorf("DDI ID: got %q", v.ID)
		}
		if v.Intrvl != "discrete" {
			t.Errorf("Interval: got %q", v.Intrvl)
		}
		if v.Qstn == nil {
			t.Fatal("Qstn is nil")
		}
		if v.Qstn.ResponseDomainType != "category" {
			t.Errorf("ResponseDomainType: got %q", v.Qstn.ResponseDomainType)
		}
		if v.Concept != "Gender" {
			t.Errorf("Concept: got %q", v.Concept)
		}
		if len(v.Catgry) != 3 {
			t.Errorf("Expected 3 categories, got %d", len(v.Catgry))
		}
		if v.VarFormat == nil || v.VarFormat.Type != "numeric" {
			t.Errorf("VarFormat: got %+v", v.VarFormat)
		}
	})

	t.Run("spot_check_open_text_variable", func(t *testing.T) {
		v, ok := varByName["berufliche_taetigkeit"]
		if !ok {
			t.Fatal("Variable berufliche_taetigkeit not found")
		}
		if v.Intrvl != "contin" {
			t.Errorf("Interval: got %q", v.Intrvl)
		}
		if v.Qstn == nil {
			t.Fatal("Qstn is nil")
		}
		if v.Qstn.ResponseDomainType != "text" {
			t.Errorf("ResponseDomainType: got %q", v.Qstn.ResponseDomainType)
		}
		if v.VarFormat == nil || v.VarFormat.Type != "character" {
			t.Errorf("VarFormat: got %+v", v.VarFormat)
		}
	})
}
