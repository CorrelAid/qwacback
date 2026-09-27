package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Multilingual DDI (formtransform#135) has each text in the base language,
// untagged, plus one xml:lang sibling per other language (#35):
//   - studies.language: the base language (codeBook/@xml:lang)
//   - variables.translations: {"en": {"question", "prequestion_text",
//     "ivu_instructions", "categories": {"<value>": "<label>"}}}
//   - variable_groups.translations: {"en": {"description"}}
//
// Dated before the seed migration, like concept_tags, so a fresh database
// has the fields when the seed studies are imported.
func init() {
	m.Register(func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.Add(&core.TextField{Name: "language"})
		if err := app.Save(studies); err != nil {
			return err
		}
		for _, name := range []string{"variables", "variable_groups"} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			c.Fields.Add(&core.JSONField{Name: "translations", MaxSize: 524288}) // 512 KB
			if err := app.Save(c); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		if c, err := app.FindCollectionByNameOrId("studies"); err == nil {
			c.Fields.RemoveByName("language")
			if err := app.Save(c); err != nil {
				return err
			}
		}
		for _, name := range []string{"variables", "variable_groups"} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			c.Fields.RemoveByName("translations")
			if err := app.Save(c); err != nil {
				return err
			}
		}
		return nil
	})
}
