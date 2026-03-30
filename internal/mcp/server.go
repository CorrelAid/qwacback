package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"qwacback/internal/routes"
)

var readOnlyAnnotation = mcp.ToolAnnotation{
	ReadOnlyHint:    mcp.ToBoolPtr(true),
	DestructiveHint: mcp.ToBoolPtr(false),
	OpenWorldHint:   mcp.ToBoolPtr(false),
}

// getArgs extracts the arguments map from a CallToolRequest.
func getArgs(req mcp.CallToolRequest) map[string]any {
	if m, ok := req.Params.Arguments.(map[string]any); ok {
		return m
	}
	return nil
}

func getString(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	s, _ := args[key].(string)
	return s
}

// NewMCPServer creates the MCP server with tools registered.
func NewMCPServer(app core.App) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer(
		"qwacback",
		"0.1.0",
		mcpserver.WithToolCapabilities(true),
	)

	s.AddTool(
		mcp.NewTool("search_questions",
			mcp.WithDescription("Search the question bank by question text, concept, name, or answer type. Returns assembled questions (not raw variables)."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search term")),
			mcp.WithToolAnnotation(readOnlyAnnotation),
		),
		searchQuestionsHandler(app),
	)

	s.AddTool(
		mcp.NewTool("search_studies",
			mcp.WithDescription("Search studies by title, keywords, or abstract. Optionally filter by topic classification."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search term")),
			mcp.WithString("topic", mcp.Description("Optional topic classification filter")),
			mcp.WithToolAnnotation(readOnlyAnnotation),
		),
		searchStudiesHandler(app),
	)

	s.AddTool(
		mcp.NewTool("get_question",
			mcp.WithDescription("Get a single question by ID, including its variable IDs, concept, question text, and answer type."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Question ID (15-char alphanumeric)")),
			mcp.WithToolAnnotation(readOnlyAnnotation),
		),
		getQuestionHandler(app),
	)

	s.AddTool(
		mcp.NewTool("list_questions",
			mcp.WithDescription("List all questions for a study, or all questions across all studies if no study_id is given."),
			mcp.WithString("study_id", mcp.Description("Optional study ID to filter by")),
			mcp.WithToolAnnotation(readOnlyAnnotation),
		),
		listQuestionsHandler(app),
	)

	return s
}

// NewHTTPServer creates a Streamable HTTP server wrapping the MCP server.
func NewHTTPServer(app core.App) *mcpserver.StreamableHTTPServer {
	return mcpserver.NewStreamableHTTPServer(NewMCPServer(app),
		mcpserver.WithEndpointPath("/mcp"),
	)
}

func searchQuestionsHandler(app core.App) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		q := getString(args, "query")
		if q == "" {
			return mcp.NewToolResultError("missing required parameter: query"), nil
		}

		studies, err := app.FindRecordsByFilter("studies", "", "", 0, 0)
		if err != nil {
			return mcp.NewToolResultError("failed to fetch studies"), nil
		}

		var all []routes.Question
		for _, s := range studies {
			qs, err := routes.AssembleQuestions(app, s.Id)
			if err != nil {
				continue
			}
			all = append(all, qs...)
		}

		matched := routes.FilterAndRankQuestions(all, q)
		if len(matched) > 20 {
			matched = matched[:20]
		}

		return mcp.NewToolResultJSON(matched)
	}
}

func searchStudiesHandler(app core.App) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		q := getString(args, "query")
		if q == "" {
			return mcp.NewToolResultError("missing required parameter: query"), nil
		}

		filter := "title ~ {:q} || abstract ~ {:q} || keywords ~ {:q}"
		params := dbx.Params{"q": q}

		topic := getString(args, "topic")
		if topic != "" {
			filter = "(" + filter + ") && topic_classifications ~ {:topic}"
			params["topic"] = topic
		}

		records, err := app.FindRecordsByFilter("studies", filter, "", 20, 0, params)
		if err != nil {
			return mcp.NewToolResultError("search failed"), nil
		}

		type studyResult struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Abstract string `json:"abstract"`
			Author   string `json:"author"`
			Nation   string `json:"nation"`
		}

		results := make([]studyResult, 0, len(records))
		for _, r := range records {
			results = append(results, studyResult{
				ID:       r.Id,
				Title:    r.GetString("title"),
				Abstract: r.GetString("abstract"),
				Author:   r.GetString("author"),
				Nation:   r.GetString("nation"),
			})
		}

		return mcp.NewToolResultJSON(results)
	}
}

func getQuestionHandler(app core.App) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		id := getString(args, "id")
		if id == "" {
			return mcp.NewToolResultError("missing required parameter: id"), nil
		}

		var studyID string
		if grp, err := app.FindRecordById("variable_groups", id); err == nil {
			studyID = grp.GetString("study")
		} else if v, err := app.FindRecordById("variables", id); err == nil {
			studyID = v.GetString("study")
		} else {
			return mcp.NewToolResultError("question not found"), nil
		}

		questions, err := routes.AssembleQuestions(app, studyID)
		if err != nil {
			return mcp.NewToolResultError("failed to assemble questions"), nil
		}

		for _, q := range questions {
			if q.ID == id {
				return mcp.NewToolResultJSON(q)
			}
		}

		return mcp.NewToolResultError("question not found"), nil
	}
}

func listQuestionsHandler(app core.App) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		studyID := getString(args, "study_id")

		if studyID != "" {
			questions, err := routes.AssembleQuestions(app, studyID)
			if err != nil {
				return mcp.NewToolResultError("failed to assemble questions"), nil
			}
			return mcp.NewToolResultJSON(questions)
		}

		studies, err := app.FindRecordsByFilter("studies", "", "", 0, 0)
		if err != nil {
			return mcp.NewToolResultError("failed to fetch studies"), nil
		}

		var all []routes.Question
		for _, s := range studies {
			qs, err := routes.AssembleQuestions(app, s.Id)
			if err != nil {
				continue
			}
			all = append(all, qs...)
		}

		result, _ := json.Marshal(all)
		return mcp.NewToolResultText(string(result)), nil
	}
}
