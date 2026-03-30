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
		if v.Concept.Value != "Interpersonal trust" {
			t.Errorf("Concept: got %q", v.Concept.Value)
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
		if vg1.Concept.Value != "Civic network awareness" {
			t.Errorf("VG1 concept: got %q", vg1.Concept.Value)
		}
		if vg1.Txt != "If you did want to change things around here, do you know who to contact to help you in the following groups…?" {
			t.Errorf("VG1 txt: got %q", vg1.Txt)
		}
	})
}

// TestExportVarGrpCodebookToXML creates a group with member variables and verifies the codebook output.
func TestExportVarGrpCodebookToXML(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_grp_codebook")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDataDir)

	testApp, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Cleanup()

	// Create study
	studyCol, _ := testApp.FindCollectionByNameOrId("studies")
	study := core.NewRecord(studyCol)
	study.Set("title", "Test Study")
	study.Set("abstract", "Test")
	if err := testApp.Save(study); err != nil {
		t.Fatal(err)
	}

	// Create group
	grpCol, _ := testApp.FindCollectionByNameOrId("variable_groups")
	grp := core.NewRecord(grpCol)
	grp.Set("study", study.Id)
	grp.Set("ddi_id", "VG_test")
	grp.Set("name", "test_group")
	grp.Set("type", "grid")
	grp.Set("concept", "Test Group")
	grp.Set("description", "A test group description")
	grp.Set("order", 1)
	if err := testApp.Save(grp); err != nil {
		t.Fatal(err)
	}

	// Create two member variables
	varCol, _ := testApp.FindCollectionByNameOrId("variables")
	for i, name := range []string{"v1", "v2"} {
		v := core.NewRecord(varCol)
		v.Set("study", study.Id)
		v.Set("group", grp.Id)
		v.Set("ddi_id", "V_"+name)
		v.Set("name", name)
		v.Set("concept", "Variable "+name)
		v.Set("question", "Question for "+name)
		v.Set("interval", "discrete")
		v.Set("var_format_type", "numeric")
		v.Set("answer_type", "single_choice")
		v.Set("categories", `[{"value":"1","label":"Yes","is_missing":false},{"value":"2","label":"No","is_missing":false}]`)
		v.Set("order", i+1)
		if err := testApp.Save(v); err != nil {
			t.Fatal(err)
		}
	}

	// Export
	xmlBytes, err := ExportVarGrpCodebookToXML(testApp, grp)
	if err != nil {
		t.Fatalf("ExportVarGrpCodebookToXML failed: %v", err)
	}

	// Parse as DataDscr
	var dd DataDscr
	if err := xml.Unmarshal(xmlBytes, &dd); err != nil {
		t.Fatalf("Failed to parse output: %v\n%s", err, string(xmlBytes))
	}

	// Verify group
	if len(dd.VarGrp) != 1 {
		t.Fatalf("Expected 1 varGrp, got %d", len(dd.VarGrp))
	}
	if dd.VarGrp[0].ID != "VG_test" {
		t.Errorf("VarGrp ID: got %q", dd.VarGrp[0].ID)
	}
	if dd.VarGrp[0].Name != "test_group" {
		t.Errorf("VarGrp Name: got %q", dd.VarGrp[0].Name)
	}
	if dd.VarGrp[0].Type != "grid" {
		t.Errorf("VarGrp Type: got %q", dd.VarGrp[0].Type)
	}
	if dd.VarGrp[0].Concept.Value != "Test Group" {
		t.Errorf("VarGrp Concept: got %q", dd.VarGrp[0].Concept.Value)
	}
	if dd.VarGrp[0].Txt != "A test group description" {
		t.Errorf("VarGrp Txt: got %q", dd.VarGrp[0].Txt)
	}
	if dd.VarGrp[0].Var != "V_v1 V_v2" {
		t.Errorf("VarGrp var attr: got %q", dd.VarGrp[0].Var)
	}

	// Verify member variables
	if len(dd.Vars) != 2 {
		t.Fatalf("Expected 2 vars, got %d", len(dd.Vars))
	}
	for i, v := range dd.Vars {
		expectedName := []string{"v1", "v2"}[i]
		if v.Name != expectedName {
			t.Errorf("Var %d: expected name %q, got %q", i, expectedName, v.Name)
		}
		if v.Qstn == nil {
			t.Fatalf("Var %d: Qstn is nil", i)
		}
		if v.Qstn.ResponseDomainType != "category" {
			t.Errorf("Var %d: expected responseDomainType category, got %q", i, v.Qstn.ResponseDomainType)
		}
		if len(v.Catgry) != 2 {
			t.Errorf("Var %d: expected 2 categories, got %d", i, len(v.Catgry))
		}
	}
}

