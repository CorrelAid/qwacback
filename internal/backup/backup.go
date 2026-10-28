// Package backup configures PocketBase's scheduled backups from the
// environment (#56), so a deployment sets them like its other settings.
package backup

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
)

// DefaultKeep is how many scheduled backups are kept when PB_BACKUP_KEEP is
// not set.
const DefaultKeep = 7

// ApplyFromEnv sets the backup schedule from PB_BACKUP_CRON and, when
// PB_BACKUP_S3_BUCKET is set, stores the backups in S3 instead of
// pb_data/backups:
//
//	PB_BACKUP_CRON              cron expression, e.g. "0 3 * * *" (03:00 UTC daily)
//	PB_BACKUP_KEEP              scheduled backups to keep (default 7)
//	PB_BACKUP_S3_BUCKET         bucket
//	PB_BACKUP_S3_REGION         region
//	PB_BACKUP_S3_ENDPOINT       endpoint, e.g. https://s3.eu-central-1.amazonaws.com
//	PB_BACKUP_S3_ACCESS_KEY     access key
//	PB_BACKUP_S3_SECRET         secret
//	PB_BACKUP_S3_FORCE_PATH_STYLE  "1" or "true" for path-style addressing
//
// Without PB_BACKUP_CRON it changes nothing, so backups set in the dashboard
// (Settings > Backups) stay as they are. With it, the environment owns the
// backup settings: changes made in the dashboard are overwritten on the next
// start. It reports whether it saved settings.
func ApplyFromEnv(app core.App) (bool, error) {
	cron := os.Getenv("PB_BACKUP_CRON")
	if cron == "" {
		return false, nil
	}

	keep := DefaultKeep
	if v := os.Getenv("PB_BACKUP_KEEP"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return false, fmt.Errorf("PB_BACKUP_KEEP must be a positive number, not %q", v)
		}
		keep = n
	}

	settings, err := app.Settings().Clone()
	if err != nil {
		return false, err
	}
	settings.Backups.Cron = cron
	settings.Backups.CronMaxKeep = keep
	if bucket := os.Getenv("PB_BACKUP_S3_BUCKET"); bucket != "" {
		forcePathStyle, _ := strconv.ParseBool(os.Getenv("PB_BACKUP_S3_FORCE_PATH_STYLE"))
		settings.Backups.S3 = core.S3Config{
			Enabled:        true,
			Bucket:         bucket,
			Region:         os.Getenv("PB_BACKUP_S3_REGION"),
			Endpoint:       os.Getenv("PB_BACKUP_S3_ENDPOINT"),
			AccessKey:      os.Getenv("PB_BACKUP_S3_ACCESS_KEY"),
			Secret:         os.Getenv("PB_BACKUP_S3_SECRET"),
			ForcePathStyle: forcePathStyle,
		}
	} else {
		settings.Backups.S3.Enabled = false
	}
	if err := app.Save(settings); err != nil {
		return false, fmt.Errorf("invalid backup settings: %w", err)
	}
	return true, nil
}
