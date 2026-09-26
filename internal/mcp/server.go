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

// getStringList reads a JSON array of strings from args[key]. MCP tool
// arguments with `array` schema arrive as []interface{} of strings (a lone
// string is accepted too, since models sometimes send one); this helper
// coerces them and drops non-string entries so the caller can pass
// the result straight to filter helpers.
func getStringList(args map[string]any, key string) []string {
	if args == nil {
		return nil
	}
	if s, ok := args[key].(string); ok && s != "" {
		return []string{s}
	}
	raw, ok := args[key].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
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
			mcp.WithDescription("Search the question bank by question text, concept, name, or answer type. Returns assembled questions (not raw variables). Multi-term queries are split on whitespace and commas; a question matches if any term matches, and is ranked by the number of terms matched then by field weight. Both German and English stemming is applied, and German umlauts are folded (Qualität matches Qualitaet). Common survey terms are translated between German and English, so query in either language (Vertrauen also finds English trust items). Example: query='Wirkung Bildungsprogramm Zufriedenheit' returns the union of the three single-term hits."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search terms, whitespace or comma separated. Single keywords work too. Example: 'trust' or 'Wirkung Bildungsprogramm Zufriedenheit'.")),
			mcp.WithArray("study_id", mcp.Description("Optional list of study IDs to include. Repeatable. If omitted, all studies are searched.")),
			mcp.WithArray("exclude_study", mcp.Description("Optional list of study IDs to exclude (e.g. demographic standards). Repeatable.")),
			mcp.WithToolAnnotation(readOnlyAnnotation),
		),
		searchQuestionsHandler(app),
	)

	s.AddTool(
		mcp.NewTool("search_studies",
			mcp.WithDescription("Search studies by title, keywords, or abstract. Optionally filter by topic classification. Uses substring matching — search one term at a time for best results. For broad topics, run multiple searches with different terms. Examples: query='health' finds health-related studies; query='2020' finds studies from 2020."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Single search term (substring match, case-insensitive). Use one keyword, not a phrase. Example: 'Engagement' not 'bürgerschaftliches Engagement Motivation'.")),
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
		mcpserver.WithStateLess(true),
	)
}

func searchQuestionsHandler(app core.App) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		q := getString(args, "query")
		if q == "" {
			return mcp.NewToolResultError("missing required parameter: query"), nil
		}

		matched, err := routes.SearchQuestions(app, q, getStringList(args, "study_id"), getStringList(args, "exclude_study"))
		if err != nil {
			return mcp.NewToolResultError("failed to fetch studies"), nil
		}
		if len(matched) > 20 {
			matched = matched[:20]
		}
		if matched == nil {
			matched = []routes.Question{}
		}

		b, _ := json.Marshal(matched)
		return mcp.NewToolResultText(string(b)), nil
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

		b, _ := json.Marshal(results)
		return mcp.NewToolResultText(string(b)), nil
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
				b, _ := json.Marshal(q)
				return mcp.NewToolResultText(string(b)), nil
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
			b, _ := json.Marshal(questions)
			return mcp.NewToolResultText(string(b)), nil
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
