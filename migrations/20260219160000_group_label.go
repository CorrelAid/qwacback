package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// variable_groups.label: the group's short printed name from varGrp/labl,
// e.g. a scale's name ("ATTARI-12") or a question number; empty when the
// codebook has none. formtransform >= 0.9.0 writes it from the XLSForm
// `label_short` column; other languages go to translations[lang].label.
//
// Dated after every other migration. No backfill: CDL Schematron rejected
// labl on varGrp before formtransform 0.9.0, so no study imported earlier
// carries one, and the seed sync re-imports the seed studies
// (importer.Revision).
func init() {
	m.Register(func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("variable_groups")
		if err != nil {
			return err
		}
		groups.Fields.Add(&core.TextField{Name: "label"})
		return app.Save(groups)
	}, func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("variable_groups")
		if err != nil {
			return err
		}
		groups.Fields.RemoveByName("label")
		return app.Save(groups)
	})
}
