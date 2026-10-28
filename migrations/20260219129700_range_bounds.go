package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// variables.range_bounds: a range variable's bounds as the codebook states
// them, {"min": 0, "max": 10, "step": 0.5}, each key only when stated. min
// and max come from <valrng><range min max>, step from the
// <notes type="cdl:parameters"> formtransform writes from the XLSForm
// parameters. JSON rather than number fields, because a number field can't
// tell an unstated bound from 0.
//
// Dated before the seed migration, so a fresh database has the field when the
// seed studies are imported.
func init() {
	m.Register(func(app core.App) error {
		variables, err := app.FindCollectionByNameOrId("variables")
		if err != nil {
			return err
		}
		variables.Fields.Add(&core.JSONField{Name: "range_bounds", MaxSize: 1024})
		return app.Save(variables)
	}, func(app core.App) error {
		variables, err := app.FindCollectionByNameOrId("variables")
		if err != nil {
			return err
		}
		variables.Fields.RemoveByName("range_bounds")
		return app.Save(variables)
	})
}
