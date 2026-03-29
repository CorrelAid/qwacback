package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/dbx"
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

// seedSearchTestData imports prove_it.xml into a temporary PocketBase and returns
// the data dir and the study ID.
func seedSearchTestData(t *testing.T) (string, string) {
	t.Helper()
	testDataDir, err := os.MkdirTemp("", "pb_test_search")
	if err != nil {
		t.Fatal(err)
	}
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
	studies, _ := app.FindRecordsByFilter("studies", "", "", 1, 0)
	if len(studies) == 0 {
		t.Fatal("no studies after seeding")
	}
	studyID := studies[0].Id
	app.Cleanup()
	return testDataDir, studyID
}

func searchTestApp(testDataDir string) func(t testing.TB) *tests.TestApp {
	return func(t testing.TB) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}
		testApp.OnServe().BindFunc(func(se *core.ServeEvent) error {
			return RegisterRoutes(testApp, se, nil, "../..")
		})
		return testApp
	}
}

func TestSearchStudiesRoute(t *testing.T) {
	testDataDir, _ := seedSearchTestData(t)
	defer os.RemoveAll(testDataDir)
	setupTestApp := searchTestApp(testDataDir)

	scenarios := []tests.ApiScenario{
		{
			Name:           "missing q parameter",
			Method:         http.MethodGet,
			URL:            "/api/search/studies",
			ExpectedStatus: 400,
			TestAppFactory: setupTestApp,
		},
		{
			Name:            "search by title",
			Method:          http.MethodGet,
			URL:             "/api/search/studies?q=Prove",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"items"`, `"totalItems"`, `"title"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "search by keyword",
			Method:          http.MethodGet,
			URL:             "/api/search/studies?q=trust",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"items"`, `"totalItems"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "search no results",
			Method:          http.MethodGet,
			URL:             "/api/search/studies?q=zzzznonexistentzzzz",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"items":[]`, `"totalItems":0`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "filter by topic",
			Method:          http.MethodGet,
			URL:             "/api/search/studies?q=Prove&topic=Template",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"totalItems":1`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "filter by topic no match",
			Method:          http.MethodGet,
			URL:             "/api/search/studies?q=Prove&topic=Nonexistent",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"items":[]`, `"totalItems":0`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestSearchQuestionsRoute(t *testing.T) {
	testDataDir, _ := seedSearchTestData(t)
	defer os.RemoveAll(testDataDir)
	setupTestApp := searchTestApp(testDataDir)

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
			URL:             "/api/search/questions?q=trust",
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
			URL:             "/api/search/questions?q=trust&page=1&perPage=2",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"page":1`, `"perPage":2`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestQuestionsView(t *testing.T) {
	testDataDir, studyID := seedSearchTestData(t)
	defer os.RemoveAll(testDataDir)
	setupTestApp := searchTestApp(testDataDir)

	scenarios := []tests.ApiScenario{
		{
			Name:           "not found",
			Method:         http.MethodGet,
			URL:            "/api/studies/nonexistent00/questions",
			ExpectedStatus: 404,
			TestAppFactory: setupTestApp,
		},
		{
			Name:            "returns questions",
			Method:          http.MethodGet,
			URL:             "/api/studies/" + studyID + "/questions",
			ExpectedStatus:  200,
			ExpectedContent: []string{`"name"`, `"concept"`, `"answer_type"`, `"variable_ids"`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}

	// Verify deduplication: prove_it.xml has 28 variables and 2 grid groups.
	// Grid member vars should be merged into their group question.
	t.Run("fewer questions than variables", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "pb_test_questions_dedup")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)

		app, err := tests.NewTestApp(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

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

		studies, _ := app.FindRecordsByFilter("studies", "", "", 1, 0)
		if len(studies) == 0 {
			t.Fatal("no studies found")
		}
		sid := studies[0].Id

		vars, _ := app.FindRecordsByFilter("variables", "study = {:sid}", "", 0, 0, dbx.Params{"sid": sid})
		groups, _ := app.FindRecordsByFilter("variable_groups", "study = {:sid}", "", 0, 0, dbx.Params{"sid": sid})

		questions, err := assembleQuestions(app, sid)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("assembled %d questions from %d variables + %d groups", len(questions), len(vars), len(groups))

		if len(questions) == 0 {
			t.Fatal("no questions assembled")
		}
		if len(groups) > 0 && len(questions) >= len(vars) {
			t.Errorf("expected fewer questions than variables (groups should merge), got %d questions vs %d variables", len(questions), len(vars))
		}
	})
}

// TestSearchQuestionsOrdering verifies that results matching in higher-priority
// fields rank above those matching in lower-priority fields.
// In prove_it.xml:
//   - council_trust: matches "trust" in question, concept, AND name (score 6+5+4=15)
//   - neighbour_trust: matches "trust" in concept and name only (score 5+4=9)
// So council_trust must appear before neighbour_trust.
func TestSearchQuestionsOrdering(t *testing.T) {
	testDataDir, _ := seedSearchTestData(t)
	defer os.RemoveAll(testDataDir)
	setupTestApp := searchTestApp(testDataDir)

	scenario := tests.ApiScenario{
		Name:           "question+concept+name match ranks above concept+name match",
		Method:         http.MethodGet,
		URL:            "/api/search/questions?q=trust",
		ExpectedStatus: 200,
		TestAppFactory: setupTestApp,
		AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal("failed to read response body:", err)
			}

			var resp struct {
				Items []struct {
					Name string `json:"name"`
				} `json:"items"`
			}
			if err := json.Unmarshal(body, &resp); err != nil {
				t.Fatal("failed to decode response:", err)
			}

			if len(resp.Items) < 2 {
				t.Fatalf("expected at least 2 results, got %d", len(resp.Items))
			}

			councilIdx := -1
			neighbourIdx := -1
			for i, item := range resp.Items {
				if item.Name == "council_trust" {
					councilIdx = i
				}
				if item.Name == "neighbour_trust" {
					neighbourIdx = i
				}
			}

			if councilIdx == -1 {
				t.Fatal("council_trust not found in results")
			}
			if neighbourIdx == -1 {
				t.Fatal("neighbour_trust not found in results")
			}
			if councilIdx >= neighbourIdx {
				t.Errorf("council_trust (matches question+concept+name) should rank before neighbour_trust (concept+name only), got indices %d vs %d", councilIdx, neighbourIdx)
			}
		},
	}
	scenario.Test(t)
}
