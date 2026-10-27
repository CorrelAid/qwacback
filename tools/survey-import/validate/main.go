// Command validate checks DDI codebooks against the pinned schematron worker
// (XSD + CDL Schematron), the same check /api/import and the seed test run.
//
//	NATS_URL=nats://127.0.0.1:4223 NATS_TOKEN=surveyimport go run ./tools/survey-import/validate FILE...
package main

import (
	"fmt"
	"os"
	"time"

	"qwacback/internal/schematron"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: validate FILE...")
		os.Exit(2)
	}
	client, err := schematron.NewNatsClient(os.Getenv("NATS_URL"), os.Getenv("NATS_TOKEN"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(2)
	}
	defer client.Close()
	if err := client.WaitForWorker(10 * time.Second); err != nil {
		fmt.Fprintln(os.Stderr, "worker:", err)
		os.Exit(2)
	}
	failed := false
	for _, path := range os.Args[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			failed = true
			continue
		}
		resp, err := client.Validate(data)
		if err != nil {
			fmt.Printf("%s: ERROR %v\n", path, err)
			failed = true
			continue
		}
		if resp.Valid {
			fmt.Printf("%s: valid\n", path)
			continue
		}
		failed = true
		fmt.Printf("%s: INVALID\n", path)
		for _, e := range resp.Errors {
			fmt.Printf("  %+v\n", e)
		}
	}
	if failed {
		os.Exit(1)
	}
}
