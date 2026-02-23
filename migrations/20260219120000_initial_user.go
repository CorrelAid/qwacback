package migrations

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		email := os.Getenv("PB_USER_EMAIL")
		password := os.Getenv("PB_USER_PASSWORD")

		if email == "" || password == "" {
			log.Println("Skipping initial regular user creation: PB_USER_EMAIL or PB_USER_PASSWORD not set.")
			return nil
		}

		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// Check if user already exists
		existing, _ := app.FindAuthRecordByEmail("users", email)
		if existing != nil {
			log.Printf("Skipping initial regular user creation: User %s already exists.\n", email)
			return nil
		}

		record := core.NewRecord(collection)
		record.SetEmail(email)
		record.SetPassword(password)
		record.Set("verified", true)

		if err := app.Save(record); err != nil {
			return err
		}

		log.Printf("Successfully created initial regular user: %s\n", email)
		return nil
	}, nil)
}
