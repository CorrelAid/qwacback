package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"qwacback/internal/exporter"
	"qwacback/internal/importer"
	"qwacback/internal/schematron"
	"strings"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const xmlCachePrefix = "xmlc:"
const xmlCacheTTL = 5 * time.Minute
const maxUploadSize = 50 * 1024 * 1024 // 50 MB

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

// RegisterRoutes sets up the PocketBase API routes.
func RegisterRoutes(app core.App, se *core.ServeEvent, schClient schematron.Client) error {
	// Validation API - Protected by Auth
	se.Router.POST("/api/validate", func(e *core.RequestEvent) error {
		// Enforce upload size limit before reading into memory
		e.Request.Body = http.MaxBytesReader(e.Response, e.Request.Body, maxUploadSize)

		src, _, err := e.Request.FormFile("file")
		if err != nil {
			return apis.NewBadRequestError("Missing or oversized file", err)
		}
		defer src.Close()

		xmlBytes, err := io.ReadAll(io.LimitReader(src, maxUploadSize))
		if err != nil {
			return apis.NewInternalServerError("Failed to read file", err)
		}

		// Validation service is required — reject if unavailable
		if schClient == nil {
			return apis.NewInternalServerError("Validation service is unavailable", nil)
		}

		// XML validation via NATS worker (XSD + Schematron)
		resp, err := schClient.Validate(xmlBytes)
		if err != nil {
			log.Printf("ERROR: XML validation request failed: %v", err)
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
			log.Printf("ERROR: mxj failed to parse validated XML: %v", err)
			return e.JSON(200, map[string]interface{}{
				"valid":   true,
				"message": "XML is valid against schema, but could not be parsed for import",
			})
		}

		// Insert data into collections
		if err := importer.ImportCodebookData(app, mv, xmlBytes); err != nil {
			log.Printf("ERROR: failed to import XML data: %v", err)
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
	}).Bind(apis.RequireAuth())

	// Export Study - Public (cached)
	se.Router.GET("/api/studies/{id}/export", func(e *core.RequestEvent) error {
		studyId := e.Request.PathValue("id")

		setSecureXMLHeaders(e, fmt.Sprintf("attachment; filename=\"study-%s.xml\"", studyId))

		if cached, ok := getXMLCache(app, "export:"+studyId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			_, err := e.Response.Write(cached)
			return err
		}

		study, err := app.FindRecordById("studies", studyId)
		if err != nil {
			return apis.NewNotFoundError("Study not found", err)
		}

		// Generate XML
		xmlBytes, err := exporter.ExportStudyToXML(app, study)
		if err != nil {
			return apis.NewInternalServerError("Failed to generate XML", err)
		}

		// Add XML declaration manually as mxj doesn't add it by default
		xmlBytes = append([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"), xmlBytes...)

		// Validate via NATS worker (XSD + Schematron)
		if schClient != nil {
			resp, err := schClient.Validate(xmlBytes)
			if err != nil {
				log.Printf("WARNING: XML validation unavailable on export: %v", err)
			} else if !resp.Valid {
				log.Printf("ERROR: exported XML failed validation for study %s: %v", studyId, resp.Errors)
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

		setSecureXMLHeaders(e, "")

		if cached, ok := getXMLCache(app, "var:"+varId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			_, err := e.Response.Write(cached)
			return err
		}

		record, err := app.FindRecordById("variables", varId)
		if err != nil {
			return apis.NewNotFoundError("Variable not found", err)
		}

		xmlBytes, err := exporter.ExportVariableToXML(record)
		if err != nil {
			return apis.NewInternalServerError("Failed to generate XML", err)
		}

		setXMLCache(app, "var:"+varId, xmlBytes)
		e.Response.Header().Set("X-Cache", "MISS")
		_, err = e.Response.Write(xmlBytes)
		return err
	})

	// Variable group XML fragment - Public (cached)
	se.Router.GET("/api/variable-groups/{id}/xml", func(e *core.RequestEvent) error {
		grpId := e.Request.PathValue("id")

		setSecureXMLHeaders(e, "")

		if cached, ok := getXMLCache(app, "grp:"+grpId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			_, err := e.Response.Write(cached)
			return err
		}

		record, err := app.FindRecordById("variable_groups", grpId)
		if err != nil {
			return apis.NewNotFoundError("Variable group not found", err)
		}

		xmlBytes, err := exporter.ExportVarGrpToXML(app, record)
		if err != nil {
			return apis.NewInternalServerError("Failed to generate XML", err)
		}

		setXMLCache(app, "grp:"+grpId, xmlBytes)
		e.Response.Header().Set("X-Cache", "MISS")
		_, err = e.Response.Write(xmlBytes)
		return err
	})

	// Study XML fragment - Public (cached)
	se.Router.GET("/api/studies/{id}/xml", func(e *core.RequestEvent) error {
		studyId := e.Request.PathValue("id")

		setSecureXMLHeaders(e, "")

		if cached, ok := getXMLCache(app, "study:"+studyId); ok {
			e.Response.Header().Set("X-Cache", "HIT")
			_, err := e.Response.Write(cached)
			return err
		}

		study, err := app.FindRecordById("studies", studyId)
		if err != nil {
			return apis.NewNotFoundError("Study not found", err)
		}

		xmlBytes, err := exporter.ExportStdyDscrToXML(study)
		if err != nil {
			return apis.NewInternalServerError("Failed to generate XML", err)
		}

		setXMLCache(app, "study:"+studyId, xmlBytes)
		e.Response.Header().Set("X-Cache", "MISS")
		_, err = e.Response.Write(xmlBytes)
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
