package routes

import (
	"fmt"
	"io"
	"log"
	"qwacback/internal/exporter"
	"qwacback/internal/importer"
	"qwacback/internal/schematron"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterRoutes sets up the PocketBase API routes.
func RegisterRoutes(app core.App, se *core.ServeEvent, schClient schematron.Client) error {
	// Validation API - Protected by Auth
	se.Router.POST("/api/validate", func(e *core.RequestEvent) error {
		src, _, err := e.Request.FormFile("file")
		if err != nil {
			return apis.NewBadRequestError("Missing file", err)
		}
		defer src.Close()

		xmlBytes, err := io.ReadAll(src)
		if err != nil {
			return apis.NewInternalServerError("Failed to read file", err)
		}

		// XML validation via NATS worker (XSD + Schematron)
		if schClient != nil {
			resp, err := schClient.Validate(xmlBytes)
			if err != nil {
				log.Printf("WARNING: XML validation unavailable: %v", err)
			} else if !resp.Valid {
				return e.JSON(400, map[string]interface{}{
					"valid":  false,
					"errors": resp.Errors,
				})
			}
		}

		// Use mxj to parse the XML
		mv, err := mxj.NewMapXml(xmlBytes)
		if err != nil {
			return e.JSON(200, map[string]interface{}{
				"valid":   true,
				"message": "XML is valid against schema, but mxj failed to parse",
				"error":   err.Error(),
			})
		}

		// Insert data into collections
		if err := importer.ImportCodebookData(app, mv, xmlBytes); err != nil {
			return e.JSON(200, map[string]interface{}{
				"valid":   true,
				"message": "XML is valid, but failed to insert into database",
				"error":   err.Error(),
			})
		}

		return e.JSON(200, map[string]interface{}{
			"valid":   true,
			"message": "XML is valid and imported successfully",
		})
	}).Bind(apis.RequireAuth())

	// Export Study - Protected by Auth
	se.Router.GET("/api/studies/{id}/export", func(e *core.RequestEvent) error {
		studyId := e.Request.PathValue("id")

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
				return apis.NewInternalServerError("Generated XML failed validation", nil)
			}
		}

		// Return as file download
		e.Response.Header().Set("Content-Type", "application/xml")
		e.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"study-%s.xml\"", studyId))

		_, err = e.Response.Write(xmlBytes)
		return err
	}).Bind(apis.RequireAuth())

	// Variable XML fragment
	se.Router.GET("/api/variables/{id}/xml", func(e *core.RequestEvent) error {
		record, err := app.FindRecordById("variables", e.Request.PathValue("id"))
		if err != nil {
			return apis.NewNotFoundError("Variable not found", err)
		}

		xmlBytes, err := exporter.ExportVariableToXML(record)
		if err != nil {
			return apis.NewInternalServerError("Failed to generate XML", err)
		}

		e.Response.Header().Set("Content-Type", "application/xml")
		_, err = e.Response.Write(xmlBytes)
		return err
	}).Bind(apis.RequireAuth())

	// Variable group XML fragment
	se.Router.GET("/api/variable-groups/{id}/xml", func(e *core.RequestEvent) error {
		record, err := app.FindRecordById("variable_groups", e.Request.PathValue("id"))
		if err != nil {
			return apis.NewNotFoundError("Variable group not found", err)
		}

		xmlBytes, err := exporter.ExportVarGrpToXML(app, record)
		if err != nil {
			return apis.NewInternalServerError("Failed to generate XML", err)
		}

		e.Response.Header().Set("Content-Type", "application/xml")
		_, err = e.Response.Write(xmlBytes)
		return err
	}).Bind(apis.RequireAuth())

	// Study XML fragment (stdyDscr only)
	se.Router.GET("/api/studies/{id}/xml", func(e *core.RequestEvent) error {
		study, err := app.FindRecordById("studies", e.Request.PathValue("id"))
		if err != nil {
			return apis.NewNotFoundError("Study not found", err)
		}

		xmlBytes, err := exporter.ExportStdyDscrToXML(study)
		if err != nil {
			return apis.NewInternalServerError("Failed to generate XML", err)
		}

		e.Response.Header().Set("Content-Type", "application/xml")
		_, err = e.Response.Write(xmlBytes)
		return err
	}).Bind(apis.RequireAuth())

	return nil
}
