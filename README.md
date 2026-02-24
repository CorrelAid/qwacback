# qwacback - Questions Worth Asking Continuously

A robust question bank and metadata repository for civil society, built on **PocketBase** with a strict DDI-Codebook integration.

## Key Features

- **Strict DDI Validation:** XSD schema validation + custom Schematron business rules, all handled by a Java worker (Saxon HE + SchXslt2) over NATS.
- **Automated Ingestion:** Effortlessly import DDI-XML files into `studies` and `variables` collections.
- **DDI Export:** Export your studies back to validated DDI-XML format.
- **Production-Ready:**
  - **Rate Limiting:** Built-in protection for API abuse.
  - **Memory Management:** Configured with `GOMEMLIMIT` for constrained environments.
  - **Settings Encryption:** Support for encrypting sensitive app settings.
- **Auto-Seeding:** Automatically creates admin and regular user accounts and seeds the Prove It! study on first run.

## Architecture

```
                    +-----------+
POST /api/validate  |           |  Embedded NATS server
  ───────────────►  | qwacback  |  ───► validation worker (Java)
                    |  (Go)     |  Import to DB
                    +-----------+
                         │
                    +-----------+
                    | validation|  1. XSD validation (javax.xml)
                    |  worker   |  2. Schematron (SchXslt2 + Saxon HE)
                    |  (Java)   |
                    +-----------+
```

Two Docker services: `qwacback` (Go/PocketBase with embedded NATS) and `schematron-worker` (Java).

If `NATS_PORT` is not set, qwacback degrades gracefully (no validation, import only).

## Project Structure

```
internal/
  importer/     XML parsing → PocketBase records
  exporter/     PocketBase records → DDI-XML
  routes/       Custom API routes (validate, export)
  schematron/   Go NATS client (interface, mock, types)
migrations/     Schema setup, settings, user init, seed data
xml/            DDI-Codebook 2.5 XSD schemas
schematron/     Schematron rules (.sch)
seed_data/      Prove It! Toolkit seed XML
schematron-worker/  Java validation microservice (see its own README)
```

## Getting Started

### Docker Compose (recommended)

Spins up both services (qwacback + validation worker) with a single command:

```bash
docker compose up -d --build
```

Access the PocketBase Dashboard at `http://localhost:8090/_/`.

#### Default Credentials (configured in `docker-compose.yml`):
- **Admin:** `admin@example.com` / `yourpassword123`
- **User:** `user@example.com` / `userpassword123`

### Local Development

#### Prerequisites

- **Go 1.25+**
- **JDK 17+** and **Gradle** (only if you want XML validation locally)

#### Without Validation (PocketBase only)

Run the Go server directly — imports work but XSD/Schematron validation is skipped:

```bash
go run main.go serve
```

#### With Validation

Start the embedded NATS server and the Java worker:

```bash
# 1. Start qwacback with embedded NATS
NATS_PORT=4222 go run main.go serve &

# 2. Build and start the validation worker
cd schematron-worker
gradle shadowJar
NATS_URL=nats://localhost:4222 java -jar build/libs/schematron-worker-1.0.0-all.jar
```

The worker connects to the embedded NATS server in qwacback and handles all XSD + Schematron validation.

### API Endpoints

All custom endpoints require an **Authenticated** session (Bearer token).

- **POST `/api/validate`**: Validates a DDI XML file (XSD + Schematron) and imports it.
  - **Body**: `multipart/form-data` with a `file` field.
- **GET `/api/studies/{id}/export`**: Exports a study and its variables as a validated DDI-XML file download.
- **GET `/api/studies/{id}/xml`**: Returns the `<stdyDscr>` XML fragment for a study.
- **GET `/api/variables/{id}/xml`**: Returns the `<var>` XML fragment for a single variable.
- **GET `/api/variable-groups/{id}/xml`**: Returns the `<varGrp>` XML fragment for a variable group.

## Development & Testing

### Go Tests

Go tests use a mock validation client — no NATS, Java worker, or system dependencies needed.

```bash
go test -v ./internal/...
```

### Java Tests (Validation Worker)

Tests run XSD and Schematron validation directly against the real schema files. No NATS needed.

```bash
# With Docker (recommended - no local JDK required)
docker run --rm -v "$(pwd)":/app -w /app/schematron-worker gradle:8.12-jdk17 gradle test --no-daemon

# With local Gradle + JDK 17
cd schematron-worker
gradle test
```

See [schematron-worker/README.md](schematron-worker/README.md) for more details.

### Full Integration Test

Spin up both services and test the complete validation pipeline:

```bash
docker compose up -d --build

# Get a user token
TOKEN=$(curl -s -X POST http://localhost:8090/api/collections/users/auth-with-password \
  -H 'Content-Type: application/json' \
  -d '{"identity":"user@example.com","password":"userpassword123"}' | jq -r '.token')

# Validate and import an XML file
curl -X POST http://localhost:8090/api/validate \
  -H "Authorization: $TOKEN" \
  -F "file=@seed_data/prove_it.xml"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PB_ADMIN_EMAIL` | Initial Superuser Email | `admin@example.com` |
| `PB_ADMIN_PASSWORD` | Initial Superuser Password | `yourpassword123` |
| `PB_USER_EMAIL` | Initial Regular User Email | `user@example.com` |
| `PB_USER_PASSWORD` | Initial Regular User Password | `userpassword123` |
| `PB_ENCRYPTION_KEY` | 32-char key for settings encryption | (optional) |
| `GOMEMLIMIT` | Soft memory limit for the Go GC | `512MiB` |
| `TRUST_PROXY` | Set to `true` if behind a reverse proxy | `false` |
| `NATS_PORT` | Port for the embedded NATS server | (optional) |

## Resources

- **DDI Alliance:** [DDI Codebook 2.5 Specification](https://ddialliance.org/Specification/DDI-Codebook/2.5/)
- **PocketBase:** [Documentation](https://pocketbase.io/docs/)
- **SchXslt2:** [Codeberg](https://codeberg.org/SchXslt/schxslt2)
- **NATS:** [Documentation](https://docs.nats.io/)
