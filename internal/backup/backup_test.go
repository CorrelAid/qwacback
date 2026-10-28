package backup

import (
	"testing"

	_ "qwacback/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

func newApp(t *testing.T) *tests.TestApp {
	t.Helper()
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	return app
}

func TestApplyFromEnvWithoutCronChangesNothing(t *testing.T) {
	t.Setenv("PB_BACKUP_CRON", "")
	app := newApp(t)
	app.Settings().Backups.Cron = "0 1 * * *" // as set in the dashboard
	ok, err := ApplyFromEnv(app)
	if err != nil || ok {
		t.Fatalf("ApplyFromEnv = %v, %v", ok, err)
	}
	if app.Settings().Backups.Cron != "0 1 * * *" {
		t.Errorf("cron %q", app.Settings().Backups.Cron)
	}
}

func TestApplyFromEnv(t *testing.T) {
	t.Setenv("PB_BACKUP_CRON", "0 3 * * *")
	t.Setenv("PB_BACKUP_KEEP", "")
	t.Setenv("PB_BACKUP_S3_BUCKET", "qwac-backups")
	t.Setenv("PB_BACKUP_S3_REGION", "eu-central-1")
	t.Setenv("PB_BACKUP_S3_ENDPOINT", "https://s3.example.org")
	t.Setenv("PB_BACKUP_S3_ACCESS_KEY", "key")
	t.Setenv("PB_BACKUP_S3_SECRET", "secret")
	t.Setenv("PB_BACKUP_S3_FORCE_PATH_STYLE", "true")
	app := newApp(t)
	if ok, err := ApplyFromEnv(app); err != nil || !ok {
		t.Fatalf("ApplyFromEnv = %v, %v", ok, err)
	}
	b := app.Settings().Backups
	if b.Cron != "0 3 * * *" || b.CronMaxKeep != DefaultKeep {
		t.Errorf("cron %q keep %d", b.Cron, b.CronMaxKeep)
	}
	if !b.S3.Enabled || b.S3.Bucket != "qwac-backups" || b.S3.Secret != "secret" || !b.S3.ForcePathStyle {
		t.Errorf("s3 %+v", b.S3)
	}
}

func TestApplyFromEnvRejectsBadValues(t *testing.T) {
	for _, env := range []map[string]string{
		{"PB_BACKUP_CRON": "0 3 * * *", "PB_BACKUP_KEEP": "none"},
		{"PB_BACKUP_CRON": "every night"},
	} {
		t.Run(env["PB_BACKUP_CRON"]+env["PB_BACKUP_KEEP"], func(t *testing.T) {
			t.Setenv("PB_BACKUP_KEEP", "")
			t.Setenv("PB_BACKUP_S3_BUCKET", "")
			for k, v := range env {
				t.Setenv(k, v)
			}
			if _, err := ApplyFromEnv(newApp(t)); err == nil {
				t.Error("accepted")
			}
		})
	}
}
