# DDI ↔ XLSForm Conversion API

This document describes the endpoints that convert between DDI Codebook 2.5 XML and XLSForm JSON.

The XLSForm JSON mirrors the XLSForm spreadsheet: a **survey**, a **choices** and a **settings** sheet, each a list of rows keyed by column name (`type`, `name`, `label`, `hint`, `relevant`, `label::English (en)`, …).

**Implementation.** Both directions are `@correlaid/formtransform` (release pinned in `.registry-version`), run by the `ddi-emitter` Node sidecar (`ddi-emitter/`) and reached over HTTP from `internal/converter/ddi_client.go`:

- XLSForm → DDI: `xlsformToDdi`
- DDI → XLSForm: `ddiToXlsform` (since formtransform v0.7.0; qwacback's own Go converter is gone, #37)

A codebook formtransform writes carries the whole form: standard DDI wherever DDI has an element, typed `<notes type="cdl:…">` for the rest (skip logic, constraints, appearance, list names, settings, …). `ddiToXlsform` turns such a codebook back into the same form. See formtransform's [`ddi2xlsform` README](https://github.com/CorrelAid/formtransform/blob/main/src/pipelines/ddi2xlsform/README.md) for the full mapping.

## Endpoints

### 1. Convert DDI to XLSForm

**Endpoint:** `POST /api/convert/ddi-to-xlsform`

**Access:** Public (no authentication required)

**Request:**
- Content-Type: `application/xml` or `text/xml`
- Body: a whole `<codeBook>`, a `<dataDscr>`, or bare `<var>` / `<varGrp>` elements

**Response:**
- Content-Type: `application/json`
- Body: the three sheets, plus `warnings`:

```json
{
  "survey": [{"type": "select_one gender", "name": "gender", "label": "What is your gender?"}],
  "choices": [
    {"list_name": "gender", "name": "1", "label": "Male"},
    {"list_name": "gender", "name": "2", "label": "Female"}
  ],
  "settings": [],
  "warnings": [
    {"code": "ddi-field-missing", "message": "Not a CDL codebook: no skip logic (cdl:relevant); the form gets none (formtransform#155)"}
  ]
}
```

`settings` is a list of rows, like the other sheets (empty, or one row). DDI is never refused if it can be read: DDI that formtransform didn't write converts as far as its standard elements go, with one `ddi-field-missing` warning for each field only a CDL codebook carries. Without `qstn/@seqNo` or any `cdl:` note, identical category sets come back as one list.

**Example:**

```bash
curl -X POST http://localhost:8090/api/convert/ddi-to-xlsform \
  -H "Content-Type: application/xml" \
  --data '<var ID="V1" name="gender" intrvl="discrete">
    <qstn responseDomainType="category">
      <qstnLit>What is your gender?</qstnLit>
    </qstn>
    <catgry><catValu>1</catValu><labl>Male</labl></catgry>
    <catgry><catValu>2</catValu><labl>Female</labl></catgry>
    <concept>Gender</concept>
    <varFormat type="numeric" schema="other"/>
  </var>'
```

### 2. Convert XLSForm to DDI

**Endpoint:** `POST /api/convert/xlsform-to-ddi`

**Access:** Public (no authentication required)

**Request:**
- Content-Type: `application/json`
- Body: `{"survey": [...], "choices": [...], "settings": {...}}`. `settings` may be an object or a one-row list. The sheets are forwarded to formtransform unchanged, so every column it reads reaches it.

**Response:**
- Content-Type: `application/xml`
- Body: the children of formtransform's `<dataDscr>`: a single `<var>` or `<varGrp>` bare, anything else wrapped in `<dataDscr>`

**Response shape note.** The endpoint returns a DDI fragment, not the full `<codeBook>` formtransform writes. The elements inside `<dataDscr>` are passed through as formtransform emits them, in its order; only the DDI namespace declaration and the `files` attributes (which point at the dropped `<fileDscr>`) are removed. What formtransform puts on `<stdyDscr>` (settings, languages, note rows without a question) is not part of the fragment.

**Example:**

```bash
curl -X POST http://localhost:8090/api/convert/xlsform-to-ddi \
  -H "Content-Type: application/json" \
  --data '{
    "survey": [
      {"type": "integer", "name": "age", "label": "How old are you?", "hint": "In years", "constraint": ". < 120"}
    ],
    "choices": [],
    "settings": {}
  }'
```

**Response:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<var ID="V_age" name="age" intrvl="contin" dcml="0">
  <qstn responseDomainType="numeric" seqNo="1">
    <qstnLit>How old are you?</qstnLit>
    <postQTxt>In years</postQTxt>
  </qstn>
  <valrng>
    <range maxExclusive="120"></range>
  </valrng>
  <concept>How old are you?</concept>
  <varFormat type="numeric" schema="other"></varFormat>
  <notes type="cdl:constraint" subject="xlsform-xpath">. &lt; 120</notes>
</var>
```

### 3. Stored studies and questions

- `GET /api/studies/{id}/export`: the study's codebook, exactly as imported.
- `GET /api/studies/{id}/xlsform`: `ddiToXlsform` of that codebook.
- `GET /api/questions/{id}/xml`: the question's elements, cut from the stored codebook unchanged. A group question is a `<dataDscr>` with its `<varGrp>`, the groups nested in it (`@varGrp`) and every `<var>` they refer to. A standalone question is its bare `<var>`. A grid row is a `<dataDscr>` with the grid's `<varGrp>`, narrowed to that one row, and its `<var>`. The codebook's `xml:lang` goes on the root.
- `GET /api/questions/{id}/xlsform`: `ddiToXlsform` of that fragment. A question keeps its skip logic, which may refer to questions outside the fragment.

Exports read the stored codebook, not the database fields, so everything the codebook carries comes back out (#37).

## How form fields map to DDI

formtransform's convention (v0.7.0), the part qwacback reads into its records:

| XLSForm | DDI | qwacback field (variables) |
|---|---|---|
| label | `qstn/qstnLit` | `question` |
| hint | `qstn/postQTxt` | `hint` |
| guidance_hint | `qstn/ivuInstr` | `ivu_instructions` |
| a note before the question; a grid's or select_multiple's shared text | `qstn/preQTxt` | `prequestion_text` |
| relevant | `<notes type="cdl:relevant" subject="xlsform-xpath">`, and as prose in `<universe clusion="I">` | `universe` (the prose) |
| group | `<varGrp type="section">`, nested via `@varGrp`; grids `type="grid"` | grid and multipleResp groups become `variable_groups`; sections stay in the codebook only |
| integer / decimal / range / date / time | `var/@dcml="0"` / numeric without it / `valrng/range` without `cdl:constraint` / `varFormat/@category` | `answer_type` (`integer`, `decimal`, `range`, `date`, `time`) |

**Changed in formtransform v0.7.0 (#39):** the hint moved from `preQTxt` to `postQTxt`. `preQTxt` now holds only a lead-in note, or a grid's or select_multiple's shared text.

## Multilingual forms

XLSForm → DDI takes multilingual forms as XLSForm writes them. Use `label::<Language> (<code>)` / `hint::<Language> (<code>)` columns on survey and choice rows, and name the base language in `settings.default_language`:

```json
{
  "survey": [{"type": "integer", "name": "alter", "label::Deutsch (de)": "Alter?", "label::English (en)": "Age?"}],
  "settings": {"default_language": "Deutsch (de)"}
}
```

The response has each text element in the base language, untagged and first, followed by one `xml:lang` sibling per other language (formtransform#135). The base language is set as `xml:lang` on the fragment's root element (`<var>`, `<varGrp>` or `<dataDscr>`), since the `<codeBook>` that formtransform declares it on isn't part of the response:

```xml
<var xml:lang="de" ID="V_alter" name="alter" intrvl="contin" dcml="0">
  <qstn responseDomainType="numeric" seqNo="1">
    <qstnLit>Alter?</qstnLit>
    <qstnLit xml:lang="en">Age?</qstnLit>
  </qstn>
  …
```

On import, qwacback stores the base-language text (the untagged element, or the one matching `codeBook/@xml:lang`) in the usual fields. It stores the other languages in `translations` on variables and variable groups, and the base language as `language` on the study. DDI → XLSForm writes one `label::<lang>` column per language, as formtransform does.

## Error Handling

- **200 OK**: Successful conversion.
- **400 Bad Request**: the input is at fault, and the message says why.
  - XLSForm → DDI: formtransform's reason, which names the question: types outside the supported subset (`rank`, `geopoint`, …), selects whose list has no choices, or a form without any answerable question (notes produce no DDI variables).
  - DDI → XLSForm: XML that isn't well-formed, or holds neither a `<codeBook>` nor a `<var>`.
- **413 Payload Too Large**: the request body exceeds the limit.
- **503 Service Unavailable**: the `ddi-emitter` sidecar is unreachable or failed. Retry later; the input is not at fault.

```json
{
  "status": 400,
  "message": "Failed to convert DDI to XLSForm: The input holds no <codeBook> and no <var>.",
  "data": {}
}
```

## Library version pin

The converter is `@correlaid/formtransform`. The release tag is pinned in `.registry-version` at the repo root and consumed by:

- `ddi-emitter/package.json` and `ddi-emitter/package-lock.json` (download the prebuilt tarball — no build step, no `git` in the image)
- `docker-compose.yml` pulls `ghcr.io/correlaid/schematron-worker:<tag>`

Bump `.registry-version` and update every consumer in the same PR (`cd ddi-emitter && npm install <new tarball URL>` refreshes both npm files). `scripts/check-registry-version.sh` checks that they agree and runs in the release workflow. Never pin to a branch or `latest`.
