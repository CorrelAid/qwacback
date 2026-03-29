# qwacback - Questions Worth Asking Continuously

[![AI-Assisted](https://img.shields.io/badge/AI--assisted-Claude%20Code-blueviolet?logo=anthropic&logoColor=white)](./AI_DISCLOSURE.md)

A question bank and metadata repository for civil society, built on **PocketBase** with a strict DDI-Codebook integration.

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
  converter/    Bidirectional DDI ↔ XLSForm conversion
  examples/     Static answer type examples (XLSForm + DDI)
  exporter/     PocketBase records → DDI-XML
  importer/     XML parsing → PocketBase records
  routes/       Custom API routes (validate, export, convert, examples)
  schematron/   Go NATS client (interface, mock, types)
migrations/     Schema setup, settings, user init, seed data
xml/            DDI-Codebook 2.5 XSD schemas
schematron/     Schematron rules (.sch)
seed_data/      Prove It! Toolkit seed XML
schematron-worker/  Java validation microservice (see its own README)
```

## DDI Data Model & Workflow

Studies are described in [DDI Codebook 2.5](https://ddialliance.org/Specification/DDI-Codebook/2.5/) XML. The application enforces a strict subset of that standard — see [DDI_MARKUP_GUIDE.md](DDI_MARKUP_GUIDE.md) for the full conventions. The key rules:

**Answer types → DDI encoding**

| `answer_type` | `intrvl` | `responseDomainType` | Container |
|--------|----------|----------------------|-----------|
| `integer` | `contin` | `numeric` | `<var>` |
| `text` | `discrete` | `text` | `<var>` |
| `single_choice` | `discrete` | `category` | `<var>` + `<catgry>` per option |
| `multiple_choice` | `discrete` | `multiple` | `<varGrp type="multipleResp">` + binary `<var>` per option |
| `grid` | `discrete` | `category` | `<varGrp type="grid">` + `<var>` per item |

**Subcategory flags** (booleans on `single_choice` and `multiple_choice`):

| Flag | Effect |
|------|--------|
| `has_other` | A companion `_other` text variable exists for free-text specification |
| `has_long_list` | Categories come from an external code list via `concept/@vocab` |

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

#### Validation & Import

- **POST `/api/validate`**: Validates a DDI XML file (XSD + Schematron) and imports it.
  - **Body**: `multipart/form-data` with a `file` field.

#### XML Export

- **GET `/api/studies/{id}/export`**: Exports a study and its variables as a validated DDI-XML file download.
- **GET `/api/variables/{id}/xml`**: Returns the `<var>` XML fragment for a single variable.
- **GET `/api/variable-groups/{id}/xml`**: Returns the `<varGrp>` XML fragment for a variable group.

#### Format Conversion

- **POST `/api/convert/ddi-to-xlsform`**: Converts a DDI XML fragment (`<var>` or `<varGrp>`) to XLSForm JSON format.
  - **Body**: DDI XML fragment
  - **Content-Type**: `application/xml` or `text/xml`
  - **Response**: XLSForm JSON
- **POST `/api/convert/xlsform-to-ddi`**: Converts an XLSForm JSON question or group to DDI XML format.
  - **Body**: XLSForm JSON
  - **Content-Type**: `application/json`
  - **Response**: DDI XML fragment

For detailed documentation on the conversion endpoints, see [CONVERSION_API.md](CONVERSION_API.md).

#### Search

- **GET `/api/search/studies?q=<term>`**: Search studies by title, keywords, and abstract. Results ranked by relevance (title > keywords > abstract).
  - **Optional filter**: `&topic=<classification>` — restrict to studies matching a topic classification.
  - **Pagination**: `&page=1&perPage=20` (default 20, max 100).

- **GET `/api/search/questions?q=<term>`**: Search variables by question text, concept, name, prequestion text, categories, and answer type. Results ranked by relevance (question > concept > name > prequestion_text > categories > answer_type).
  - **Pagination**: `&page=1&perPage=20` (default 20, max 100).

#### Examples

- **GET `/api/examples`**: Returns answer type examples as a JSON array. Each example includes XLSForm and DDI Codebook representations.
- **GET `/api/examples/{type}`**: Returns a single example by type identifier.

Available types: `single_choice`, `multiple_choice`, `single_choice_other`, `multiple_choice_other`, `grid`, `integer`, `text`, `single_choice_long_list`, `multiple_choice_long_list`

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

PocketBase runs all migrations in `migrations/` in lexicographic order on startup.

- **Before first production deploy:** edit the initial migration files directly.
- **After first production deploy:** add a new file per change (e.g. `migrations/20260301000000_add_foo_field.go`) — never edit existing ones.

The Go test suite validates migrations on every run via `tests.NewTestApp`.

### Testing against production data

To test a migration against real data before deploying:

```bash
# Export the production DB from the Docker volume
docker run --rm \
  -v qwacback_pb_dataz:/data \
  -v $(pwd):/backup \
  alpine cp /data/data.db /backup/prod_backup.db

# Run migrations against the copy (errors appear in startup logs)
mkdir -p /tmp/test_pb_data && cp prod_backup.db /tmp/test_pb_data/data.db
go run main.go serve --dir=/tmp/test_pb_data
```

Before deploying a migration that drops or renames columns, back up first:

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
