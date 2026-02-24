package migrations

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"qwacback/internal/importer"
	"qwacback/internal/schematron"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Only seed if studies collection is empty
		collection, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}

		total, err := app.CountRecords(collection.Id)
		if err != nil {
			return err
		}

		if total > 0 {
			log.Println("Skipping seeding: studies already exist.")
			return nil
		}

		seedFiles, err := filepath.Glob("seed_data/*.xml")
		if err != nil {
			return fmt.Errorf("failed to scan seed_data directory: %w", err)
		}
		if len(seedFiles) == 0 {
			log.Println("No XML files found in seed_data/.")
			return nil
		}

		// Optional NATS client for validation
		var client schematron.Client
		natsPort := os.Getenv("NATS_PORT")
		if natsPort != "" {
			c, err := schematron.NewNatsClient("nats://localhost:"+natsPort, os.Getenv("NATS_TOKEN"))
			if err != nil {
				log.Printf("Warning: Could not connect to NATS for seed validation: %v", err)
			} else {
				defer c.Close()
				client = c
			}
		}

		for _, path := range seedFiles {
			xmlData, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Warning: could not read seed file %s: %v", path, err)
				continue
			}

			if client != nil {
				resp, err := client.Validate(xmlData)
				if err != nil {
					log.Printf("Warning: Schematron validation unavailable for %s: %v", path, err)
				} else if !resp.Valid {
					return fmt.Errorf("seed data %s failed Schematron validation: %v", path, resp.Errors)
				} else {
					log.Printf("Seed data %s passed Schematron validation.", path)
				}
			}

			mv, err := mxj.NewMapXml(xmlData)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}

			if err := importer.ImportCodebookData(app, mv, xmlData); err != nil {
				return fmt.Errorf("failed to import %s: %w", path, err)
			}

			log.Printf("Seeded %s successfully.", path)
		}

		return nil
	}, nil)
}
