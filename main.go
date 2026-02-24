package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "qwacback/migrations"
	"qwacback/internal/routes"
	"qwacback/internal/schematron"
)

func main() {
	app := pocketbase.New()

	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Dashboard
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	// Start embedded NATS server if NATS_PORT is set
	var schClient schematron.Client
	natsPort := os.Getenv("NATS_PORT")
	natsToken := os.Getenv("NATS_TOKEN")
	if natsPort != "" {
		port, err := strconv.Atoi(natsPort)
		if err != nil {
			log.Fatalf("Invalid NATS_PORT: %v", err)
		}

		opts := &server.Options{
			Port: port,
		}
		if natsToken != "" {
			opts.Authorization = natsToken
		} else {
			log.Println("WARNING: NATS_TOKEN not set; NATS server is unauthenticated")
		}

		ns, err := server.NewServer(opts)
		if err != nil {
			log.Fatalf("Failed to create embedded NATS server: %v", err)
		}
		ns.Start()

		if !ns.ReadyForConnections(10 * time.Second) {
			log.Fatal("Embedded NATS server failed to start")
		}
		log.Printf("Embedded NATS server listening on port %d", port)

		natsURL := fmt.Sprintf("nats://localhost:%d", port)
		client, err := schematron.NewNatsClient(natsURL, natsToken)
		if err != nil {
			log.Printf("WARNING: Could not connect to embedded NATS: %v", err)
		} else {
			schClient = client
			defer client.Close()
		}
		defer ns.Shutdown()
	}

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		if err := routes.RegisterRoutes(app, se, schClient); err != nil {
			return err
		}
		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
