package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Adds `tags` to variables and variable groups: the <concept> elements after
// the first one in the DDI, as [{"lang": "de", "text": "Vertrauen"}]. They are
// extra search terms, typically the concept in the other language (#19).
//
// Dated before the seed migration so a fresh database has the field when the
// seed studies are imported; PocketBase still applies it to existing
// databases, which have not run it yet.
func init() {
	m.Register(func(app core.App) error {
		for _, name := range []string{"variables", "variable_groups"} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			c.Fields.Add(&core.JSONField{Name: "tags", MaxSize: 65536}) // 64 KB
			if err := app.Save(c); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		for _, name := range []string{"variables", "variable_groups"} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			c.Fields.RemoveByName("tags")
			if err := app.Save(c); err != nil {
				return err
			}
		}
		return nil
	})
}
