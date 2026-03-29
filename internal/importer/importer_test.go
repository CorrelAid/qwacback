package importer_test

import (
	"os"
	"path/filepath"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// TestSeedDataImport verifies that every XML file in seed_data/ can be
// parsed and imported into a fresh database. This runs during `go test`
// and in the Docker build, catching broken seed data before deployment.
func TestSeedDataImport(t *testing.T) {
	seedFiles, err := filepath.Glob("../../seed_data/*.xml")
	if err != nil {
		t.Fatal(err)
	}
	if len(seedFiles) == 0 {
		t.Fatal("No seed XML files found in seed_data/")
	}

	for _, path := range seedFiles {
		t.Run(filepath.Base(path), func(t *testing.T) {
			dir, err := os.MkdirTemp("", "pb_test_seed_*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

			app, err := tests.NewTestApp(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer app.Cleanup()

			xmlData, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			mv, err := mxj.NewMapXml(xmlData)
			if err != nil {
				t.Fatalf("failed to parse XML: %v", err)
			}

			if err := importer.ImportCodebookData(app, mv, xmlData); err != nil {
				t.Fatalf("failed to import: %v", err)
			}

			studies, _ := app.FindRecordsByFilter("studies", "", "", 0, 0)
			vars, _ := app.FindRecordsByFilter("variables", "", "", 0, 0)

			if len(studies) == 0 {
				t.Error("no studies created after import")
			}
			if len(vars) == 0 {
				t.Error("no variables created after import")
			}

			t.Logf("imported %d studies, %d variables", len(studies), len(vars))
		})
	}
}
