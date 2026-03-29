package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"qwacback/internal/converter"
	"qwacback/internal/examples"
	"qwacback/internal/exporter"
	"qwacback/internal/importer"
	"qwacback/internal/schematron"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const xmlCachePrefix = "xmlc:"
const xmlCacheTTL = 5 * time.Minute
const maxUploadSize = 50 * 1024 * 1024 // 50 MB
const maxConvertSize = 1 * 1024 * 1024  // 1 MB — conversion endpoints handle small fragments

// pocketbaseIDRegex matches PocketBase's 15-character alphanumeric record IDs.
var pocketbaseIDRegex = regexp.MustCompile(`^[a-z0-9]{15}$`)

type cachedResponse struct {
	Data    []byte
	Expires time.Time
}

func getXMLCache(app core.App, key string) ([]byte, bool) {
	v := app.Store().Get(xmlCachePrefix + key)
	if v == nil {
		return nil, false
	}
	c, ok := v.(cachedResponse)
	if !ok || time.Now().After(c.Expires) {
		app.Store().Remove(xmlCachePrefix + key)
		return nil, false
	}
	return c.Data, true
}

func setXMLCache(app core.App, key string, data []byte) {
	app.Store().Set(xmlCachePrefix+key, cachedResponse{
		Data:    data,
		Expires: time.Now().Add(xmlCacheTTL),
	})
}

func clearXMLCache(app core.App) {
	for k := range app.Store().GetAll() {
		if strings.HasPrefix(k, xmlCachePrefix) {
			app.Store().Remove(k)
		}
	}
}