// TestRoundTripGroupCodebook_ProveIt imports prove_it.xml and verifies group codebook round-trip.
func TestRoundTripGroupCodebook_ProveIt(t *testing.T) {
	testApp, _, _ := roundTripSetup(t, "../../seed_data/prove_it.xml")

	// Find all groups
	groups, err := testApp.FindRecordsByFilter("variable_groups", "", "order", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) == 0 {
		t.Fatal("No variable groups found after import")
	}

	for _, grp := range groups {
		t.Run(grp.GetString("name"), func(t *testing.T) {
			xmlBytes, err := ExportVarGrpCodebookToXML(testApp, grp)
			if err != nil {
				t.Fatalf("ExportVarGrpCodebookToXML failed: %v", err)
			}

			var dd DataDscr
			if err := xml.Unmarshal(xmlBytes, &dd); err != nil {
				t.Fatalf("Failed to parse output: %v\n%s", err, string(xmlBytes))
			}

			if len(dd.VarGrp) != 1 {
				t.Fatalf("Expected 1 varGrp, got %d", len(dd.VarGrp))
			}

			// Verify var attr references match actual vars
			varIDs := strings.Fields(dd.VarGrp[0].Var)
			if len(varIDs) != len(dd.Vars) {
				t.Errorf("var attr has %d IDs but %d vars present", len(varIDs), len(dd.Vars))
			}
			for i, v := range dd.Vars {
				if v.ID != varIDs[i] {
					t.Errorf("Var %d: ID %q not in var attr position %d (%q)", i, v.ID, i, varIDs[i])
				}
			}
		})
	}
}

// TestRoundTripGroupCodebook_SVR imports svr_fb_studie.xml (has multipleResp and grid groups)
// and verifies all group codebook exports.
func TestRoundTripGroupCodebook_SVR(t *testing.T) {
	testApp, _, _ := roundTripSetup(t, "../../seed_data/svr_fb_studie.xml")

	groups, err := testApp.FindRecordsByFilter("variable_groups", "", "order", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) == 0 {
		t.Fatal("No variable groups found after import")
	}

	for _, grp := range groups {
		t.Run(grp.GetString("name"), func(t *testing.T) {
			xmlBytes, err := ExportVarGrpCodebookToXML(testApp, grp)
			if err != nil {
				t.Fatalf("ExportVarGrpCodebookToXML failed: %v", err)
			}

			var dd DataDscr
			if err := xml.Unmarshal(xmlBytes, &dd); err != nil {
				t.Fatalf("Failed to parse output: %v\n%s", err, string(xmlBytes))
			}

			if len(dd.VarGrp) != 1 {
				t.Fatalf("Expected 1 varGrp, got %d", len(dd.VarGrp))
			}
			if dd.VarGrp[0].ID != grp.GetString("ddi_id") {
				t.Errorf("VarGrp ID: expected %q, got %q", grp.GetString("ddi_id"), dd.VarGrp[0].ID)
			}
			if dd.VarGrp[0].Type != grp.GetString("type") {
				t.Errorf("VarGrp Type: expected %q, got %q", grp.GetString("type"), dd.VarGrp[0].Type)
			}

			// Each referenced var ID should have a matching var element
			varIDs := strings.Fields(dd.VarGrp[0].Var)
			if len(varIDs) != len(dd.Vars) {
				t.Errorf("var attr has %d IDs but %d vars present", len(varIDs), len(dd.Vars))
			}
			idSet := make(map[string]bool)
			for _, v := range dd.Vars {
				idSet[v.ID] = true
			}
			for _, id := range varIDs {
				if !idSet[id] {
					t.Errorf("var attr references %q but no matching var element found", id)
				}
			}
		})
	}
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
		if len(cb.DataDscr.Vars) != 45 {
			t.Errorf("Expected 45 variables, got %d", len(cb.DataDscr.Vars))
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
		if v.Concept.Value != "Gender" {
			t.Errorf("Concept: got %q", v.Concept.Value)
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
		if v.Intrvl != "discrete" {
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
