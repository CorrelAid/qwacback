package routes

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
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

// acceptAll is a schematron.Client that finds every file valid, so the
// import route can be tested without the validation worker.
type acceptAll struct{}

func (acceptAll) WaitForWorker(time.Duration) error { return nil }
func (acceptAll) Validate([]byte) (*schematron.ValidationResponse, error) {
	return &schematron.ValidationResponse{Valid: true}, nil
}
func (acceptAll) Close() {}

// importBody is a multipart body with xmlData as its file.
func importBody(xmlData []byte) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, _ := w.CreateFormFile("file", "demo.xml")
	part.Write(xmlData)
	w.Close()
	return &buf, w.FormDataContentType()
}

// importApp is an app with the routes, a superuser and a regular user, so
// their tokens are valid in the app the request runs against.
func importApp(t testing.TB) (app *tests.TestApp, superuserToken, userToken string) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	tokenFor := func(collection, email string) string {
		c, err := app.FindCollectionByNameOrId(collection)
		if err != nil {
			t.Fatal(err)
		}
		record := core.NewRecord(c)
		record.SetEmail(email)
		record.SetPassword("1234567890")
		if err := app.Save(record); err != nil {
			t.Fatal(err)
		}
		token, err := record.NewAuthToken()
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	superuserToken = tokenFor(core.CollectionNameSuperusers, "admin@example.com")
	userToken = tokenFor("users", "user@example.com")
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		if err := RegisterRoutes(app, se, acceptAll{}, "../.."); err != nil {
			return err
		}
		return se.Next()
	})
	return app, superuserToken, userToken
}

func TestImportReturnsStudyID(t *testing.T) {
	os.Setenv("QWACBACK_SKIP_SEED", "1")
	defer os.Unsetenv("QWACBACK_SKIP_SEED")

	xmlData, err := os.ReadFile("../../seed_data/demo.xml")
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []struct {
		name    string
		as      string // "superuser", "user" or "" (guest)
		status  int
		content []string
	}{
		{"superuser imports and gets the study ID", "superuser", 200, []string{`"imported":true`, `"study_id":"`}},
		{"regular user is refused", "user", 403, []string{`"status":403`}},
		{"guest is refused", "", 401, []string{`"status":401`}},
	} {
		app, superuserToken, userToken := importApp(t)
		buf, contentType := importBody(xmlData)
		headers := map[string]string{"Content-Type": contentType}
		switch s.as {
		case "superuser":
			headers["Authorization"] = superuserToken
		case "user":
			headers["Authorization"] = userToken
		}
		(&tests.ApiScenario{
			Name:            s.name,
			Method:          http.MethodPost,
			URL:             "/api/import",
			Body:            buf,
			Headers:         headers,
			ExpectedStatus:  s.status,
			ExpectedContent: s.content,
			TestAppFactory:  func(testing.TB) *tests.TestApp { return app },
		}).Test(t)
	}
}

// A title that exists gives 409 unless the client asks to replace the study;
// a seed study is never replaced through the API (#56).
func TestImportExistingTitle(t *testing.T) {
	xmlData, err := os.ReadFile("../../seed_data/demo.xml")
	if err != nil {
		t.Fatal(err)
	}
	mv, err := mxj.NewMapXml(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	studyID := importer.StudyID(mv)
	// The replacement: the same study, different bytes.
	changed := append(bytes.TrimRight(xmlData, "\n"), []byte("\n<!-- replaced -->\n")...)

	for _, s := range []struct {
		name     string
		existing string // source of the study already there
		url      string
		body     []byte
		status   int
		content  []string
		codebook []byte // the study's codebook afterwards
	}{
		{"upload exists: 409", importer.SourceUpload, "/api/import", changed, 409,
			[]string{`"imported":false`, `"study_id":"` + studyID + `"`, `replace=1`}, xmlData},
		{"upload exists, replace=1: replaced", importer.SourceUpload, "/api/import?replace=1", changed, 200,
			[]string{`"imported":true`, `"replaced":true`, `"study_id":"` + studyID + `"`}, changed},
		{"seed exists: 409", importer.SourceSeed, "/api/import", changed, 409,
			[]string{`"imported":false`, `seed_data/`}, xmlData},
		{"seed exists, replace=1: still 409", importer.SourceSeed, "/api/import?replace=1", changed, 409,
			[]string{`"imported":false`, `seed_data/`}, xmlData},
	} {
		app, superuserToken, _ := importApp(t)
		if _, err := importer.ImportStudy(app, mv, xmlData, s.existing, ""); err != nil {
			t.Fatal(err)
		}
		vars, _ := app.CountRecords("variables", dbx.HashExp{"study": studyID})
		// A cached export of the old study must not outlive a replace.
		setXMLCache(app, "export:"+studyID, []byte("stale"))

		buf, contentType := importBody(s.body)
		(&tests.ApiScenario{
			Name:            s.name,
			Method:          http.MethodPost,
			URL:             s.url,
			Body:            buf,
			Headers:         map[string]string{"Content-Type": contentType, "Authorization": superuserToken},
			ExpectedStatus:  s.status,
			ExpectedContent: s.content,
			TestAppFactory:  func(testing.TB) *tests.TestApp { return app },
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				study, err := app.FindRecordById("studies", studyID)
				if err != nil {
					t.Fatalf("study is gone: %v", err)
				}
				if study.GetString("codebook") != string(s.codebook) {
					t.Error("study holds the wrong codebook")
				}
				if study.GetString("source") != s.existing {
					t.Errorf("source %q, want %q", study.GetString("source"), s.existing)
				}
				if n, _ := app.CountRecords("variables", dbx.HashExp{"study": studyID}); n != vars {
					t.Errorf("%d variables, want %d", n, vars)
				}
				_, cached := getXMLCache(app, "export:"+studyID)
				if replaced := s.status == 200; cached == replaced {
					t.Errorf("export cached: %v after replaced: %v", cached, replaced)
				}
			},
		}).Test(t)
	}
}