func parsePagination(e *core.RequestEvent) (page, perPage int) {
	page, _ = strconv.Atoi(e.Request.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ = strconv.Atoi(e.Request.URL.Query().Get("perPage"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return
}

// scoreRecord computes a relevance score for a record by checking which fields
// contain the query string. Fields earlier in the list receive higher weight.
func scoreRecord(r *core.Record, q string, fields []string) int {
	q = strings.ToLower(q)
	score := 0
	for i, field := range fields {
		if strings.Contains(strings.ToLower(r.GetString(field)), q) {
			score += len(fields) - i
		}
	}
	return score
}

// rankAndPaginate sorts records by relevance score (descending) and returns
// the requested page along with total count.
func rankAndPaginate(records []*core.Record, q string, fields []string, page, perPage int) ([]*core.Record, int) {
	sort.SliceStable(records, func(i, j int) bool {
		return scoreRecord(records[i], q, fields) > scoreRecord(records[j], q, fields)
	})
	total := len(records)
	offset := (page - 1) * perPage
	if offset >= total {
		return nil, total
	}
	end := offset + perPage
	if end > total {
		end = total
	}
	return records[offset:end], total
}

// Question represents a researcher-level survey question assembled from
// variables and variable groups. A question is "one thing asked of the
// respondent" — it may map to a single variable, a variable group, or
// a hierarchy of groups.
type Question struct {
	ID           string   `json:"id"`
	StudyID      string   `json:"study_id"`
	Name         string   `json:"name"`
	Concept      string   `json:"concept"`
	QuestionText string   `json:"question_text"`
	AnswerType   string   `json:"answer_type"`
	VariableIDs  []string `json:"variable_ids"`
	GroupID      string   `json:"group_id,omitempty"`
	Order        float64  `json:"order"`
}

// assembleQuestions builds a question-level view from variables and groups.
// The importer flattens the DDI group hierarchy: child groups (e.g. _choices
// under a type="other" parent) are not stored. All member vars are assigned
// directly to the top-level group. So the rules here are simple:
//   - Each group → 1 question (its member vars are absorbed).
//   - Each standalone var (no group) → 1 question.
func assembleQuestions(app core.App, studyID string) ([]Question, error) {
	groups, err := app.FindRecordsByFilter("variable_groups", "study = {:sid}", "order", 0, 0, dbx.Params{"sid": studyID})
	if err != nil {
		return nil, err
	}

	allVars, err := app.FindRecordsByFilter("variables", "study = {:sid}", "order", 0, 0, dbx.Params{"sid": studyID})
	if err != nil {
		return nil, err
	}

	// Index vars by group
	varsByGroup := make(map[string][]*core.Record)
	groupedVarIDs := make(map[string]bool)
	for _, v := range allVars {
		gid := v.GetString("group")
		if gid != "" {
			varsByGroup[gid] = append(varsByGroup[gid], v)
			groupedVarIDs[v.Id] = true
		}
	}

	var questions []Question

	// 1. Groups → questions
	for _, g := range groups {
		gType := g.GetString("type")
		q := Question{
			ID:           g.Id,
			StudyID:      studyID,
			Name:         g.GetString("name"),
			Concept:      g.GetString("concept"),
			QuestionText: g.GetString("description"),
			GroupID:      g.Id,
			Order:        g.GetFloat("order"),
		}

		for _, v := range varsByGroup[g.Id] {
			q.VariableIDs = append(q.VariableIDs, v.Id)
		}

		// Determine answer_type from group type
		switch gType {
		case "other":
			// Check if any member var has responseDomainType="multiple" (→ multiple_choice_other)
			hasMultiple := false
			for _, v := range varsByGroup[g.Id] {
				if v.GetString("answer_type") == "multiple_choice" {
					hasMultiple = true
					break
				}
			}
			if hasMultiple {
				q.AnswerType = "multiple_choice_other"
			} else {
				q.AnswerType = "single_choice_other"
			}
		case "grid":
			q.AnswerType = "grid"
		case "multipleResp":
			q.AnswerType = "multiple_choice"
		}

		// Use first member var's question text if group description is empty
		if q.QuestionText == "" && len(varsByGroup[g.Id]) > 0 {
			first := varsByGroup[g.Id][0]
			q.QuestionText = first.GetString("prequestion_text")
			if q.QuestionText == "" {
				q.QuestionText = first.GetString("question")
			}
		}

		questions = append(questions, q)
	}

	// 2. Standalone vars (no group) → questions
	for _, v := range allVars {
		if groupedVarIDs[v.Id] {
			continue
		}
		questions = append(questions, Question{
			ID:           v.Id,
			StudyID:      studyID,
			Name:         v.GetString("name"),
			Concept:      v.GetString("concept"),
			QuestionText: v.GetString("question"),
			AnswerType:   v.GetString("answer_type"),
			VariableIDs:  []string{v.Id},
			Order:        v.GetFloat("order"),
		})
	}

	// Sort by order
	sort.SliceStable(questions, func(i, j int) bool {
		return questions[i].Order < questions[j].Order
	})

	return questions, nil
}

// RegisterRoutes sets up the PocketBase API routes.
// rootDir is the project root directory (used to locate schema files).
func RegisterRoutes(app core.App, se *core.ServeEvent, schClient schematron.Client, rootDir ...string) error {
	root := "."
	if len(rootDir) > 0 && rootDir[0] != "" {
		root = rootDir[0]
	}
	// Validation API - Protected by Auth
	se.Router.POST("/api/validate", func(e *core.RequestEvent) error {
		// Enforce upload size limit before reading into memory
		e.Request.Body = http.MaxBytesReader(e.Response, e.Request.Body, maxUploadSize)

		src, _, err := e.Request.FormFile("file")
		if err != nil {
			return apis.NewBadRequestError("Missing or oversized file", nil)
		}
		defer src.Close()

		xmlBytes, err := io.ReadAll(io.LimitReader(src, maxUploadSize))
		if err != nil {
			return apis.NewInternalServerError("Failed to read file", nil)
		}

		// Validation service is required — reject if unavailable
		if schClient == nil {
			return apis.NewInternalServerError("Validation service is unavailable", nil)
		}

		// XML validation via NATS worker (XSD + Schematron)
		resp, err := schClient.Validate(xmlBytes)
		if err != nil {
			log.Printf("ERROR: XML validation request failed")
			return apis.NewInternalServerError("Validation service error", nil)
		}
		if !resp.Valid {
			return e.JSON(400, map[string]interface{}{
				"valid":  false,
				"errors": resp.Errors,
			})
		}

		// Use mxj to parse the XML
		mv, err := mxj.NewMapXml(xmlBytes)
		if err != nil {
			log.Printf("ERROR: mxj failed to parse validated XML")
			return e.JSON(200, map[string]interface{}{
				"valid":   true,
				"message": "XML is valid against schema, but could not be parsed for import",
			})
		}

		// Insert data into collections
		if err := importer.ImportCodebookData(app, mv, xmlBytes); err != nil {
			log.Printf("ERROR: failed to import XML data")
			return e.JSON(200, map[string]interface{}{
				"valid":   true,
				"message": "XML is valid, but failed to import into the database",
			})
		}

		// Clear XML cache after import (data changed)
		clearXMLCache(app)

		return e.JSON(200, map[string]interface{}{
			"valid":   true,
			"message": "XML is valid and imported successfully",
		})
	})

	// Export Study - Public (cached)
	se.Router.GET("/api/studies/{id}/export", func(e *core.RequestEvent) error {
		studyId := e.Request.PathValue("id")
		if !pocketbaseIDRegex.MatchString(studyId) {
			return apis.NewBadRequestError("Invalid ID format", nil)
		}

		setSecureXMLHeaders(e, fmt.Sprintf("attachment; filename=\"study-%s.xml\"", studyId))

		if cached, ok := getXMLCache(app, "export:"+studyId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			_, err := e.Response.Write(cached)
			return err
		}

		study, err := app.FindRecordById("studies", studyId)
		if err != nil {
			return apis.NewNotFoundError("Study not found", nil)
		}

		// Generate XML
		xmlBytes, err := exporter.ExportStudyToXML(app, study)
		if err != nil {
			log.Printf("ERROR: failed to export study %s", studyId)
			return apis.NewInternalServerError("Failed to generate XML", nil)
		}

		// Add XML declaration manually as mxj doesn't add it by default
		xmlBytes = append([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"), xmlBytes...)

		// Validate via NATS worker (XSD + Schematron)
		if schClient != nil {
			resp, err := schClient.Validate(xmlBytes)
			if err != nil {
				log.Printf("WARNING: XML validation unavailable on export for study %s", studyId)
			} else if !resp.Valid {
				log.Printf("ERROR: exported XML failed validation for study %s", studyId)
				return apis.NewInternalServerError("Generated XML failed validation", nil)
			}
		}

		setXMLCache(app, "export:"+studyId, xmlBytes)
		e.Response.Header().Set("X-Cache", "MISS")
		_, err = e.Response.Write(xmlBytes)
		return err
	})

	// Variable XML fragment - Public (cached)
	se.Router.GET("/api/variables/{id}/xml", func(e *core.RequestEvent) error {
		varId := e.Request.PathValue("id")
		if !pocketbaseIDRegex.MatchString(varId) {
			return apis.NewBadRequestError("Invalid ID format", nil)
		}

		setSecureXMLHeaders(e, "")

		if cached, ok := getXMLCache(app, "var:"+varId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			_, err := e.Response.Write(cached)
			return err
		}

		record, err := app.FindRecordById("variables", varId)
		if err != nil {
			return apis.NewNotFoundError("Variable not found", nil)
		}

		xmlBytes, err := exporter.ExportVariableWithGroupToXML(app, record)
		if err != nil {
			log.Printf("ERROR: failed to export variable %s", varId)
			return apis.NewInternalServerError("Failed to generate XML", nil)
		}

		setXMLCache(app, "var:"+varId, xmlBytes)
		e.Response.Header().Set("X-Cache", "MISS")
		_, err = e.Response.Write(xmlBytes)
		return err
	})

	// Variable group XML fragment - Public (cached)
	se.Router.GET("/api/variable-groups/{id}/xml", func(e *core.RequestEvent) error {
		grpId := e.Request.PathValue("id")
		if !pocketbaseIDRegex.MatchString(grpId) {
			return apis.NewBadRequestError("Invalid ID format", nil)
		}

		setSecureXMLHeaders(e, "")

		if cached, ok := getXMLCache(app, "grp:"+grpId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			_, err := e.Response.Write(cached)
			return err
		}

		record, err := app.FindRecordById("variable_groups", grpId)
		if err != nil {
			return apis.NewNotFoundError("Variable group not found", nil)
		}

		xmlBytes, err := exporter.ExportVarGrpToXML(app, record)
		if err != nil {
			log.Printf("ERROR: failed to export variable group %s", grpId)
			return apis.NewInternalServerError("Failed to generate XML", nil)
		}

		setXMLCache(app, "grp:"+grpId, xmlBytes)
		e.Response.Header().Set("X-Cache", "MISS")
		_, err = e.Response.Write(xmlBytes)
		return err
	})

	// Variable group DDI codebook fragment - Public (cached)
	se.Router.GET("/api/variable-groups/{id}/codebook", func(e *core.RequestEvent) error {
		grpId := e.Request.PathValue("id")
		if !pocketbaseIDRegex.MatchString(grpId) {
			return apis.NewBadRequestError("Invalid ID format", nil)
		}

		setSecureXMLHeaders(e, "")

		if cached, ok := getXMLCache(app, "grpcb:"+grpId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			_, err := e.Response.Write(cached)
			return err
		}

		record, err := app.FindRecordById("variable_groups", grpId)
		if err != nil {
			return apis.NewNotFoundError("Variable group not found", nil)
		}

		xmlBytes, err := exporter.ExportVarGrpCodebookToXML(app, record)
		if err != nil {
			log.Printf("ERROR: failed to export variable group codebook %s", grpId)
			return apis.NewInternalServerError("Failed to generate XML", nil)
		}

		setXMLCache(app, "grpcb:"+grpId, xmlBytes)
		e.Response.Header().Set("X-Cache", "MISS")
		_, err = e.Response.Write(xmlBytes)
		return err
	})

	// Study XLSForm - Public (cached)
	se.Router.GET("/api/studies/{id}/xlsform", func(e *core.RequestEvent) error {
		studyId := e.Request.PathValue("id")
		if !pocketbaseIDRegex.MatchString(studyId) {
			return apis.NewBadRequestError("Invalid ID format", nil)
		}

		if cached, ok := getXMLCache(app, "xlsform:study:"+studyId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			e.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
			e.Response.Header().Set("X-Content-Type-Options", "nosniff")
			_, err := e.Response.Write(cached)
			return err
		}

		study, err := app.FindRecordById("studies", studyId)
		if err != nil {
			return apis.NewNotFoundError("Study not found", nil)
		}

		xmlBytes, err := exporter.ExportStudyToXML(app, study)
		if err != nil {
			log.Printf("ERROR: failed to export study %s for XLSForm conversion", studyId)
			return apis.NewInternalServerError("Failed to generate XML", nil)
		}

		xlsformJSON, err := converter.DDIToXLSForm(xmlBytes)
		if err != nil {
			log.Printf("ERROR: failed to convert study %s to XLSForm", studyId)
			return apis.NewInternalServerError("Failed to convert to XLSForm", nil)
		}

		setXMLCache(app, "xlsform:study:"+studyId, xlsformJSON)
		e.Response.Header().Set("X-Cache", "MISS")
		e.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		e.Response.Header().Set("X-Content-Type-Options", "nosniff")
		_, err = e.Response.Write(xlsformJSON)
		return err
	})

	// Variable XLSForm - Public (cached)
	se.Router.GET("/api/variables/{id}/xlsform", func(e *core.RequestEvent) error {
		varId := e.Request.PathValue("id")
		if !pocketbaseIDRegex.MatchString(varId) {
			return apis.NewBadRequestError("Invalid ID format", nil)
		}

		if cached, ok := getXMLCache(app, "xlsform:var:"+varId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			e.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
			e.Response.Header().Set("X-Content-Type-Options", "nosniff")
			_, err := e.Response.Write(cached)
			return err
		}

		record, err := app.FindRecordById("variables", varId)
		if err != nil {
			return apis.NewNotFoundError("Variable not found", nil)
		}

		xmlBytes, err := exporter.ExportVariableWithGroupToXML(app, record)
		if err != nil {
			log.Printf("ERROR: failed to export variable %s for XLSForm conversion", varId)
			return apis.NewInternalServerError("Failed to generate XML", nil)
		}

		xlsformJSON, err := converter.DDIToXLSForm(xmlBytes)
		if err != nil {
			log.Printf("ERROR: failed to convert variable %s to XLSForm", varId)
			return apis.NewInternalServerError("Failed to convert to XLSForm", nil)
		}

		setXMLCache(app, "xlsform:var:"+varId, xlsformJSON)
		e.Response.Header().Set("X-Cache", "MISS")
		e.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		e.Response.Header().Set("X-Content-Type-Options", "nosniff")
		_, err = e.Response.Write(xlsformJSON)
		return err
	})

	// Variable group XLSForm - Public (cached)
	se.Router.GET("/api/variable-groups/{id}/xlsform", func(e *core.RequestEvent) error {
		grpId := e.Request.PathValue("id")
		if !pocketbaseIDRegex.MatchString(grpId) {
			return apis.NewBadRequestError("Invalid ID format", nil)
		}

		if cached, ok := getXMLCache(app, "xlsform:grp:"+grpId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			e.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
			e.Response.Header().Set("X-Content-Type-Options", "nosniff")
			_, err := e.Response.Write(cached)
			return err
		}

		record, err := app.FindRecordById("variable_groups", grpId)
		if err != nil {
			return apis.NewNotFoundError("Variable group not found", nil)
		}

		xmlBytes, err := exporter.ExportVarGrpCodebookToXML(app, record)
		if err != nil {
			log.Printf("ERROR: failed to export variable group %s for XLSForm conversion", grpId)
			return apis.NewInternalServerError("Failed to generate XML", nil)
		}

		xlsformJSON, err := converter.DDIToXLSForm(xmlBytes)
		if err != nil {
			log.Printf("ERROR: failed to convert variable group %s to XLSForm", grpId)
			return apis.NewInternalServerError("Failed to convert to XLSForm", nil)
		}

		setXMLCache(app, "xlsform:grp:"+grpId, xlsformJSON)
		e.Response.Header().Set("X-Cache", "MISS")
		e.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		e.Response.Header().Set("X-Content-Type-Options", "nosniff")
		_, err = e.Response.Write(xlsformJSON)
		return err
	})

	// Examples - Public
	se.Router.GET("/api/examples", func(e *core.RequestEvent) error {
		return e.JSON(200, examples.GetAll())
	})

	se.Router.GET("/api/examples/{answer_type}", func(e *core.RequestEvent) error {
		answerType := e.Request.PathValue("answer_type")
		ex := examples.GetByType(answerType)
		if ex == nil {
			return apis.NewNotFoundError("Answer type not found", nil)
		}
		return e.JSON(200, ex)
	})

	// Documentation - Public
	se.Router.GET("/api/docs/markup-guide", func(e *core.RequestEvent) error {
		data, err := os.ReadFile(filepath.Join(root, "DDI_MARKUP_GUIDE.md"))
		if err != nil {
			return apis.NewInternalServerError("Failed to read markup guide", nil)
		}
		e.Response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		e.Response.Header().Set("X-Content-Type-Options", "nosniff")
		_, err = e.Response.Write(data)
		return err
	})

	// Schema files - Public
	se.Router.GET("/api/schemas/schematron", func(e *core.RequestEvent) error {
		data, err := os.ReadFile(filepath.Join(root, "schematron", "ddi_custom_rules.sch"))
		if err != nil {
			return apis.NewInternalServerError("Failed to read schematron file", nil)
		}
		e.Response.Header().Set("Content-Type", "application/xml; charset=utf-8")
		e.Response.Header().Set("X-Content-Type-Options", "nosniff")
		_, err = e.Response.Write(data)
		return err
	})

	xsdDir := filepath.Join(root, "xml")
	se.Router.GET("/api/schemas/xsd", func(e *core.RequestEvent) error {
		var files []string
		err := filepath.Walk(xsdDir, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(p, ".xsd") {
				rel, _ := filepath.Rel(xsdDir, p)
				files = append(files, rel)
			}
			return nil
		})
		if err != nil {
			return apis.NewInternalServerError("Failed to list XSD files", nil)
		}
		return e.JSON(200, files)
	})

	se.Router.GET("/api/schemas/xsd/{path...}", func(e *core.RequestEvent) error {
		reqPath := e.Request.PathValue("path")
		// Prevent directory traversal
		cleaned := filepath.Clean(reqPath)
		if strings.Contains(cleaned, "..") {
			return apis.NewBadRequestError("Invalid path", nil)
		}
		if !strings.HasSuffix(cleaned, ".xsd") {
			return apis.NewBadRequestError("Only .xsd files are served", nil)
		}

		fullPath := filepath.Join(xsdDir, cleaned)
		// Belt-and-suspenders: verify resolved path is still under xsdDir
		rel, err := filepath.Rel(xsdDir, fullPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return apis.NewBadRequestError("Invalid path", nil)
		}
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return apis.NewNotFoundError("XSD file not found", nil)
		}
		e.Response.Header().Set("Content-Type", "application/xml; charset=utf-8")
		e.Response.Header().Set("X-Content-Type-Options", "nosniff")
		_, err = e.Response.Write(data)
		return err
	})

	// Search studies - Public
	// Relevance: title > keywords > abstract
	// Optional filter: topic (matches topic_classifications)
	se.Router.GET("/api/search/studies", func(e *core.RequestEvent) error {
		q := strings.TrimSpace(e.Request.URL.Query().Get("q"))
		if q == "" {
			return apis.NewBadRequestError("Missing search query parameter 'q'", nil)
		}
		if len(q) > 200 {
			return apis.NewBadRequestError("Query too long (max 200 characters)", nil)
		}

		page, perPage := parsePagination(e)

		filter := "title ~ {:q} || abstract ~ {:q} || keywords ~ {:q}"
		params := dbx.Params{"q": q}

		topic := strings.TrimSpace(e.Request.URL.Query().Get("topic"))
		if topic != "" {
			filter = "(" + filter + ") && topic_classifications ~ {:topic}"
			params["topic"] = topic
		}

		allRecords, err := app.FindRecordsByFilter("studies", filter, "", 0, 0, params)
		if err != nil {
			return apis.NewInternalServerError("Search failed", nil)
		}

		studyFields := []string{"title", "keywords", "abstract"}
		records, totalItems := rankAndPaginate(allRecords, q, studyFields, page, perPage)

		type studyResult struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Abstract string `json:"abstract"`
			Author   string `json:"author"`
			Nation   string `json:"nation"`
		}

		items := make([]studyResult, 0, len(records))
		for _, r := range records {
			items = append(items, studyResult{
				ID:       r.Id,
				Title:    r.GetString("title"),
				Abstract: r.GetString("abstract"),
				Author:   r.GetString("author"),
				Nation:   r.GetString("nation"),
			})
		}

		return e.JSON(200, map[string]interface{}{
			"page":       page,
			"perPage":    perPage,
			"totalItems": totalItems,
			"totalPages": (totalItems + perPage - 1) / perPage,
			"items":      items,
		})
	})

	// Search questions - Public
	// Assembles questions from all studies, then searches and ranks by relevance.
	// Relevance: question_text > concept > name > answer_type
	se.Router.GET("/api/search/questions", func(e *core.RequestEvent) error {
		q := strings.TrimSpace(e.Request.URL.Query().Get("q"))
		if q == "" {
			return apis.NewBadRequestError("Missing search query parameter 'q'", nil)
		}
		if len(q) > 200 {
			return apis.NewBadRequestError("Query too long (max 200 characters)", nil)
		}

		page, perPage := parsePagination(e)

		// Assemble questions from all studies
		studies, err := app.FindRecordsByFilter("studies", "", "", 0, 0)
		if err != nil {
			return apis.NewInternalServerError("Search failed", nil)
		}

		var allQuestions []Question
		for _, s := range studies {
			qs, err := assembleQuestions(app, s.Id)
			if err != nil {
				continue
			}
			allQuestions = append(allQuestions, qs...)
		}

		// Filter questions matching the query
		qLower := strings.ToLower(q)
		var matched []Question
		for _, question := range allQuestions {
			if strings.Contains(strings.ToLower(question.QuestionText), qLower) ||
				strings.Contains(strings.ToLower(question.Concept), qLower) ||
				strings.Contains(strings.ToLower(question.Name), qLower) ||
				strings.Contains(strings.ToLower(question.AnswerType), qLower) {
				matched = append(matched, question)
			}
		}

		// Rank by relevance: question_text > concept > name > answer_type
		fields := []string{"question_text", "concept", "name", "answer_type"}
		sort.SliceStable(matched, func(i, j int) bool {
			si, sj := 0, 0
			for fi, field := range fields {
				weight := len(fields) - fi
				vi, vj := "", ""
				switch field {
				case "question_text":
					vi, vj = matched[i].QuestionText, matched[j].QuestionText
				case "concept":
					vi, vj = matched[i].Concept, matched[j].Concept
				case "name":
					vi, vj = matched[i].Name, matched[j].Name
				case "answer_type":
					vi, vj = matched[i].AnswerType, matched[j].AnswerType
				}
				if strings.Contains(strings.ToLower(vi), qLower) {
					si += weight
				}
				if strings.Contains(strings.ToLower(vj), qLower) {
					sj += weight
				}
			}
			return si > sj
		})

		// Paginate
		totalItems := len(matched)
		offset := (page - 1) * perPage
		if offset > totalItems {
			offset = totalItems
		}
		end := offset + perPage
		if end > totalItems {
			end = totalItems
		}
		pageItems := matched[offset:end]

		return e.JSON(200, map[string]interface{}{
			"page":       page,
			"perPage":    perPage,
			"totalItems": totalItems,
			"totalPages": (totalItems + perPage - 1) / perPage,
			"items":      pageItems,
		})
	})

	// Questions view - Public
	// Returns a question-level view of a study's variables and groups.
	se.Router.GET("/api/studies/{id}/questions", func(e *core.RequestEvent) error {
		studyId := e.Request.PathValue("id")
		if !pocketbaseIDRegex.MatchString(studyId) {
			return apis.NewBadRequestError("Invalid ID format", nil)
		}

		if _, err := app.FindRecordById("studies", studyId); err != nil {
			return apis.NewNotFoundError("Study not found", nil)
		}

		questions, err := assembleQuestions(app, studyId)
		if err != nil {
			return apis.NewInternalServerError("Failed to assemble questions", nil)
		}

		return e.JSON(200, questions)
	})

	// Convert DDI to XLSForm - Public
	se.Router.POST("/api/convert/ddi-to-xlsform", func(e *core.RequestEvent) error {
		// Read the DDI XML from request body (1 MB limit for conversion fragments)
		ddiXML, err := io.ReadAll(io.LimitReader(e.Request.Body, maxConvertSize))
		if err != nil {
			return apis.NewBadRequestError("Failed to read request body", nil)
		}

		// Convert DDI to XLSForm
		xlsformJSON, err := converter.DDIToXLSForm(ddiXML)
		if err != nil {
			return apis.NewBadRequestError("Failed to convert DDI to XLSForm", nil)
		}

		// Set JSON response headers
		e.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		e.Response.Header().Set("X-Content-Type-Options", "nosniff")
		_, err = e.Response.Write(xlsformJSON)
		return err
	})

	// Convert XLSForm to DDI - Public
	se.Router.POST("/api/convert/xlsform-to-ddi", func(e *core.RequestEvent) error {
		// Read the XLSForm JSON from request body (1 MB limit for conversion fragments)
		xlsformJSON, err := io.ReadAll(io.LimitReader(e.Request.Body, maxConvertSize))
		if err != nil {
			return apis.NewBadRequestError("Failed to read request body", nil)
		}

		// Convert XLSForm to DDI
		ddiXML, err := converter.XLSFormToDDI(xlsformJSON)
		if err != nil {
			return apis.NewBadRequestError("Failed to convert XLSForm to DDI", nil)
		}

		// Set XML response headers
		setSecureXMLHeaders(e, "")
		_, err = e.Response.Write(ddiXML)
		return err
	})

	return nil
}

// setSecureXMLHeaders sets Content-Type and security headers on all XML responses.
// Pass a non-empty disposition to trigger a file download (e.g. "attachment; filename=...").
func setSecureXMLHeaders(e *core.RequestEvent, disposition string) {
	h := e.Response.Header()
	h.Set("Content-Type", "application/xml; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	if disposition != "" {
		h.Set("Content-Disposition", disposition)
	}
}
