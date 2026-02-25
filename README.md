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

## DDI Data Model & Workflow

Studies are described in [DDI Codebook 2.5](https://ddialliance.org/Specification/DDI-Codebook/2.5/) XML. The application enforces a strict subset of that standard — see [DDI_MARKUP_GUIDE.md](DDI_MARKUP_GUIDE.md) for the full conventions. The key rules:

**Question types → DDI encoding**

| Format | `intrvl` | `responseDomainType` | Container |
|--------|----------|----------------------|-----------|
| `open_number` | `discrete` | `numeric` | `<var>` |
| `open_text` | `contin` | `text` | `<var>` |
| `single_choice` | `discrete` | `category` | `<var>` + `<catgry>` per option |
| `checkboxes` | `discrete` | `multiple` | `<varGrp type="multipleResp">` + binary `<var>` per option |
| `grid` | `discrete` | `category` | `<varGrp type="grid">` + `<var>` per item |

**Mandatory fields** on every `<var>` and `<varGrp>`:
- `name` attribute — `snake_case` machine identifier (column name)
- `concept` element — human-readable description of what the variable measures
- `varFormat` element — technical data type (`numeric` or `character`)
- `qstn/qstnLit` — question text as shown to the respondent

**Import → export pipeline**

```
DDI-XML file
   │
   ▼
POST /api/validate   →  XSD + Schematron validation (NATS worker)
                     →  Parse with mxj + token-based extractor
                     →  Insert into PocketBase collections
                              studies / variable_groups / variables
   │
   ▼
GET /api/studies/{id}/export
                     →  Build CodeBook struct from DB records
                     →  xml.MarshalIndent → validate → serve
```

---

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

## Database & Migrations

### Schema changes

PocketBase runs all migrations in `migrations/` in lexicographic order on startup. The rules:

| Phase | What to do |
|---|---|
| Before first production deploy | Edit the initial migration files directly |
| After first production deploy | Add a **new** migration file per change — never edit existing ones |

New migration files get a timestamp prefix so they sort after existing ones:

```bash
# example: migrations/20260301000000_add_foo_field.go
func init() {
    m.Register(func(app core.App) error {
        col, err := app.FindCollectionByNameOrId("studies")
        if err != nil { return err }
        col.Fields.Add(&core.TextField{Name: "foo"})
        return app.Save(col)
    }, nil) // second arg is optional down-migration
}
```

### Testing migrations

The Go test suite already validates migrations on every run — `tests.NewTestApp` applies all migrations against a fresh SQLite DB. The round-trip tests in `internal/exporter/exporter_test.go` also exercise the full import pipeline against that migrated schema.

To test a migration against a **copy of production data**:

```bash
# 1. Copy the production DB out of the Docker volume
docker run --rm \
  -v qwacback_pb_dataz:/data \
  -v $(pwd):/backup \
  alpine cp /data/data.db /backup/prod_backup.db

# 2. Run qwacback against the copy in a temp directory
mkdir -p /tmp/test_pb_data
cp prod_backup.db /tmp/test_pb_data/data.db
go run main.go serve --dir=/tmp/test_pb_data
# → migrations run on startup; check logs for errors
```

### Backups before breaking changes

Always back up before deploying a migration that drops or renames columns:

```bash
docker run --rm \
  -v qwacback_pb_dataz:/data \
  -v $(pwd)/backups:/backups \
  alpine cp /data/data.db /backups/data_$(date +%Y%m%d_%H%M%S).db
```

The SQLite file lives at `/app/pb_data/data.db` inside the container (volume `pb_dataz`).

---

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
