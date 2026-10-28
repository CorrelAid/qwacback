# qwacback - Questions Worth Asking Continuously

[![AI-Assisted](https://img.shields.io/badge/AI--assisted-Claude%20Code-blueviolet?logo=anthropic&logoColor=white)](./AI_DISCLOSURE.md)

A question bank for civil society surveys. Import DDI-Codebook XML, browse and search questions, export as DDI or XLSForm.

## How It Works

The core concept is a **question** — one thing you ask a respondent. Under the hood, questions are stored as [DDI Codebook 2.5](https://ddialliance.org/Specification/DDI-Codebook/2.5/) variables and variable groups (see [DDI_MARKUP_GUIDE.md](DDI_MARKUP_GUIDE.md)), but the API presents them as questions:

| Question type | DDI storage | Example |
|---|---|---|
| Simple (integer, text, single choice) | 1 `<var>` | "How old are you?" |
| Multiple choice | `<varGrp type="multipleResp">` + binary `<var>` per option | "Which devices do you own?" |
| Grid / Likert | `<varGrp type="grid">` + `<var>` per item | "Rate your trust in: Parliament, Police, ..."; or a table of number fields ("How many employees: in total, of whom male, female, ...") |
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
GET /api/studies/{id}/export  →  The codebook as imported
GET /api/studies/{id}/xlsform →  Convert to XLSForm JSON
```

### Architecture

Three Docker services:

- **qwacback** (Go/PocketBase) — API, database, embedded NATS server
- **ddi-emitter** (Node) — XLSForm → DDI proxy over `@correlaid/formtransform`
- **schematron-worker** (Java/Saxon HE) — XSD + Schematron validation over NATS

If `NATS_PORT` is not set, qwacback runs without validation (import-only mode).

## API

### Questions

- **GET `/api/questions`** — List all questions across all studies.
- **GET `/api/questions/{id}`** — Get a single question with full detail: embedded study, group, and variable data (categories, prequestion text, hint, interviewer instructions, skip logic as prose in `universe`, etc.). No additional API calls needed for a detail view.

Each study carries its terms of use, so a client can show them with every reused question. `/api/questions/{id}` embeds them in `study` as `license`, `copyright`, `citation` and `source_uri`; the `studies` records have them as `license`, `copyright`, `citation` and `holdings_uri`. They are read from the codebook (see DDI_MARKUP_GUIDE.md §6) and are empty when it states none.

A study's short name for lists and tags, e.g. `ATTARI-12`, is `short_title` on the `studies` records and `study.short_title` in `/api/questions/{id}`, from the codebook's `titlStmt/altTitl`; its subtitle (`titlStmt/subTitl`) is `subtitle` and `study.subtitle` (DDI_MARKUP_GUIDE.md §5). `/api/search/studies` returns `short_title` too. Both are empty when the codebook has none; then show `title`.
- **GET `/api/questions/{id}/xml`** — DDI-XML fragment for a single question, cut unchanged from the stored codebook.
- **GET `/api/questions/{id}/xlsform`** — XLSForm JSON for a single question.
- **GET `/api/studies/{id}/questions`** — List all questions for a single study.
- **GET `/api/search/questions?q=<terms>`** — Search questions by question text, concept, name, and answer type. Supports `&page=` and `&perPage=` (default 20, max 100).
  - `q` may hold several terms, separated by spaces or commas; a question matches if **any** term matches, and questions matching more terms rank first (then question text > concept > name > answer type).
  - Matching ignores case and umlaut spelling (`Qualitaet` = `Qualität`), finds word forms via German and English stemming (`Zufriedenheit` finds "zufrieden", `Weiterempfehlung` finds "weiterempfehlen"), finds terms of 5+ letters inside compounds (`Vertrauen` finds "Institutionenvertrauen"). Tags are searched too: `<concept>` elements after the first one on a `<var>`/`<varGrp>`, e.g. `<concept xml:lang="de">Vertrauen</concept>` on an English item, make it findable in the other language without changing its wording.
  - `&study_id=<id>` (repeatable) limits the search to those studies; `&exclude_study=<id>` (repeatable) leaves studies out, e.g. the demographic standards. The MCP tool `search_questions` takes the same filters.

Questions of a multilingual study also carry `language` (the base language of `question_text`) and `translations` (the question text in the other languages, `{"en": "…"}`). `/api/questions/{id}` also returns the full `translations` of each variable and of the group. The search matches translated question text like the question text.

### Studies

- **GET `/api/search/studies?q=<term>`** — Search studies by title, short title, keywords, and abstract. Optional `&topic=<classification>` filter. Supports pagination.

### Import & Validation

- **POST `/api/validate`** — Validate a DDI XML file (XSD + Schematron) without importing. Body: `multipart/form-data` with `file` field.
- **POST `/api/import`** — Validate and import a DDI XML file. Same body format. **Requires superuser auth.** Responds `{"valid": true, "imported": true, "replaced": false, "study_id": "…", "message": …}` on success; `imported: false` when the file is valid but couldn't be stored. Imports are kept across deploys; see [Data and deployment](#data-and-deployment).
  - A study's ID is derived from its title. If a study with the title exists, the import is refused with **409** and that `study_id`, unless `?replace=1` is given: then the study, its groups and its variables are replaced in one transaction, keeping their IDs (`"replaced": true`).
  - A study from `seed_data/` is never replaced through the API, not even with `?replace=1` (409): change its file instead.

### Export & Conversion

- **GET `/api/studies/{id}/export`** — Export study as validated DDI-XML download.
- **GET `/api/studies/{id}/xlsform`** — Export study as XLSForm JSON.
- **POST `/api/convert/ddi-to-xlsform`** — Convert a DDI XML fragment to XLSForm JSON.
- **POST `/api/convert/xlsform-to-ddi`** — Convert XLSForm JSON to DDI XML.

For conversion details, see [CONVERSION_API.md](CONVERSION_API.md).

### Reference

- **GET `/api/examples`** — Answer type examples (XLSForm + DDI pairs).
- **GET `/api/examples/{type}`** — Single example by type (`single_choice`, `multiple_choice`, `grid`, `integer`, `text`, etc.).
- **GET `/api/question-types`** — The registry's question-type catalogue of the pinned formtransform, keyed by qwacback's answer types: `registryType`, `label` per language (`en`, `de`), `kind`, `base`, `aliases` and `presentation` (`choice`: `one`/`multiple`, `withOther`, `longList`, `grid`, `appearance`). A `grid` has no `choice`: its rows are choice rows (answer type `grid`) or number rows (`integer`, `decimal`), and a row's `grid_note` is a sub-heading printed before it. `withOther`, `longList` and `appearance` are the registry's. `?registry=1` returns every registry type keyed by registry slug instead. Clients get the catalogue here instead of importing `@correlaid/formtransform`.
- **GET `/api/docs/markup-guide`** — DDI encoding conventions.

### PocketBase built-in API

The `studies` collection is publicly readable via PocketBase's standard REST API (`GET /api/collections/studies/records`). Individual `variables` and `variable_groups` records are also publicly readable by ID (`GET /api/collections/{name}/records/{id}`), but bulk listing those collections requires authentication. Write access to all collections is admin-only.

Prefer the custom `/api/questions/*` endpoints over direct collection access — they return assembled, frontend-ready data.

---

## Getting Started

### Docker Compose (recommended)

```bash
docker compose up -d --build
```

Access the PocketBase Dashboard at `http://localhost:8090/_/`.

### Data and deployment

The database lives in `/app/pb_data` (SQLite, plus PocketBase's settings, logs and local backups). Mount a **persistent volume** on it: `docker-compose.yml` does (named volume `pb_data`), and production (Coolify) needs one too. Without a volume every start begins with an empty database, which still works: it is re-seeded from `seed_data/`, but uploads, dashboard edits and settings are lost.

Studies have a `source`:

- **`seed`**: a file in `seed_data/`. **`seed_data/` is the source of truth for these.** On every start, after the migrations, the seed sync (`internal/seed`) brings them in line with the files:
  - a new file is imported;
  - a changed file (its `seed_hash`, the importer revision and the file's sha256, differs) replaces the study, its groups and its variables in one transaction;
  - an unchanged file is skipped;
  - a seed study whose file is gone is deleted, with its groups and variables.

  New and changed files are validated first when the worker is connected (`NATS_PORT`); a rejected file fails the start before anything is written. Edits to a seed study in the dashboard last until its file changes; to change a study, a question or its search tags, edit the DDI file and merge. `QWACBACK_SKIP_SEED=1` skips the sync.
- **`upload`**: imported through `/api/import`. The seed sync never touches it. If a seed file has the title of an upload, the sync skips that file and logs a warning instead of overwriting the upload; rename or delete the upload to let the seed in.

To add a published questionnaire, use `tools/survey-import/` (see its `PLAN.md`): transcribe the source literally to XLSForm JSON, build the codebook with the pinned formtransform, validate it with the pinned worker, and record source and license in `SOURCE.md`.

Record IDs are derived from the study title and the variable/group name, so they stay the same across deploys and replacements as long as those don't change.

#### Migrations

The database is persistent, so migrations run on existing data, not only on a fresh database:

- **Date a new migration after all existing ones.** Older migrations (`codebook`, `license`, …) are dated before `20260219130000`, the former one-time seed; that ordering only worked because every start began empty.
- **A new field needs a backfill for existing rows**, e.g. re-derived from the stored `studies.codebook`. For seed studies, raising `importer.Revision` instead makes the next start re-import them all; uploads still need the backfill.

#### Backups

PocketBase makes scheduled backups (a zip of `pb_data` without its `backups/` folder). Set them either in the dashboard (Settings → Backups) or through the environment: `PB_BACKUP_CRON` (e.g. `0 3 * * *`, daily at 03:00 UTC), `PB_BACKUP_KEEP` (default 7) and, to store them in S3 instead of `pb_data/backups`, `PB_BACKUP_S3_*` (see [Environment Variables](#environment-variables)). When `PB_BACKUP_CRON` is set, the environment owns the backup settings and overwrites dashboard changes on every start. Backups kept only in `pb_data/backups` sit on the same volume as the database; use S3 to survive losing the volume.

**Restore:** in the dashboard, Settings → Backups → a backup → Restore; PocketBase replaces `pb_data` and restarts. Or by hand: stop the container, unzip the backup into the volume (replacing `data.db`, `auxiliary.db` and `storage/`), start it. Either way the seed sync then brings the seed studies in line with the deployed `seed_data/`.

#### Coolify

Production runs on Coolify, configured outside this repo. On the qwacback resource:

1. **Persistent Storage → add a volume** with destination path `/app/pb_data` (if the resource is deployed from `docker-compose.yml`, Coolify takes the `pb_data` volume from it; check that it is listed).
2. **Environment variables** for backups, e.g. `PB_BACKUP_CRON=0 3 * * *`, `PB_BACKUP_KEEP=14` and the `PB_BACKUP_S3_*` of the backup bucket (secret marked as such), or set them in the dashboard after the first deploy.
3. Redeploy. The first start with the empty volume seeds every study from `seed_data/`; from then on uploads and settings persist.

### Published images

[.github/workflows/release.yml](.github/workflows/release.yml) builds qwacback and its sidecar and publishes them to GitHub Container Registry:

- `ghcr.io/correlaid/qwacback` — the Go/PocketBase API
- `ghcr.io/correlaid/qwacback-ddi-emitter` — the DDI ↔ XLSForm sidecar. qwacback needs it for `/api/convert/*`, every `/xlsform` endpoint, `/api/examples` and `/api/question-types`; point `DDI_EMITTER_URL` at it. Deploy it with the same tag as qwacback.

The validation worker (`ghcr.io/correlaid/schematron-worker`) is published by [CorrelAid/formtransform](https://github.com/CorrelAid/formtransform) and pulled into this stack via `docker-compose.yml`. The same is true for the `@correlaid/formtransform` library used by `ddi-emitter`. Both are version-pinned together via `.registry-version`; `scripts/check-registry-version.sh` (run by the release workflow) fails if they disagree.

**Every push to `main`** updates the `latest` tag (and a `main` tag). **Pushing a `v*` git tag** additionally publishes semver-pinned tags:

| Git action | Image tags produced |
|---|---|
| Push to `main` | `latest`, `main` |
| Push tag `v1.2.3` | `1.2.3`, `1.2`, `1` |

To cut a versioned release:

```bash
git tag v1.2.3
git push origin v1.2.3
```

The tag must start with `v` and point to a commit that includes `.github/workflows/release.yml`. Pin to `1.2.3` in production; use `latest` only for dev/staging.

Default credentials (see `docker-compose.yml`):
- **Admin:** `admin@example.com` / `yourpassword123`
- **User:** `user@example.com` / `userpassword123`

### Local Development

**Without validation** (PocketBase only — imports work, validation skipped):

```bash
go run main.go serve
```

**With validation** (Docker required, since the worker is now a published image):

```bash
# Start qwacback with embedded NATS (NATS_TOKEN is required when NATS_PORT is set)
NATS_PORT=4222 NATS_TOKEN=localdev go run main.go serve &

# Pull and start the published validation worker
docker run -d --rm --name schematron-worker \
  --network host \
  -e NATS_URL=nats://localhost:4222 \
  -e NATS_TOKEN=localdev \
  ghcr.io/correlaid/schematron-worker:$(cat .registry-version)
```

## Project Structure

```
internal/
  converter/    DDI ↔ XLSForm conversion, both ways via ddi-emitter
  examples/     Static answer type examples (XLSForm + DDI)
  ddixml/       Copies DDI elements as XML tokens
  exporter/     Study and question DDI, from the codebook as imported
  importer/     XML parsing → PocketBase records
  routes/       API endpoints, question assembly, search
  schematron/   Go NATS client for validation worker
  seed/         Seed sync: keeps the seed studies in line with seed_data/ on every start
  backup/       Scheduled backups from PB_BACKUP_* environment variables
migrations/     Schema setup, settings, user init, backfills
ddi-emitter/    Node sidecar: DDI ↔ XLSForm via @correlaid/formtransform
seed_data/      The reference studies: DDI-XML files synced into the database on every start
tools/survey-import/  Turning published questionnaires into seed codebooks: build and validate tools, one folder per source
.registry-version   Pinned formtransform release tag (drives ddi-emitter and the worker image)
```

## Development & Testing

### Go Tests

Tests run during Docker build (`go test` in Dockerfile) and locally. No NATS or Java needed.

```bash
go test ./internal/...
```

### Integration Test

```bash
docker compose up -d --build

# Validate only — no auth required
curl -X POST http://localhost:8090/api/validate \
  -F "file=@seed_data/prove_it.xml"

# Import — requires superuser auth
TOKEN=$(curl -s -X POST http://localhost:8090/api/collections/superusers/auth-with-password \
  -H 'Content-Type: application/json' \
  -d '{"identity":"admin@example.com","password":"yourpassword123"}' | jq -r '.token')

curl -X POST http://localhost:8090/api/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@seed_data/prove_it.xml"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PB_ADMIN_EMAIL` | Initial superuser email | `admin@example.com` |
| `PB_ADMIN_PASSWORD` | Initial superuser password | `yourpassword123` |
| `PB_USER_EMAIL` | Initial regular user email | `user@example.com` |
| `PB_USER_PASSWORD` | Initial regular user password | `userpassword123` |
| `PB_ENCRYPTION_KEY` | Exactly 32 characters; encrypts the stored app settings (e.g. backup S3 credentials). Unset: settings are stored in plain text. Set it before the first start with a persistent volume and keep it: settings saved with one key can't be read with another | (optional) |
| `GOMEMLIMIT` | Soft memory limit for Go GC | `512MiB` |
| `NATS_PORT` | Port for embedded NATS server | (optional — no validation without it) |
| `NATS_TOKEN` | Auth token for embedded NATS server | (required when `NATS_PORT` is set) |
| `QWACBACK_SKIP_SEED` | Set to `1` to skip the seed sync of `seed_data/*.xml` on start (seed studies stay as they are) | (unset) |
| `PB_BACKUP_CRON` | Schedule for PocketBase backups, e.g. `0 3 * * *`; when set, the environment owns the backup settings | (unset — dashboard settings) |
| `PB_BACKUP_KEEP` | Scheduled backups to keep | `7` |
| `PB_BACKUP_S3_BUCKET` | Store backups in this S3 bucket instead of `pb_data/backups` | (unset — local) |
| `PB_BACKUP_S3_REGION`, `PB_BACKUP_S3_ENDPOINT`, `PB_BACKUP_S3_ACCESS_KEY`, `PB_BACKUP_S3_SECRET` | S3 connection for the backups | (unset) |
| `PB_BACKUP_S3_FORCE_PATH_STYLE` | `true` for path-style S3 addressing (e.g. MinIO) | `false` |

## MCP Server

qwacback exposes a [Model Context Protocol](https://modelcontextprotocol.io/) server at `/mcp` using Streamable HTTP transport. This lets AI assistants (e.g. Claude) search and browse the question bank directly.

### Available tools

| Tool | Description |
|---|---|
| `search_questions` | Search questions by text, concept, name, or answer type |
| `search_studies` | Search studies by title, short title, keywords, or abstract (optional `topic` filter) |
| `get_question` | Get a single question by ID in full: study, group and every variable with item text, hint, categories and grid note |
| `list_questions` | List all questions, optionally filtered by study |

All tools are read-only.

### Authentication

`GET /mcp` (tool discovery) is public. `POST /mcp` and `DELETE /mcp` (tool calls and session teardown) require a superuser token in the `Authorization: Bearer <token>` header.

### Client configuration

Add to your MCP client config (e.g. Claude Desktop, Claude Code):

```json
{
  "mcpServers": {
    "qwacback": {
      "type": "streamable-http",
      "url": "http://localhost:8090/mcp",
      "headers": {
        "Authorization": "Bearer <superuser-token>"
      }
    }
  }
}
```

Obtain a token via:

```bash
curl -s -X POST http://localhost:8090/api/collections/superusers/auth-with-password \
  -H 'Content-Type: application/json' \
  -d '{"identity":"admin@example.com","password":"yourpassword123"}' | jq -r '.token'
```

## Resources

- [DDI Codebook 2.5 Specification](https://ddialliance.org/Specification/DDI-Codebook/2.5/)
- [DDI Markup Guide](DDI_MARKUP_GUIDE.md) — project-specific encoding conventions
- [PocketBase Documentation](https://pocketbase.io/docs/)
