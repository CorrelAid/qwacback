package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"qwacback/internal/converter"

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
