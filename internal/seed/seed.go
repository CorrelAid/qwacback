// Package seed keeps the seed studies in line with seed_data/ (#56).
//
// The database is persistent, so seed_data/ can't be imported once into an
// empty database. Sync runs on every start instead: it imports new seed files,
// replaces a study whose file changed, skips unchanged ones and deletes seed
// studies whose file is gone. Studies uploaded through /api/import are never
// touched.
package seed

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"qwacback/internal/importer"
	"qwacback/internal/schematron"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
)

// Result says what a Sync did, by seed file (or, for Deleted, study title).
type Result struct {
	Imported  []string
	Replaced  []string
	Unchanged []string
	Deleted   []string
	// Skipped are seed files whose title belongs to a study that isn't a seed
	// study, e.g. an upload. The sync refuses to overwrite it.
	Skipped []string
}

// Changed says whether the sync changed or removed existing studies, so
// cached exports may be stale.
func (r Result) Changed() bool {
	return len(r.Replaced) > 0 || len(r.Deleted) > 0
}

type seedFile struct {
	path string
	raw  []byte
	mv   mxj.Map
	id   string
	hash string
}

// Sync brings the seed studies in line with the *.xml files in dir. When
// client is not nil, every file it imports must pass validation first; a
// rejected file fails the sync before anything is written. A missing dir is
// not an error: Sync then leaves the database as it is.
func Sync(app core.App, dir string, client schematron.Client) (Result, error) {
	var res Result

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		log.Printf("Seed sync: %s not found, leaving the studies as they are.", dir)
		return res, nil
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*.xml"))
	if err != nil {
		return res, fmt.Errorf("failed to scan %s: %w", dir, err)
	}

	// Read every file first, so a broken one fails the sync before any change.
	var files []seedFile
	byID := map[string]string{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return res, fmt.Errorf("failed to read %s: %w", path, err)
		}
		mv, err := mxj.NewMapXml(raw)
		if err != nil {
			return res, fmt.Errorf("failed to parse %s: %w", path, err)
		}
		if importer.StudyTitle(mv) == "" {
			return res, fmt.Errorf("%s has no study title", path)
		}
		id := importer.StudyID(mv)
		if other, dup := byID[id]; dup {
			return res, fmt.Errorf("%s and %s have the same study title", other, path)
		}
		byID[id] = path
		files = append(files, seedFile{path, raw, mv, id, importer.SeedHash(raw)})
	}

	// Decide what to do with each file.
	var toImport, toReplace []seedFile
	for _, f := range files {
		existing, err := app.FindRecordById("studies", f.id)
		switch {
		case err != nil:
			toImport = append(toImport, f)
		case existing.GetString("source") != importer.SourceSeed:
			// A seed and an upload with the same title: refuse, don't
			// overwrite. Not fatal, so the app still starts and the upload
			// can be renamed or deleted.
			log.Printf("WARNING: seed sync skips %s: study %s %q is not a seed study (source %q).",
				f.path, f.id, existing.GetString("title"), existing.GetString("source"))
			res.Skipped = append(res.Skipped, f.path)
		case existing.GetString("seed_hash") == f.hash:
			res.Unchanged = append(res.Unchanged, f.path)
		default:
			toReplace = append(toReplace, f)
		}
	}

	if client != nil && len(toImport)+len(toReplace) > 0 {
		if err := client.WaitForWorker(60 * time.Second); err != nil {
			return res, fmt.Errorf("schematron worker unavailable: %w", err)
		}
		for _, f := range append(append([]seedFile{}, toImport...), toReplace...) {
			resp, err := client.Validate(f.raw)
			if err != nil {
				return res, fmt.Errorf("seed data %s: Schematron validation unavailable: %w", f.path, err)
			}
			if !resp.Valid {
				return res, fmt.Errorf("seed data %s failed Schematron validation: %v", f.path, resp.Errors)
			}
		}
	}

	for _, f := range toImport {
		if err := app.RunInTransaction(func(tx core.App) error {
			_, err := importer.ImportStudy(tx, f.mv, f.raw, importer.SourceSeed, f.hash)
			return err
		}); err != nil {
			return res, fmt.Errorf("failed to import %s: %w", f.path, err)
		}
		log.Printf("Seed sync: imported %s.", f.path)
		res.Imported = append(res.Imported, f.path)
	}

	for _, f := range toReplace {
		if _, err := importer.ReplaceStudy(app, f.mv, f.raw, importer.SourceSeed, f.hash); err != nil {
			return res, fmt.Errorf("failed to replace %s: %w", f.path, err)
		}
		log.Printf("Seed sync: replaced %s.", f.path)
		res.Replaced = append(res.Replaced, f.path)
	}

	// Seed studies whose file is gone. Their groups and variables go with
	// them (cascade delete).
	seeds, err := app.FindAllRecords("studies")
	if err != nil {
		return res, err
	}
	for _, s := range seeds {
		if s.GetString("source") != importer.SourceSeed || byID[s.Id] != "" {
			continue
		}
		if err := app.Delete(s); err != nil {
			return res, fmt.Errorf("failed to delete seed study %q: %w", s.GetString("title"), err)
		}
		log.Printf("Seed sync: deleted %q, its seed file is gone.", s.GetString("title"))
		res.Deleted = append(res.Deleted, s.GetString("title"))
	}

	log.Printf("Seed sync: %d imported, %d replaced, %d unchanged, %d deleted, %d skipped.",
		len(res.Imported), len(res.Replaced), len(res.Unchanged), len(res.Deleted), len(res.Skipped))
	return res, nil
}
