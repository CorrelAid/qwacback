package mcp_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/clbanning/mxj/v2"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/pocketbase/pocketbase/tests"

	"qwacback/internal/importer"
	qwacmcp "qwacback/internal/mcp"
	_ "qwacback/migrations"
)

func setupTestMCP(t *testing.T) (*mcpclient.Client, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "pb_test_mcp")
	if err != nil {
		t.Fatal(err)
	}

	app, err := tests.NewTestApp(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	// Seed with prove_it.xml
	xmlData, err := os.ReadFile("../../seed_data/prove_it.xml")
	if err != nil {
		app.Cleanup()
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	mv, err := mxj.NewMapXml(xmlData)
	if err != nil {
		app.Cleanup()
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	if err := importer.ImportCodebookData(app, mv, xmlData); err != nil {
		app.Cleanup()
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	// Create MCP server and test HTTP server
	mcpSrv := qwacmcp.NewMCPServer(app)
	ts := mcpserver.NewTestStreamableHTTPServer(mcpSrv)

	client, err := mcpclient.NewStreamableHttpClient(ts.URL + "/mcp")
	if err != nil {
		ts.Close()
		app.Cleanup()
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = client.Initialize(ctx, gomcp.InitializeRequest{})
	if err != nil {
		ts.Close()
		app.Cleanup()
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	cleanup := func() {
		ts.Close()
		app.Cleanup()
		os.RemoveAll(dir)
	}

	return client, cleanup
}

func TestMCP_ListTools(t *testing.T) {
	client, cleanup := setupTestMCP(t)
	defer cleanup()

	ctx := context.Background()
	tools, err := client.ListTools(ctx, gomcp.ListToolsRequest{})
	if err != nil {
		t.Fatal(err)
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools.Tools {
		toolNames[tool.Name] = true
	}

	expected := []string{"search_questions", "search_studies", "get_question", "list_questions"}
	for _, name := range expected {
		if !toolNames[name] {
			t.Errorf("expected tool %q not found", name)
		}
	}
}

func TestMCP_SearchQuestions(t *testing.T) {
	client, cleanup := setupTestMCP(t)
	defer cleanup()

	ctx := context.Background()
	result, err := client.CallTool(ctx, gomcp.CallToolRequest{
		Params: gomcp.CallToolParams{
			Name:      "search_questions",
			Arguments: map[string]any{"query": "trust"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %+v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestMCP_SearchQuestions_MissingQuery(t *testing.T) {
	client, cleanup := setupTestMCP(t)
	defer cleanup()

	ctx := context.Background()
	result, err := client.CallTool(ctx, gomcp.CallToolRequest{
		Params: gomcp.CallToolParams{
			Name:      "search_questions",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !result.IsError {
		t.Error("expected error for missing query")
	}
}

func TestMCP_SearchStudies(t *testing.T) {
	client, cleanup := setupTestMCP(t)
	defer cleanup()

	ctx := context.Background()
	result, err := client.CallTool(ctx, gomcp.CallToolRequest{
		Params: gomcp.CallToolParams{
			Name:      "search_studies",
			Arguments: map[string]any{"query": "Prove"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %+v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestMCP_ListQuestions(t *testing.T) {
	client, cleanup := setupTestMCP(t)
	defer cleanup()

	ctx := context.Background()
	result, err := client.CallTool(ctx, gomcp.CallToolRequest{
		Params: gomcp.CallToolParams{
			Name:      "list_questions",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %+v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestMCP_GetQuestion_NotFound(t *testing.T) {
	client, cleanup := setupTestMCP(t)
	defer cleanup()

	ctx := context.Background()
	result, err := client.CallTool(ctx, gomcp.CallToolRequest{
		Params: gomcp.CallToolParams{
			Name:      "get_question",
			Arguments: map[string]any{"id": "nonexistent0000"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !result.IsError {
		t.Error("expected error for nonexistent question")
	}
}

// get_question returns the items of a grid, not only its stem and variable
// IDs, so a client can read what each row asks.
func TestMCP_GetQuestion_GridItems(t *testing.T) {
	client, cleanup := setupTestMCP(t)
	defer cleanup()

	ctx := context.Background()
	call := func(name string, args map[string]any) string {
		t.Helper()
		result, err := client.CallTool(ctx, gomcp.CallToolRequest{
			Params: gomcp.CallToolParams{Name: name, Arguments: args},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError || len(result.Content) == 0 {
			t.Fatalf("%s: %+v", name, result.Content)
		}
		text, ok := result.Content[0].(gomcp.TextContent)
		if !ok {
			t.Fatalf("%s: content is %T", name, result.Content[0])
		}
		return text.Text
	}

	var found []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(call("search_questions", map[string]any{"query": "contact"})), &found); err != nil {
		t.Fatal(err)
	}
	id := ""
	for _, q := range found {
		if q.Name == "contact_knowledge" {
			id = q.ID
		}
	}
	if id == "" {
		t.Fatalf("contact_knowledge not found in %+v", found)
	}

	var detail struct {
		AnswerType string `json:"answer_type"`
		Variables  []struct {
			Question   string `json:"question"`
			Categories []struct {
				Label string `json:"label"`
			} `json:"categories"`
		} `json:"variables"`
	}
	if err := json.Unmarshal([]byte(call("get_question", map[string]any{"id": id})), &detail); err != nil {
		t.Fatal(err)
	}
	if detail.AnswerType != "grid" || len(detail.Variables) != 4 {
		t.Fatalf("answer_type %q with %d variables, want grid with 4", detail.AnswerType, len(detail.Variables))
	}
	if q := detail.Variables[0].Question; q != "Local Community Groups" {
		t.Errorf("first item = %q, want %q", q, "Local Community Groups")
	}
	if len(detail.Variables[0].Categories) == 0 {
		t.Error("first item has no categories")
	}
}
