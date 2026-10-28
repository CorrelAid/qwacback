package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// This migration used to import seed_data/*.xml once, into an empty
// database. With a persistent database it would never pick up later changes
// to seed_data/, so the seed sync (internal/seed, run on every start) does
// that now (#56).
//
// It stays registered as a no-op: databases that applied it keep a matching
// entry, and a fresh database runs it without effect.
func init() {
	m.Register(func(app core.App) error {
		return nil
	}, nil)
}
