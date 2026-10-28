package routes

import (
	"encoding/json"
	"net/http"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const shortTitleCodebook = `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5" xml:lang="en">
  <stdyDscr><citation><titlStmt>
    <titl>European Company Survey 2019</titl>
    <subTitl>Online questionnaire for Management respondents</subTitl>
    <altTitl>ECS 2019</altTitl>
  </titlStmt></citation></stdyDscr>
  <dataDscr>
    <var ID="V_size" name="size" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric" seqNo="1"><qstnLit>Approximately how many people work in this establishment?</qstnLit></qstn>
      <varFormat type="numeric" schema="other"/>
    </var>
  </dataDscr>
</codeBook>`

func importShortTitleCodebook(t testing.TB, app core.App) string {
	t.Helper()
	mv, err := mxj.NewMapXml([]byte(shortTitleCodebook))
	if err != nil {
		t.Fatal(err)
	}
	id, err := importer.ImportCodebook(app, mv, []byte(shortTitleCodebook))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// QuestionDetail embeds the study's short title and subtitle (#64).
func TestQuestionDetailShortTitle(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	studyID := importShortTitleCodebook(t, app)
	questions, err := AssembleQuestions(app, studyID)
	if err != nil || len(questions) != 1 {
		t.Fatalf("questions = %v, %v", questions, err)
	}
	d, err := QuestionDetail(app, questions[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(d["study"])
	var study map[string]any
	if err := json.Unmarshal(b, &study); err != nil {
		t.Fatal(err)
	}
	for field, want := range map[string]string{
		"title":       "European Company Survey 2019",
		"short_title": "ECS 2019",
		"subtitle":    "Online questionnaire for Management respondents",
	} {
		if got := study[field]; got != want {
			t.Errorf("study.%s = %v, want %q", field, got, want)
		}
	}
}

// The study search finds a study by its short title and returns it.
func TestSearchStudiesShortTitle(t *testing.T) {
	setupTestApp := func(t testing.TB) *tests.TestApp {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		importShortTitleCodebook(t, app)
		app.OnServe().BindFunc(func(se *core.ServeEvent) error {
			if err := RegisterRoutes(app, se, nil, "../.."); err != nil {
				return err
			}
			return se.Next()
		})
		return app
	}
	scenario := tests.ApiScenario{
		Name:            "search by short title",
		Method:          http.MethodGet,
		URL:             "/api/search/studies?q=ECS",
		ExpectedStatus:  200,
		ExpectedContent: []string{`"totalItems":1`, `"short_title":"ECS 2019"`},
		TestAppFactory:  setupTestApp,
	}
	scenario.Test(t)
}
