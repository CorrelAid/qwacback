package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// concept is optional on variables and variable groups: formtransform v0.8.0
// writes <concept> only from an XLSForm concept column, instead of copying the
// question label into it (formtransform#173).
//
// Dated before the seed migration, so a fresh database accepts the seed
// studies without concepts.
func init() {
	setRequired := func(app core.App, required bool) error {
		for _, name := range []string{"variables", "variable_groups"} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			if f, ok := c.Fields.GetByName("concept").(*core.TextField); ok {
				f.Required = required
			}
			if err := app.Save(c); err != nil {
				return err
			}
		}
		return nil
	}
	m.Register(func(app core.App) error {
		return setRequired(app, false)
	}, func(app core.App) error {
		return setRequired(app, true)
	})
}
