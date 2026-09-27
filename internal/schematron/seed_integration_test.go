package schematron

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Every seed codebook passes the pinned worker's XSD and CDL Schematron, so
// a formtransform bump that tightens the rules fails here, not on deploy.
func TestIntegration_SeedDataValidates(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set; skipping integration test")
	}
	client, err := NewNatsClient(natsURL, os.Getenv("NATS_TOKEN"))
	if err != nil {
		t.Skipf("Could not connect to NATS: %v", err)
	}
	defer client.Close()
	if err := client.WaitForWorker(5 * time.Second); err != nil {
		t.Skipf("Schematron worker not available: %v", err)
	}

	seeds, _ := filepath.Glob(filepath.Join("..", "..", "seed_data", "*.xml"))
	if len(seeds) == 0 {
		t.Fatal("no seed files")
	}
	for _, path := range seeds {
		t.Run(filepath.Base(path), func(t *testing.T) {
			xmlData, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Validate(xmlData)
			if err != nil {
				t.Fatal(err)
			}
			if !resp.Valid {
				t.Errorf("invalid: %+v", resp.Errors)
			}
		})
	}
}
