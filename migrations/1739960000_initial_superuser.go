package migrations

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		email := os.Getenv("PB_ADMIN_EMAIL")
		password := os.Getenv("PB_ADMIN_PASSWORD")

		if email == "" || password == "" {
			log.Println("Skipping initial superuser creation: PB_ADMIN_EMAIL or PB_ADMIN_PASSWORD not set.")
			return nil
		}

		collection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
		if err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter(collection.Id, "email != '__pbinstaller@example.com'", "", 0, 0)
		if err != nil {
			return err
		}

		if len(records) > 0 {
			log.Println("Skipping initial superuser creation: Real superusers already exist.")
			return nil
		}

		record := core.NewRecord(collection)
		record.SetEmail(email)
		record.SetPassword(password)

		if err := app.Save(record); err != nil {
			return err
		}

		log.Printf("Successfully created initial superuser: %s\n", email)
		return nil
	}, nil)
}
