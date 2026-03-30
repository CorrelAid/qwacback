# qwacback - Questions Worth Asking Continuously

[![AI-Assisted](https://img.shields.io/badge/AI--assisted-Claude%20Code-blueviolet?logo=anthropic&logoColor=white)](./AI_DISCLOSURE.md)

A question bank for civil society surveys. Import DDI-Codebook XML, browse and search questions, export as DDI or XLSForm.

## How It Works

The core concept is a **question** — one thing you ask a respondent. Under the hood, questions are stored as [DDI Codebook 2.5](https://ddialliance.org/Specification/DDI-Codebook/2.5/) variables and variable groups (see [DDI_MARKUP_GUIDE.md](DDI_MARKUP_GUIDE.md)), but the API presents them as questions:

| Question type | DDI storage | Example |
|---|---|---|
| Simple (integer, text, single choice) | 1 `<var>` | "How old are you?" |
| Multiple choice | `<varGrp type="multipleResp">` + binary `<var>` per option | "Which devices do you own?" |
| Grid / Likert | `<varGrp type="grid">` + `<var>` per item | "Rate your trust in: Parliament, Police, ..." |
| Semi-open (with "other") | `<varGrp type="other">` + member vars + `_other` text var | "What is your gender?" (with free text option) |

Questions belong to **studies** — a study is a survey or questionnaire with metadata (title, abstract, keywords, time period, etc.).

### Data flow

```
DDI-XML file
  │
  ├─► POST /api/validate     →  XSD + Schematron validation only
  │
  └─► POST /api/import       →  Validate, parse, store in DB
                                    studies / variable_groups / variables
  │
  ▼
GET /api/questions            →  Browse all questions (assembled from vars + groups)
GET /api/search/questions     →  Search by question text, concept, name
GET /api/studies/{id}/export  →  Re-export as validated DDI-XML
GET /api/studies/{id}/xlsform →  Convert to XLSForm JSON
```

### Architecture

Two Docker services:

- **qwacback** (Go/PocketBase) — API, database, embedded NATS server
- **schematron-worker** (Java/Saxon HE) — XSD + Schematron validation over NATS

If `NATS_PORT` is not set, qwacback runs without validation (import-only mode).

## API

### Questions

- **GET `/api/questions`** — List all questions across all studies.
- **GET `/api/questions/{id}`** — Get a single question.
- **GET `/api/questions/{id}/xml`** — DDI-XML fragment for a single question.
- **GET `/api/questions/{id}/xlsform`** — XLSForm JSON for a single question.
- **GET `/api/studies/{id}/questions`** — List all questions for a single study.
- **GET `/api/search/questions?q=<term>`** — Search questions by question text, concept, name, and answer type. Ranked by relevance. Supports `&page=` and `&perPage=` (default 20, max 100).

### Studies

- **GET `/api/search/studies?q=<term>`** — Search studies by title, keywords, and abstract. Optional `&topic=<classification>` filter. Supports pagination.

### Import & Validation

- **POST `/api/validate`** — Validate a DDI XML file (XSD + Schematron) without importing. Body: `multipart/form-data` with `file` field.
- **POST `/api/import`** — Validate and import a DDI XML file. Same body format.

### Export & Conversion

- **GET `/api/studies/{id}/export`** — Export study as validated DDI-XML download.
- **GET `/api/studies/{id}/xlsform`** — Export study as XLSForm JSON.
- **POST `/api/convert/ddi-to-xlsform`** — Convert a DDI XML fragment to XLSForm JSON.
- **POST `/api/convert/xlsform-to-ddi`** — Convert XLSForm JSON to DDI XML.

For conversion details, see [CONVERSION_API.md](CONVERSION_API.md).

### Reference

- **GET `/api/examples`** — Answer type examples (XLSForm + DDI pairs).
- **GET `/api/examples/{type}`** — Single example by type (`single_choice`, `multiple_choice`, `grid`, `integer`, `text`, etc.).
- **GET `/api/docs/markup-guide`** — DDI encoding conventions.
- **GET `/api/schemas/schematron`** — Schematron validation rules.
- **GET `/api/schemas/xsd`** — List available XSD files.

### PocketBase built-in API

The raw `studies`, `variable_groups`, and `variables` collections are publicly readable via PocketBase's standard REST API (`GET /api/collections/{name}/records`). Write access is admin-only.

---

## Getting Started

### Docker Compose (recommended)

```bash
docker compose up -d --build
```

Access the PocketBase Dashboard at `http://localhost:8090/_/`.

Default credentials (see `docker-compose.yml`):
- **Admin:** `admin@example.com` / `yourpassword123`
- **User:** `user@example.com` / `userpassword123`

### Local Development

**Without validation** (PocketBase only — imports work, validation skipped):

```bash
go run main.go serve
```

**With validation** (requires JDK 17+):

```bash
# Start qwacback with embedded NATS
NATS_PORT=4222 go run main.go serve &

# Build and start the validation worker
cd schematron-worker
gradle shadowJar
NATS_URL=nats://localhost:4222 java -jar build/libs/schematron-worker-1.0.0-all.jar
```

## Project Structure

```
internal/
  converter/    Bidirectional DDI ↔ XLSForm conversion
  examples/     Static answer type examples (XLSForm + DDI)
  exporter/     PocketBase records → DDI-XML
  importer/     XML parsing → PocketBase records
  routes/       API endpoints, question assembly, search
  schematron/   Go NATS client for validation worker
migrations/     Schema setup, settings, user init, seed data
xml/            DDI-Codebook 2.5 XSD schemas
schematron/     Custom Schematron rules (.sch)
seed_data/      Seed studies (DDI-XML files imported on first run)
schematron-worker/  Java validation microservice
```

## Development & Testing

### Go Tests

Tests run during Docker build (`go test` in Dockerfile) and locally. No NATS or Java needed.

```bash
go test ./internal/...
```

### Java Tests

```bash
# With Docker (no local JDK required)
docker run --rm -v "$(pwd)":/app -w /app/schematron-worker gradle:8.12-jdk17 gradle test --no-daemon

# With local Gradle + JDK 17
cd schematron-worker && gradle test
```

### Integration Test

```bash
docker compose up -d --build

TOKEN=$(curl -s -X POST http://localhost:8090/api/collections/users/auth-with-password \
  -H 'Content-Type: application/json' \
  -d '{"identity":"user@example.com","password":"userpassword123"}' | jq -r '.token')

# Validate only
curl -X POST http://localhost:8090/api/validate \
  -H "Authorization: $TOKEN" \
  -F "file=@seed_data/prove_it.xml"

# Validate and import
curl -X POST http://localhost:8090/api/import \
  -H "Authorization: $TOKEN" \
  -F "file=@seed_data/prove_it.xml"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PB_ADMIN_EMAIL` | Initial superuser email | `admin@example.com` |
| `PB_ADMIN_PASSWORD` | Initial superuser password | `yourpassword123` |
| `PB_USER_EMAIL` | Initial regular user email | `user@example.com` |
| `PB_USER_PASSWORD` | Initial regular user password | `userpassword123` |
| `PB_ENCRYPTION_KEY` | 32-char key for settings encryption | (optional) |
| `GOMEMLIMIT` | Soft memory limit for Go GC | `512MiB` |
| `NATS_PORT` | Port for embedded NATS server | (optional — no validation without it) |

## MCP Server

qwacback exposes a [Model Context Protocol](https://modelcontextprotocol.io/) server at `/mcp` using Streamable HTTP transport. This lets AI assistants (e.g. Claude) search and browse the question bank directly.

### Available tools

| Tool | Description |
|---|---|
| `search_questions` | Search questions by text, concept, name, or answer type |
| `search_studies` | Search studies by title, keywords, or abstract (optional `topic` filter) |
| `get_question` | Get a single question by ID |
| `list_questions` | List all questions, optionally filtered by study |

All tools are read-only.

### Client configuration

Add to your MCP client config (e.g. Claude Desktop, Claude Code):

```json
{
  "mcpServers": {
    "qwacback": {
      "type": "streamable-http",
      "url": "http://localhost:8090/mcp"
    }
  }
}
```

## Resources

- [DDI Codebook 2.5 Specification](https://ddialliance.org/Specification/DDI-Codebook/2.5/)
- [DDI Markup Guide](DDI_MARKUP_GUIDE.md) — project-specific encoding conventions
- [PocketBase Documentation](https://pocketbase.io/docs/)
