package seed

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"qwacback/internal/importer"
	"qwacback/internal/schematron"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const seedData = "../../seed_data"

func newApp(t *testing.T) *tests.TestApp {
	t.Helper()
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	return app
}

// seedDir is a seed_data/ with copies of the named files.
func seedDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		raw, err := os.ReadFile(filepath.Join(seedData, n))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, n), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func studyID(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mv, err := mxj.NewMapXml(raw)
	if err != nil {
		t.Fatal(err)
	}
	return importer.StudyID(mv)
}

func mustSync(t *testing.T, app core.App, dir string) Result {
	t.Helper()
	res, err := Sync(app, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func countVars(t *testing.T, app core.App, studyID string) int64 {
	t.Helper()
	n, err := app.CountRecords("variables", dbx.HashExp{"study": studyID})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// A fresh database ends up with every study in seed_data/, as seed studies.
func TestSyncFreshDatabaseImportsAllSeeds(t *testing.T) {
	app := newApp(t)
	files, _ := filepath.Glob(filepath.Join(seedData, "*.xml"))
	if len(files) == 0 {
		t.Fatal("no seed files")
	}

	res := mustSync(t, app, seedData)
	if len(res.Imported) != len(files) {
		t.Fatalf("imported %d of %d seed files", len(res.Imported), len(files))
	}
	studies, err := app.FindAllRecords("studies")
	if err != nil {
		t.Fatal(err)
	}
	if len(studies) != len(files) {
		t.Fatalf("%d studies for %d seed files", len(studies), len(files))
	}
	for _, s := range studies {
		if s.GetString("source") != importer.SourceSeed || s.GetString("seed_hash") == "" {
			t.Errorf("%q: source %q, seed_hash %q", s.GetString("title"), s.GetString("source"), s.GetString("seed_hash"))
		}
	}

	// A second start changes nothing.
	res = mustSync(t, app, seedData)
	if len(res.Unchanged) != len(files) || len(res.Imported)+len(res.Replaced)+len(res.Deleted) != 0 {
		t.Fatalf("second sync: %+v", res)
	}
}

func TestSyncNewChangedUnchangedRemoved(t *testing.T) {
	app := newApp(t)
	dir := seedDir(t, "demo.xml", "prove_it.xml")
	demoID := studyID(t, filepath.Join(dir, "demo.xml"))
	proveID := studyID(t, filepath.Join(dir, "prove_it.xml"))

	// new
	res := mustSync(t, app, dir)
	if len(res.Imported) != 2 {
		t.Fatalf("first sync: %+v", res)
	}
	demoVars := countVars(t, app, demoID)
	if demoVars == 0 {
		t.Fatal("demo has no variables")
	}

	// unchanged
	res = mustSync(t, app, dir)
	if len(res.Unchanged) != 2 || res.Changed() {
		t.Fatalf("unchanged sync: %+v", res)
	}

	// changed: same title, different bytes
	path := filepath.Join(dir, "demo.xml")
	raw, _ := os.ReadFile(path)
	changed := append(bytes.TrimRight(raw, "\n"), []byte("\n<!-- changed -->\n")...)
	if err := os.WriteFile(path, changed, 0o644); err != nil {
		t.Fatal(err)
	}
	res = mustSync(t, app, dir)
	if len(res.Replaced) != 1 || len(res.Unchanged) != 1 || !res.Changed() {
		t.Fatalf("changed sync: %+v", res)
	}
	demo, err := app.FindRecordById("studies", demoID)
	if err != nil {
		t.Fatalf("replaced study lost its ID: %v", err)
	}
	if demo.GetString("seed_hash") != importer.SeedHash(changed) || demo.GetString("codebook") != string(changed) {
		t.Error("replaced study doesn't carry the new file")
	}
	if n := countVars(t, app, demoID); n != demoVars {
		t.Errorf("replaced study has %d variables, want %d (no duplicates, none lost)", n, demoVars)
	}

	// removed
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	res = mustSync(t, app, dir)
	if len(res.Deleted) != 1 || len(res.Unchanged) != 1 {
		t.Fatalf("removed sync: %+v", res)
	}
	if _, err := app.FindRecordById("studies", demoID); err == nil {
		t.Error("study of the removed seed file still exists")
	}
	if n := countVars(t, app, demoID); n != 0 {
		t.Errorf("%d variables of the removed study left", n)
	}
	if _, err := app.FindRecordById("studies", proveID); err != nil {
		t.Error("the other seed study is gone")
	}
}

func TestSyncLeavesUploadsAlone(t *testing.T) {
	app := newApp(t)

	upload := func(name string) string {
		raw, err := os.ReadFile(filepath.Join(seedData, name))
		if err != nil {
			t.Fatal(err)
		}
		mv, _ := mxj.NewMapXml(raw)
		id, err := importer.ImportCodebook(app, mv, raw)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	snailID := upload("snail.xml") // not a seed file here
	demoID := upload("demo.xml")   // same title as a seed file

	dir := seedDir(t, "demo.xml", "prove_it.xml")
	res := mustSync(t, app, dir)
	if len(res.Imported) != 1 || len(res.Skipped) != 1 || len(res.Deleted) != 0 {
		t.Fatalf("sync: %+v", res)
	}
	for _, id := range []string{snailID, demoID} {
		s, err := app.FindRecordById("studies", id)
		if err != nil {
			t.Fatalf("upload %s is gone: %v", id, err)
		}
		if s.GetString("source") != importer.SourceUpload {
			t.Errorf("upload %q became %q", s.GetString("title"), s.GetString("source"))
		}
	}

	// Removing every seed file deletes the seed study only.
	empty := t.TempDir()
	res = mustSync(t, app, empty)
	if len(res.Deleted) != 1 {
		t.Fatalf("empty sync: %+v", res)
	}
	if n, _ := app.CountRecords("studies"); n != 2 {
		t.Errorf("%d studies left, want the 2 uploads", n)
	}
}

// A missing seed directory leaves the database alone instead of deleting
// every seed study.
func TestSyncMissingDirChangesNothing(t *testing.T) {
	app := newApp(t)
	mustSync(t, app, seedDir(t, "demo.xml"))
	res := mustSync(t, app, filepath.Join(t.TempDir(), "nope"))
	if len(res.Deleted) != 0 {
		t.Fatalf("sync: %+v", res)
	}
	if n, _ := app.CountRecords("studies"); n != 1 {
		t.Errorf("%d studies, want 1", n)
	}
}

type rejectAll struct{}

func (rejectAll) WaitForWorker(time.Duration) error { return nil }
func (rejectAll) Validate([]byte) (*schematron.ValidationResponse, error) {
	return &schematron.ValidationResponse{Valid: false}, nil
}
func (rejectAll) Close() {}

// A seed the worker rejects fails the sync before anything is written.
func TestSyncRejectedSeedFails(t *testing.T) {
	app := newApp(t)
	if _, err := Sync(app, seedDir(t, "demo.xml"), rejectAll{}); err == nil {
		t.Fatal("sync accepted a rejected seed")
	}
	if n, _ := app.CountRecords("studies"); n != 0 {
		t.Errorf("%d studies imported", n)
	}
}
