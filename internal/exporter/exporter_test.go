package exporter

import (
	"encoding/xml"
	"os"
	"testing"

	_ "qwacback/migrations"

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
