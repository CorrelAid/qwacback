package migrations

import (
	"qwacback/internal/importer"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// studies.license_id: the SPDX ID of the Creative Commons license that
// studies.license starts with (e.g. "CC-BY-4.0"), empty for other terms, so
// clients can show the license's button without parsing the prose.
//
// Dated before the seed migration, so a fresh database has the field when the
// seed studies are imported; studies imported before it get it from their
// license here.
func init() {
	m.Register(func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.Add(&core.TextField{Name: "license_id"})
		if err := app.Save(studies); err != nil {
			return err
		}
		records, err := app.FindAllRecords(studies)
		if err != nil {
			return err
		}
		for _, r := range records {
			if id := importer.LicenseID(r.GetString("license")); id != "" {
				r.Set("license_id", id)
				if err := app.Save(r); err != nil {
					return err
				}
			}
		}
		return nil
	}, func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.RemoveByName("license_id")
		return app.Save(studies)
	})
}
