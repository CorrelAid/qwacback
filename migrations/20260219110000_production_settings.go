package migrations

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings := app.Settings()

		// 1. Set App Name and URL from environment or defaults
		appName := os.Getenv("APP_NAME")
		if appName == "" {
			appName = "QWAC-Back"
		}
		settings.Meta.AppName = appName

		appURL := os.Getenv("APP_URL")
		if appURL != "" {
			settings.Meta.AppURL = appURL
		}

		// 3. Configure Logs
		settings.Logs.MaxDays = 30

		if err := app.Save(settings); err != nil {
			return err
		}

		// // 4. Enable MFA and OTP for Superusers collection
		// superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
		// if err != nil {
		// 	return err
		// }

		// superusers.MFA.Enabled = true
		// superusers.OTP.Enabled = true

		// if err := app.Save(superusers); err != nil {
		// 	return err
		// }

		log.Println("Successfully applied production settings migration.")
		return nil
	}, nil)
}
