package migrations

import (
	"log"
	"os"
	"path/filepath"

	"qwacback/internal/importer"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Where a study came from, so the seed sync knows which studies it owns (#56):
//   - studies.source: "seed" (a file in seed_data/) or "upload" (/api/import)
//   - studies.seed_hash: for a seed study, importer.SeedHash of its file; the
//     sync replaces the study when the file's hash differs
//
// Backfill: a study whose ID is that of a file now in seed_data/ is a seed
// study, every other one an upload. Seed studies get an empty seed_hash, so
// the next sync re-imports them from their file.
//
// Dated after every earlier migration: a database is persistent now, so a
// migration has to work on existing rows, not only on a fresh database.
func init() {
	m.Register(func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.Add(
			&core.SelectField{
				Name:      "source",
				Values:    []string{importer.SourceSeed, importer.SourceUpload},
				MaxSelect: 1,
			},
			&core.TextField{Name: "seed_hash"},
		)
		if err := app.Save(studies); err != nil {
			return err
		}

		seedIDs := map[string]bool{}
		files, _ := filepath.Glob(filepath.Join("seed_data", "*.xml"))
		for _, path := range files {
			raw, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			mv, err := mxj.NewMapXml(raw)
			if err != nil {
				continue
			}
			seedIDs[importer.StudyID(mv)] = true
		}

		records, err := app.FindAllRecords(studies)
		if err != nil {
			return err
		}
		for _, r := range records {
			if seedIDs[r.Id] {
				r.Set("source", importer.SourceSeed)
			} else {
				r.Set("source", importer.SourceUpload)
			}
			r.Set("seed_hash", "")
			if err := app.SaveNoValidate(r); err != nil {
				return err
			}
		}
		if len(records) > 0 {
			log.Printf("Backfilled source on %d studies.", len(records))
		}
		return nil
	}, func(app core.App) error {
		studies, err := app.FindCollectionByNameOrId("studies")
		if err != nil {
			return err
		}
		studies.Fields.RemoveByName("source")
		studies.Fields.RemoveByName("seed_hash")
		return app.Save(studies)
	})
}
