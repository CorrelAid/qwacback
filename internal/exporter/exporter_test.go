package exporter

import (
	"encoding/xml"
	"os"
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
	variable.Set("label", "Variable 1")
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

// TestRoundTripPreservesData imports the prove_it.xml seed data, exports it,
// and verifies that all semantically meaningful data survives the round-trip.
func TestRoundTripPreservesData(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_roundtrip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDataDir)

	testApp, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Cleanup()

	// Import seed data
	xmlData, err := os.ReadFile("../../seed_data/prove_it.xml")
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

	// Find the imported study
	studies, err := testApp.FindRecordsByFilter("studies", "", "", 0, 0)
	if err != nil || len(studies) == 0 {
		t.Fatal("No study found after import")
	}
	study := studies[0]

	// Export
	exported, err := ExportStudyToXML(testApp, study)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Parse exported XML into CodeBook struct
	var cb CodeBook
	if err := xml.Unmarshal(exported, &cb); err != nil {
		t.Fatalf("Failed to parse exported XML: %v\n%s", err, string(exported))
	}

	// --- Study metadata ---
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
		if cb.StdyDscr.Citation.ProdStmt == nil {
			t.Fatal("ProdStmt is nil")
		}
		if cb.StdyDscr.Citation.ProdStmt.Producer.Value != "New Economics Foundation" {
			t.Errorf("Producer: got %q", cb.StdyDscr.Citation.ProdStmt.Producer.Value)
		}
		if cb.StdyDscr.Citation.Holdings == nil {
			t.Fatal("Holdings is nil")
		}
		if cb.StdyDscr.Citation.Holdings.URI != "https://www.nefconsulting.com/what-we-do/evaluation-impact-assessment/prove-it/downloads/" {
			t.Errorf("Holdings URI: got %q", cb.StdyDscr.Citation.Holdings.URI)
		}
		if cb.StdyDscr.StdyInfo.SumDscr.AnlyUnit != "Individuals" {
			t.Errorf("AnlyUnit: got %q", cb.StdyDscr.StdyInfo.SumDscr.AnlyUnit)
		}
		if cb.StdyDscr.StdyInfo.SumDscr.DataKind != "Survey Data" {
			t.Errorf("DataKind: got %q", cb.StdyDscr.StdyInfo.SumDscr.DataKind)
		}
		if cb.StdyDscr.StdyInfo.Subject == nil || len(cb.StdyDscr.StdyInfo.Subject.TopcClas) != 2 {
			t.Errorf("Expected 2 topic classifications, got %v", cb.StdyDscr.StdyInfo.Subject)
		}
		if cb.StdyDscr.StdyInfo.Abstract.Content == "" {
			t.Error("Abstract is empty")
		}
	})

	// --- Variable counts ---
	t.Run("variable_counts", func(t *testing.T) {
		if len(cb.DataDscr.Vars) != 28 {
			t.Errorf("Expected 28 variables, got %d", len(cb.DataDscr.Vars))
		}
	})

	// --- Variable group counts ---
	t.Run("variable_group_counts", func(t *testing.T) {
		if len(cb.DataDscr.VarGrp) != 2 {
			t.Errorf("Expected 2 variable groups, got %d", len(cb.DataDscr.VarGrp))
		}
	})

	// Build lookup maps for detailed checks
	varByName := make(map[string]Var)
	for _, v := range cb.DataDscr.Vars {
		varByName[v.Name] = v
	}
	grpByID := make(map[string]VarGrp)
	for _, g := range cb.DataDscr.VarGrp {
		grpByID[g.ID] = g
	}

	// --- responseDomainType preserved for variables where it is set ---
	t.Run("response_domain_type", func(t *testing.T) {
		// All variables except other_comments have responseDomainType="category".
		// other_comments has responseDomainType="text".
		// Variables with XHTML qstnLit may have Qstn==nil on export if the question
		// text could not be extracted as a plain string.
		exceptions := map[string]string{
			"other_comments": "text",
		}
		for _, v := range cb.DataDscr.Vars {
			if v.Qstn == nil {
				continue // XHTML qstnLit variables may not produce a Qstn element
			}
			expected := "category"
			if e, ok := exceptions[v.Name]; ok {
				expected = e
			}
			if v.Qstn.ResponseDomainType != expected {
				t.Errorf("Variable %s: expected responseDomainType %q, got %q", v.Name, expected, v.Qstn.ResponseDomainType)
			}
		}
	})

	// --- Spot-check a standalone variable (select_one, plain-text qstnLit) ---
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

	// --- Spot-check a grid variable (matrix) with preQTxt ---
	t.Run("grid_variable_contact_community_groups", func(t *testing.T) {
		v, ok := varByName["contact_community_groups"]
		if !ok {
			t.Fatal("Variable contact_community_groups not found")
		}
		if v.ID != "V4a" {
			t.Errorf("DDI ID: got %q", v.ID)
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

	// --- Variable groups ---
	t.Run("variable_groups", func(t *testing.T) {
		vg1, ok := grpByID["VG1"]
		if !ok {
			t.Fatal("VarGrp VG1 not found")
		}
		if vg1.Type != "grid" {
			t.Errorf("VG1 type: got %q", vg1.Type)
		}
		// VG1 has no <labl> element in the current XML — label is expected to be empty.
		if vg1.Labl != "" {
			t.Errorf("VG1 label: expected empty, got %q", vg1.Labl)
		}
		if vg1.Txt != "If you did want to change things around here, do you know who to contact to help you in the following groups…?" {
			t.Errorf("VG1 txt: got %q", vg1.Txt)
		}

		// VG3 (section type) should not be present — section groups are excluded.
		if _, ok := grpByID["VG3"]; ok {
			t.Error("VarGrp VG3 (section) should not be exported")
		}
	})
}
