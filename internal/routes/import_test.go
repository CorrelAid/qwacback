package routes

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"qwacback/internal/schematron"
	_ "qwacback/migrations"

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

func TestImportReturnsStudyID(t *testing.T) {
	os.Setenv("QWACBACK_SKIP_SEED", "1")
	defer os.Unsetenv("QWACBACK_SKIP_SEED")

	xmlData, err := os.ReadFile("../../seed_data/demo.xml")
	if err != nil {
		t.Fatal(err)
	}
	body := func() (*bytes.Buffer, string) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, _ := w.CreateFormFile("file", "demo.xml")
		part.Write(xmlData)
		w.Close()
		return &buf, w.FormDataContentType()
	}

	// Each scenario gets its own app with a superuser and a regular user, so
	// their tokens are valid in the app the request runs against.
	newApp := func(t testing.TB) (app *tests.TestApp, superuserToken, userToken string) {
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
		app, superuserToken, userToken := newApp(t)
		buf, contentType := body()
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
