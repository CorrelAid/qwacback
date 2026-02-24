package routes

import (
	"net/http"
	"os"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
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
			ExpectedStatus:  200,
			ExpectedContent: []string{`"items":`},
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

func TestXMLFragmentRoutes(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_xml_routes")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDataDir)

	// Seed test data by importing prove_it.xml
	var studyID, varID, groupID string
	{
		app, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}

		xmlData, err := os.ReadFile("../../seed_data/prove_it.xml")
		if err != nil {
			t.Fatal(err)
		}
		mv, err := mxj.NewMapXml(xmlData)
		if err != nil {
			t.Fatal(err)
		}
		if err := importer.ImportCodebookData(app, mv, xmlData); err != nil {
			t.Fatal(err)
		}

		// Get IDs for test URLs
		studies, _ := app.FindRecordsByFilter("studies", "", "", 1, 0)
		if len(studies) == 0 {
			t.Fatal("No study found after import")
		}
		studyID = studies[0].Id

		vars, _ := app.FindRecordsByFilter("variables", "", "", 1, 0)
		if len(vars) == 0 {
			t.Fatal("No variable found after import")
		}
		varID = vars[0].Id

		groups, _ := app.FindRecordsByFilter("variable_groups", "", "", 1, 0)
		if len(groups) == 0 {
			t.Fatal("No variable group found after import")
		}
		groupID = groups[0].Id

		app.Cleanup()
	}

	// Create a user and get a token
	var userToken string
	{
		app, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}

		collection, _ := app.FindCollectionByNameOrId("users")
		record, err := app.FindAuthRecordByEmail("users", "test@example.com")
		if err != nil {
			record = core.NewRecord(collection)
			record.SetEmail("test@example.com")
			record.SetPassword("1234567890")
			record.Set("verified", true)
			if err := app.Save(record); err != nil {
				t.Fatal(err)
			}
		}
		userToken, err = record.NewAuthToken()
		if err != nil {
			t.Fatal(err)
		}
		app.Cleanup()
	}

	setupTestApp := func(t testing.TB) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}
		testApp.OnServe().BindFunc(func(se *core.ServeEvent) error {
			return RegisterRoutes(testApp, se, nil)
		})
		return testApp
	}

	scenarios := []tests.ApiScenario{
		// --- Variable XML ---
		{
			Name:            "variable xml as guest",
			Method:          http.MethodGet,
			URL:             "/api/variables/" + varID + "/xml",
			ExpectedStatus:  200,
			ExpectedContent: []string{`<var`, `<qstn`, `responseDomainType=`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "variable xml as authenticated user",
			Method: http.MethodGet,
			URL:    "/api/variables/" + varID + "/xml",
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`<var`, `<labl>`, `<qstn`, `responseDomainType=`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:           "variable xml not found",
			Method:         http.MethodGet,
			URL:            "/api/variables/nonexistent00/xml",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
		// --- Variable group XML ---
		{
			Name:            "variable group xml as guest",
			Method:          http.MethodGet,
			URL:             "/api/variable-groups/" + groupID + "/xml",
			ExpectedStatus:  200,
			ExpectedContent: []string{`<varGrp`, `type=`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "variable group xml as authenticated user",
			Method: http.MethodGet,
			URL:    "/api/variable-groups/" + groupID + "/xml",
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`<varGrp`, `<labl>`, `type=`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:           "variable group xml not found",
			Method:         http.MethodGet,
			URL:            "/api/variable-groups/nonexistent00/xml",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
		// --- Study XML ---
		{
			Name:            "study xml as guest",
			Method:          http.MethodGet,
			URL:             "/api/studies/" + studyID + "/xml",
			ExpectedStatus:  200,
			ExpectedContent: []string{`<stdyDscr>`, `<titl>Prove It!`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "study xml as authenticated user",
			Method: http.MethodGet,
			URL:    "/api/studies/" + studyID + "/xml",
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`<stdyDscr>`, `<titl>Prove It!`, `<AuthEnty`, `New Economics Foundation`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:           "study xml not found",
			Method:         http.MethodGet,
			URL:            "/api/studies/nonexistent00/xml",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
