package migrations

import (
	"qwacback/internal/importer"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// studies.subtitle: the study's subtitle from titlStmt/subTitl, e.g.
// "2. Revision"; empty when the codebook has none (#64).
//
// Dated after every other migration. Studies imported before it, such as the
// seed studies of a fresh database, get it (and short_title, read the same
// way now) from their stored codebook here.
func init() {
	m.Register(func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.Add(&core.TextField{Name: "subtitle"})
		if err := app.Save(studies); err != nil {
			return err
		}
		records, err := app.FindAllRecords(studies)
		if err != nil {
			return err
		}
		for _, r := range records {
			mv, err := mxj.NewMapXml([]byte(r.GetString("codebook")))
			if err != nil {
				continue
			}
			shortTitle, subtitle := importer.StudyTitles(mv)
			if shortTitle == r.GetString("short_title") && subtitle == "" {
				continue
			}
			r.Set("short_title", shortTitle)
			r.Set("subtitle", subtitle)
			if err := app.Save(r); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.RemoveByName("subtitle")
		return app.Save(studies)
	})
}
