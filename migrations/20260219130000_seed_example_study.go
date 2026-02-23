package migrations

import (
	"fmt"
	"log"
	"os"
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

		// Read the Prove It! seed data
		xmlData, err := os.ReadFile("seed_data/prove_it.xml")
		if err != nil {
			log.Printf("Warning: could not find seed file seed_data/prove_it.xml: %v", err)
			return nil
		}

		// Validation via embedded NATS (optional, only if NATS_PORT is set)
		natsPort := os.Getenv("NATS_PORT")
		if natsPort != "" {
			client, err := schematron.NewNatsClient("nats://localhost:" + natsPort)
			if err != nil {
				log.Printf("Warning: Could not connect to NATS for seed validation: %v", err)
			} else {
				defer client.Close()
				resp, err := client.Validate(xmlData)
				if err != nil {
					log.Printf("Warning: Schematron validation unavailable for seed: %v", err)
				} else if !resp.Valid {
					return fmt.Errorf("seed data failed Schematron validation: %v", resp.Errors)
				} else {
					log.Println("Seed data passed Schematron validation.")
				}
			}
		}

		// Parse and import
		mv, err := mxj.NewMapXml(xmlData)
		if err != nil {
			return err
		}

		return importer.ImportCodebookData(app, mv)
	}, nil)
}
