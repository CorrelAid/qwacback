package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// variables.grid_note: a sub-heading printed between a grid's rows, before
// this variable (e.g. "davon: Geschlecht" before "Männlich"), from the
// codebook's <notes type="cdl:grid_note"> (formtransform >= 0.8.0). Its other
// languages are in translations under "grid_note".
//
// Dated before the seed migration, so a fresh database has the field when the
// seed studies are imported.
func init() {
	m.Register(func(app core.App) error {
		variables, err := app.FindCollectionByNameOrId("variables")
		if err != nil {
			return err
		}
		variables.Fields.Add(&core.TextField{Name: "grid_note"})
		return app.Save(variables)
	}, func(app core.App) error {
		variables, err := app.FindCollectionByNameOrId("variables")
		if err != nil {
			return err
		}
		variables.Fields.RemoveByName("grid_note")
		return app.Save(variables)
	})
}
