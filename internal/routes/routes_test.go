package routes

import (
	"net/http"
	"os"
	"testing"

	_ "qwacback/migrations"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestStudiesAccess(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_data")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDataDir)

	// helper to get a token for a user
	generateToken := func(email string) (string, error) {
		app, err := tests.NewTestApp(testDataDir)
		if err != nil {
			return "", err
		}
		defer app.Cleanup()

		// Ensure the user exists in the test data
		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return "", err
		}

		record, err := app.FindAuthRecordByEmail("users", email)
		if err != nil {
			record = core.NewRecord(collection)
			record.SetEmail(email)
			record.SetPassword("1234567890")
			record.Set("verified", true)
			if err := app.Save(record); err != nil {
				return "", err
			}
		}

		return record.NewAuthToken()
	}

	userToken, err := generateToken("user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	setupTestApp := func(t testing.TB) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}

		// Register routes
		testApp.OnServe().BindFunc(func(se *core.ServeEvent) error {
			return RegisterRoutes(testApp, se, nil)
		})

		return testApp
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "try as guest (aka. no Authorization header)",
			Method:          http.MethodGet,
			URL:             "/api/collections/studies/records",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"message":"The current user is not allowed to perform this action."`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "try as authenticated regular user",
			Method: http.MethodGet,
			URL:    "/api/collections/studies/records",
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"items":`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
