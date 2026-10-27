package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// What a client must show when it reuses a study's questions:
//   - studies.license: stdyDscr/dataAccs/useStmt/conditions, the terms of use
//     (e.g. "CC BY 4.0"); empty when the codebook states none
//   - studies.copyright: citation/prodStmt/copyright
//   - studies.citation: citation/biblCit, the citation to give with the items
//
// Dated before the seed migration, like codebook, so a fresh database has the
// fields when the seed studies are imported.
func init() {
	m.Register(func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.Add(
			&core.TextField{Name: "license"},
			&core.TextField{Name: "copyright"},
			&core.TextField{Name: "citation"},
		)
		return app.Save(studies)
	}, func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.RemoveByName("license")
		studies.Fields.RemoveByName("copyright")
		studies.Fields.RemoveByName("citation")
		return app.Save(studies)
	})
}
