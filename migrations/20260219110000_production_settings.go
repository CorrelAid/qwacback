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

		// 4. Configure Rate Limiting
		settings.RateLimits.Enabled = true
		settings.RateLimits.Rules = []core.RateLimitRule{
			// Guest limits (stricter)
			{Label: "/api/", MaxRequests: 60, Duration: 10, Audience: core.RateLimitRuleAudienceGuest},
			// Authenticated limits
			{Label: "/api/", MaxRequests: 300, Duration: 10, Audience: core.RateLimitRuleAudienceAuth},
			// Auth endpoints (prevent brute force)
			{Label: "*:auth", MaxRequests: 5, Duration: 10},
			// Validate endpoint – stricter for guests, relaxed for authenticated
			{Label: "POST /api/validate", MaxRequests: 10, Duration: 60, Audience: core.RateLimitRuleAudienceGuest},
			{Label: "POST /api/validate", MaxRequests: 30, Duration: 60, Audience: core.RateLimitRuleAudienceAuth},
			// Import endpoint – validate + import, stricter limits
			{Label: "POST /api/import", MaxRequests: 5, Duration: 60, Audience: core.RateLimitRuleAudienceGuest},
			{Label: "POST /api/import", MaxRequests: 20, Duration: 60, Audience: core.RateLimitRuleAudienceAuth},
			// Conversion endpoints – CPU-intensive, stricter limits
			{Label: "POST /api/convert/ddi-to-xlsform", MaxRequests: 10, Duration: 60, Audience: core.RateLimitRuleAudienceGuest},
			{Label: "POST /api/convert/xlsform-to-ddi", MaxRequests: 10, Duration: 60, Audience: core.RateLimitRuleAudienceGuest},
		}

		if err := app.Save(settings); err != nil {
			return err
		}

		// 5. Collection access rules:
		// - variables/variable_groups: ViewRule public (per-record lookups work for anonymous
		//   users on question detail pages), ListRule auth-only (prevents bulk enumeration)
		// - studies: fully public
		publicRule := ""
		authRule := "@request.auth.id != ''"
		for _, name := range []string{"variable_groups", "variables"} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			c.ListRule = &authRule
			c.ViewRule = &publicRule
			if err := app.Save(c); err != nil {
				return err
			}
		}

		// // 6. Enable MFA and OTP for Superusers collection
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
