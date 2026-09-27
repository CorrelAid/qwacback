package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"qwacback/internal/converter"
	"qwacback/internal/importer"

	"github.com/clbanning/mxj/v2"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// useFakeEmitter points the converter at handler (nil: at a closed port) for
// the rest of the test.
func useFakeEmitter(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	url := srv.URL
	if handler == nil {
		srv.Close()
	} else {
		t.Cleanup(srv.Close)
	}
	prev := converter.DDIEmitterURL
	converter.DDIEmitterURL = url
	t.Cleanup(func() { converter.DDIEmitterURL = prev })
}

func TestConvertXLSFormToDDIRoute(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_convert")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDataDir)

	setupTestApp := func(t testing.TB) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}
		testApp.OnServe().BindFunc(func(se *core.ServeEvent) error {
			if err := RegisterRoutes(testApp, se, nil, "../.."); err != nil {
				return err
			}
			return se.Next()
		})
		return testApp
	}
	form := strings.NewReader(`{"survey":[{"type":"rank","name":"a","label":"A"}]}`)

	t.Run("library rejection reaches the client", func(t *testing.T) {
		useFakeEmitter(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"type \"rank\" (question \"a\") is not in the registry"}`))
		})
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/convert/xlsform-to-ddi",
			Body:            form,
			ExpectedStatus:  400,
			ExpectedContent: []string{`is not in the registry`, `question \"a\"`},
			TestAppFactory:  setupTestApp,
		}).Test(t)
	})

	t.Run("sidecar down is 503, not 400", func(t *testing.T) {
		useFakeEmitter(t, nil)
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/convert/xlsform-to-ddi",
			Body:            strings.NewReader(`{"survey":[{"type":"integer","name":"a","label":"A"}]}`),
			ExpectedStatus:  503,
			ExpectedContent: []string{`temporarily unavailable`},
			TestAppFactory:  setupTestApp,
		}).Test(t)
	})
}

// #40: the catalogue keyed by answer type, or by registry slug.
func TestQuestionTypesRoute(t *testing.T) {
	if !ddiEmitterReachable(t) {
		t.Skip("ddi-emitter not reachable")
	}
	setupTestApp := func(t testing.TB) *tests.TestApp {
		testApp, err := tests.NewTestApp(t.(*testing.T).TempDir())
		if err != nil {
			t.Fatal(err)
		}
		testApp.OnServe().BindFunc(func(se *core.ServeEvent) error {
			if err := RegisterRoutes(testApp, se, nil, "../.."); err != nil {
				return err
			}
			return se.Next()
		})
		return testApp
	}
	(&tests.ApiScenario{
		Method:         http.MethodGet,
		URL:            "/api/question-types",
		ExpectedStatus: 200,
		ExpectedContent: []string{
			`"single_choice_other":{`, `"registryType":"select_one_other"`,
			`"withOther":true`, `"en":"Integer"`, `"de":"Ganzzahl"`, `"aliases":["int"]`,
			`"appearance":"minimal"`,
		},
		TestAppFactory: setupTestApp,
	}).Test(t)
	(&tests.ApiScenario{
		Method:          http.MethodGet,
		URL:             "/api/question-types?registry=1",
		ExpectedStatus:  200,
		ExpectedContent: []string{`"select_one_other":{`, `"id":"variant:select_one_other"`, `"deviceid":{`},
		TestAppFactory:  setupTestApp,
	}).Test(t)
}

// #46: a single question's XLSForm drops skip logic naming a question
// outside it, with a warning; the DDI keeps it.
func TestQuestionXLSFormDropsOutsideReferences(t *testing.T) {
	if !ddiEmitterReachable(t) {
		t.Skip("ddi-emitter not reachable")
	}
	x := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5">
  <stdyDscr><citation><titlStmt><titl>Outside references</titl></titlStmt></citation></stdyDscr>
  <dataDscr>
    <var ID="V_alter" name="alter" intrvl="contin" dcml="0"><qstn responseDomainType="numeric" seqNo="1"><qstnLit>Alter?</qstnLit></qstn><concept>Alter</concept><varFormat type="numeric" schema="other"/></var>
    <var ID="V_rente" name="rente" intrvl="discrete"><qstn responseDomainType="text" seqNo="2"><qstnLit>Seit?</qstnLit><backward qstn="V_alter"/></qstn><universe clusion="I">Only if Alter? &gt; 60</universe><concept>Rente</concept><varFormat type="character" schema="other"/><notes type="cdl:relevant" subject="xlsform-xpath">${alter} &gt; 60</notes></var>
  </dataDscr>
</codeBook>`)
	var rente string
	setupTestApp := func(t testing.TB) *tests.TestApp {
		testApp, err := tests.NewTestApp(t.(*testing.T).TempDir())
		if err != nil {
			t.Fatal(err)
		}
		mv, err := mxj.NewMapXml(x)
		if err != nil {
			t.Fatal(err)
		}
		if err := importer.ImportCodebookData(testApp, mv, x); err != nil {
			t.Fatal(err)
		}
		v, err := testApp.FindFirstRecordByFilter("variables", "name = 'rente'")
		if err != nil {
			t.Fatal(err)
		}
		rente = v.Id
		testApp.OnServe().BindFunc(func(se *core.ServeEvent) error {
			if err := RegisterRoutes(testApp, se, nil, "../.."); err != nil {
				return err
			}
			return se.Next()
		})
		return testApp
	}
	// The record ID is deterministic: set it up once to learn it.
	setupTestApp(t).Cleanup()

	(&tests.ApiScenario{
		Method:             http.MethodGet,
		URL:                "/api/questions/" + rente + "/xlsform",
		ExpectedStatus:     200,
		ExpectedContent:    []string{`"name":"rente"`, `"code":"reference-outside"`, `relevant of \"rente\" dropped`},
		NotExpectedContent: []string{`"relevant"`},
		TestAppFactory:     setupTestApp,
	}).Test(t)
	(&tests.ApiScenario{
		Method:          http.MethodGet,
		URL:             "/api/questions/" + rente + "/xml",
		ExpectedStatus:  200,
		ExpectedContent: []string{`cdl:relevant`},
		TestAppFactory:  setupTestApp,
	}).Test(t)
}
