package routes

import (
	"os"
	"testing"

	"qwacback/internal/importer"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// The study filters must work with several IDs at once. The first version
// used `id ~ 'a|b'`, which PocketBase treats as LIKE, so two study_id
// values matched nothing.
func TestSearchQuestions_StudyFilters(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	for _, f := range []string{"prove_it.xml", "demo.xml", "svr_fb_studie.xml"} {
		xmlData, err := os.ReadFile("../../seed_data/" + f)
		if err != nil {
			t.Fatal(err)
		}
		mv, err := mxj.NewMapXml(xmlData)
		if err != nil {
			t.Fatal(err)
		}
		if err := importer.ImportCodebookData(app, mv, xmlData); err != nil {
			t.Fatalf("import %s: %v", f, err)
		}
	}
	studies, err := app.FindRecordsByFilter("studies", "", "", 0, 0)
	if err != nil || len(studies) < 3 {
		t.Fatalf("expected 3 studies, got %d (%v)", len(studies), err)
	}

	// A query that hits one question per study: their variable names.
	var ids []string
	q := ""
	for _, s := range studies[:3] {
		qs, err := AssembleQuestions(app, s.Id)
		if err != nil || len(qs) == 0 {
			t.Fatalf("study %s has no questions (%v)", s.Id, err)
		}
		ids = append(ids, s.Id)
		q += qs[0].Name + " "
	}

	studiesIn := func(qs []Question) map[string]bool {
		out := map[string]bool{}
		for _, x := range qs {
			out[x.StudyID] = true
		}
		return out
	}
	search := func(include, exclude []string) map[string]bool {
		t.Helper()
		got, err := SearchQuestions(app, q, include, exclude)
		if err != nil {
			t.Fatal(err)
		}
		return studiesIn(got)
	}

	if got := search(nil, nil); !got[ids[0]] || !got[ids[1]] || !got[ids[2]] {
		t.Errorf("no filter: expected hits from all three studies, got %v", got)
	}
	if got := search(ids[:2], nil); !got[ids[0]] || !got[ids[1]] || got[ids[2]] {
		t.Errorf("study_id=%v: expected hits from exactly those, got %v", ids[:2], got)
	}
	if got := search(nil, ids[:1]); got[ids[0]] || !got[ids[1]] || !got[ids[2]] {
		t.Errorf("exclude_study=%s: expected the other two, got %v", ids[0], got)
	}
	if got := search(ids[:2], ids[1:2]); !got[ids[0]] || got[ids[1]] || got[ids[2]] {
		t.Errorf("study_id=%v exclude_study=%s: expected only %s, got %v", ids[:2], ids[1], ids[0], got)
	}
}
