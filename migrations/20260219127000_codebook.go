package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// A CDL codebook carries the whole form (formtransform v0.7.0), much of it in
// typed <notes> the records don't model. The codebook as imported is the
// source for every export, so nothing in it is lost (#37):
//   - studies.codebook: the imported XML; hidden, so record listings stay small
//   - variables.hint: qstn/postQTxt, the XLSForm hint
//   - variables.universe: the var's <universe>, its skip logic as prose
//
// Dated before the seed migration, like concept_tags, so a fresh database
// has the fields when the seed studies are imported.
func init() {
	m.Register(func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.Add(&core.TextField{Name: "codebook", Hidden: true, Max: 50 * 1024 * 1024})
		if err := app.Save(studies); err != nil {
			return err
		}
		vars, err := app.FindCollectionByNameOrId("variables")
		if err != nil {
			return err
		}
		vars.Fields.Add(&core.TextField{Name: "hint"}, &core.TextField{Name: "universe"})
		return app.Save(vars)
	}, func(app core.App) error {
		if c, err := app.FindCollectionByNameOrId("studies"); err == nil {
			c.Fields.RemoveByName("codebook")
			if err := app.Save(c); err != nil {
				return err
			}
		}
		c, err := app.FindCollectionByNameOrId("variables")
		if err != nil {
			return err
		}
		c.Fields.RemoveByName("hint")
		c.Fields.RemoveByName("universe")
		return app.Save(c)
	})
}
