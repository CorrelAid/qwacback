package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "qwacback/migrations"
	qwacmcp "qwacback/internal/mcp"
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

		if natsToken == "" {
			log.Fatal("NATS_TOKEN must be set when NATS_PORT is configured")
		}
		opts := &server.Options{
			Port:          port,
			Authorization: natsToken,
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
		// Log all incoming requests to stdout for diagnostics
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			start := time.Now()
			err := e.Next()
			log.Printf("[HTTP] %s %s -> %d (%s) remote=%s",
				e.Request.Method,
				e.Request.URL.RequestURI(),
				e.Status(),
				time.Since(start).Round(time.Millisecond),
				e.Request.RemoteAddr,
			)
			if err != nil {
				log.Printf("[HTTP] error: %v", err)
			}
			return err
		})

		if err := routes.RegisterRoutes(app, se, schClient); err != nil {
			return err
		}

		// Mount MCP server (Streamable HTTP) at /mcp
		// GET is public (read-only tool discovery); POST/DELETE require superuser auth
		mcpHTTP := qwacmcp.NewHTTPServer(app)
		mcpHandler := apis.WrapStdHandler(http.Handler(mcpHTTP))
		se.Router.GET("/mcp", mcpHandler)
		se.Router.POST("/mcp", mcpHandler).Bind(apis.RequireSuperuserAuth())
		se.Router.DELETE("/mcp", mcpHandler).Bind(apis.RequireSuperuserAuth())

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
