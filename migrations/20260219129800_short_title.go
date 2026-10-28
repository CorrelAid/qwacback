package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// studies.short_title: the study's short name from titlStmt/altTitl, e.g.
// "CVTS7", for lists and tags; empty when the codebook has none. The title
// stays the full name, since the record IDs are derived from it.
//
// Dated before the seed migration, so a fresh database has the field when the
// seed studies are imported.
func init() {
	m.Register(func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.Add(&core.TextField{Name: "short_title"})
		return app.Save(studies)
	}, func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.RemoveByName("short_title")
		return app.Save(studies)
	})
}
