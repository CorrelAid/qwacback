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

func TestExamplesRoutes(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_examples")
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
			return RegisterRoutes(testApp, se, nil, "../..")
		})
		return testApp
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "list all examples",
			Method:          http.MethodGet,
			URL:             "/api/examples",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"answer_type"`, `"xlsform"`, `"ddi"`, `"single_choice"`, `"text"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "get single example by type",
			Method:          http.MethodGet,
			URL:             "/api/examples/single_choice",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"answer_type":"single_choice"`, `"xlsform"`, `"ddi"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:           "get nonexistent example type",
			Method:         http.MethodGet,
			URL:            "/api/examples/nonexistent",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestSchemaRoutes(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_schemas")
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
			return RegisterRoutes(testApp, se, nil, "../..")
		})
		return testApp
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "get schematron rules",
			Method:          http.MethodGet,
			URL:             "/api/schemas/schematron",
			ExpectedStatus:  200,
			ExpectedContent: []string{`<schema`, `<pattern`, `other_variables`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "list xsd files",
			Method:          http.MethodGet,
			URL:             "/api/schemas/xsd",
			ExpectedStatus:  200,
			ExpectedContent: []string{`codebook.xsd`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "get codebook xsd",
			Method:          http.MethodGet,
			URL:             "/api/schemas/xsd/codebook.xsd",
			ExpectedStatus:  200,
			ExpectedContent: []string{`<xs:schema`, `codeBook`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "get nested xhtml xsd",
			Method:          http.MethodGet,
			URL:             "/api/schemas/xsd/XHTML/xhtml-text-1.xsd",
			ExpectedStatus:  200,
			ExpectedContent: []string{`<xs:schema`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:           "xsd file not found",
			Method:         http.MethodGet,
			URL:            "/api/schemas/xsd/nonexistent.xsd",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
		{
			Name:           "reject non-xsd file",
			Method:         http.MethodGet,
			URL:            "/api/schemas/xsd/somefile.txt",
			ExpectedStatus: 400,
			TestAppFactory: setupTestApp,
		},
		{
			Name:           "reject directory traversal",
			Method:         http.MethodGet,
			URL:            "/api/schemas/xsd/../../../etc/passwd",
			ExpectedStatus: 400,
			TestAppFactory: setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestDocsRoutes(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_docs")
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
			return RegisterRoutes(testApp, se, nil, "../..")
		})
		return testApp
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "get markup guide",
			Method:          http.MethodGet,
			URL:             "/api/docs/markup-guide",
			ExpectedStatus:  200,
			ExpectedContent: []string{`DDI Markup Guide`, `answer_type`, `<var`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

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
			return RegisterRoutes(testApp, se, nil, "../..")
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
	var varID, groupID string
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
			return RegisterRoutes(testApp, se, nil, "../..")
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
			ExpectedContent: []string{`<varGrp`, `<concept>`, `type=`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:           "variable group xml not found",
			Method:         http.MethodGet,
			URL:            "/api/variable-groups/nonexistent00/xml",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
		// --- Variable XLSForm ---
		{
			Name:            "variable xlsform",
			Method:          http.MethodGet,
			URL:             "/api/variables/" + varID + "/xlsform",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"survey"`, `"choices"`, `"type"`, `"name"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:           "variable xlsform not found",
			Method:         http.MethodGet,
			URL:            "/api/variables/nonexistent00/xlsform",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
		// --- Variable group XLSForm ---
		{
			Name:            "variable group xlsform",
			Method:          http.MethodGet,
			URL:             "/api/variable-groups/" + groupID + "/xlsform",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"survey"`, `"choices"`, `"type"`, `"name"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:           "variable group xlsform not found",
			Method:         http.MethodGet,
			URL:            "/api/variable-groups/nonexistent00/xlsform",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestSearchQuestionsRoute(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "pb_test_search")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDataDir)

	// Seed test data
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
		app.Cleanup()
	}

	setupTestApp := func(t testing.TB) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}
		testApp.OnServe().BindFunc(func(se *core.ServeEvent) error {
			return RegisterRoutes(testApp, se, nil, "../..")
		})
		return testApp
	}

	scenarios := []tests.ApiScenario{
		{
			Name:           "missing q parameter",
			Method:         http.MethodGet,
			URL:            "/api/search/questions",
			ExpectedStatus: 400,
			TestAppFactory: setupTestApp,
		},
		{
			Name:           "empty q parameter",
			Method:         http.MethodGet,
			URL:            "/api/search/questions?q=",
			ExpectedStatus: 400,
			TestAppFactory: setupTestApp,
		},
		{
			Name:            "search returns results",
			Method:          http.MethodGet,
			URL:             "/api/search/questions?q=Vertrauen",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"items"`, `"totalItems"`, `"concept"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "search no results",
			Method:          http.MethodGet,
			URL:             "/api/search/questions?q=zzzznonexistentzzzz",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"items":[]`, `"totalItems":0`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "search with pagination",
			Method:          http.MethodGet,
			URL:             "/api/search/questions?q=Vertrauen&page=1&perPage=2",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"page":1`, `"perPage":2`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
