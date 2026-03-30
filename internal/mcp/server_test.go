package mcp_test

import (
	"context"
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
