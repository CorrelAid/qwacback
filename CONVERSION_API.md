# DDI ↔ XLSForm Conversion API

This document describes the public API endpoints for converting between DDI Codebook XML format and XLSForm JSON format.

The XLSForm JSON mirrors the actual XLSForm spreadsheet structure with three sheets:
- **survey**: questions and groups (columns: type, name, label, hint, required, appearance, parameters)
- **choices**: answer options for select questions (columns: list_name, name, label)
- **settings**: form metadata (columns: form_title, form_id, version)

## Endpoints

### 1. Convert DDI to XLSForm

**Endpoint:** `POST /api/convert/ddi-to-xlsform`

**Access:** Public (no authentication required)

**Request:**
- Content-Type: `application/xml` or `text/xml`
- Body: DDI XML fragment — a `<var>`, `<varGrp>`, or `<dataDscr>` wrapper (for `select_multiple` / `multipleResp` groups containing both `<varGrp>` and `<var>` elements)

**Response:**
- Content-Type: `application/json`
- Body: XLSForm JSON with survey, choices, and settings sheets

**Example:**

```bash
curl -X POST http://localhost:8090/api/convert/ddi-to-xlsform \
  -H "Content-Type: application/xml" \
  --data '<var ID="V1" name="gender" intrvl="discrete">
    <concept>Gender</concept>
    <qstn responseDomainType="category">
      <qstnLit>What is your gender?</qstnLit>
    </qstn>
    <catgry>
      <catValu>1</catValu>
      <labl>Male</labl>
    </catgry>
    <catgry>
      <catValu>2</catValu>
      <labl>Female</labl>
    </catgry>
    <varFormat type="numeric" schema="other"/>
  </var>'
```

**Response:**
```json
{
  "survey": [
    {
      "type": "select_one gender",
      "name": "gender",
      "label": "What is your gender?"
    }
  ],
  "choices": [
    {
      "list_name": "gender",
      "name": "1",
      "label": "Male"
    },
    {
      "list_name": "gender",
      "name": "2",
      "label": "Female"
    }
  ],
  "settings": {}
}
```

### 2. Convert XLSForm to DDI

**Endpoint:** `POST /api/convert/xlsform-to-ddi`

**Access:** Public (no authentication required)

**Request:**
- Content-Type: `application/json`
- Body: XLSForm JSON with survey, choices, and settings sheets

**Response:**
- Content-Type: `application/xml`
- Body: DDI XML — a single `<var>` for simple questions, or a `<dataDscr>` wrapper containing `<varGrp type="multipleResp">` + binary `<var>` elements for `select_multiple` questions

**Example:**

```bash
curl -X POST http://localhost:8090/api/convert/xlsform-to-ddi \
  -H "Content-Type: application/json" \
  --data '{
    "survey": [
      {"type": "select_one gender", "name": "gender", "label": "What is your gender?"}
    ],
    "choices": [
      {"list_name": "gender", "name": "1", "label": "Male"},
      {"list_name": "gender", "name": "2", "label": "Female"}
    ],
    "settings": {}
  }'
```

**Response:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<var ID="V_gender" name="gender" intrvl="discrete">
  <concept>What is your gender?</concept>
  <qstn responseDomainType="category">
    <qstnLit>What is your gender?</qstnLit>
  </qstn>
  <catgry>
    <catValu>1</catValu>
    <labl>Male</labl>
  </catgry>
  <catgry>
    <catValu>2</catValu>
    <labl>Female</labl>
  </catgry>
  <varFormat type="numeric" schema="other"></varFormat>
</var>
```

## XLSForm JSON Structure

The JSON format mirrors the three sheets of an XLSForm spreadsheet:

### Survey Sheet

Each row in the survey sheet is an object with these columns:

| Column | Required | Description |
|--------|----------|-------------|
| `type` | Yes | Answer type. For select questions, includes list_name: `select_one <list_name>` |
| `name` | Yes | Variable identifier (snake_case recommended) |
| `label` | No | Question text shown to respondents |
| `hint` | No | Additional hint text (maps to DDI `preQTxt`) |
| `required` | No | `"yes"` if the question is mandatory |
| `appearance` | No | Display preference |
| `parameters` | No | Key-value pairs, e.g. `"guidance_hint=Show card"` |

Groups use `begin_group`/`end_group` rows:
```json
{
  "survey": [
    {"type": "begin_group", "name": "demographics", "label": "Demographics"},
    {"type": "integer", "name": "age", "label": "What is your age?"},
    {"type": "end_group", "name": ""}
  ]
}
```

### Choices Sheet

Each row in the choices sheet is an object with these columns:

| Column | Required | Description |
|--------|----------|-------------|
| `list_name` | Yes | References the list in the survey type column |
| `name` | Yes | Choice value (e.g. "1", "2") |
| `label` | Yes | Choice display text |

### Settings Sheet

Single object with optional form metadata:

| Column | Required | Description |
|--------|----------|-------------|
| `form_title` | No | Form title |
| `form_id` | No | Form identifier |
| `version` | No | Form version |

## Wrapping Single Questions in DDI Codebook

### Single Variable (Standalone Question)

A single question in DDI is represented by a `<var>` element. Here's the structure:

```xml
<var ID="V1" name="variable_name" intrvl="discrete">
  <concept>Variable concept/label</concept>
  <qstn responseDomainType="category">
    <preQTxt>Optional introductory text</preQTxt>
    <qstnLit>The actual question text</qstnLit>
    <ivuInstr>Optional interviewer instructions</ivuInstr>
  </qstn>
  <catgry>
    <catValu>1</catValu>
    <labl>Option 1</labl>
  </catgry>
  <catgry>
    <catValu>2</catValu>
    <labl>Option 2</labl>
  </catgry>
  <varFormat type="numeric" schema="other"/>
</var>
```

### Variable Group (Matrix/Grid Questions)

When questions are part of a group (like a matrix or multiple response set), use a `<varGrp>` element:

```xml
<varGrp ID="VG1" name="satisfaction_group" type="grid" var="V1 V2 V3">
  <concept>Group concept/label</concept>
  <txt>Optional introductory text for the group</txt>
</varGrp>
```

The `var` attribute contains space-separated IDs of the variables that belong to this group.

### Complete DDI Codebook Structure

To create a complete DDI codebook with single questions, wrap them in the full structure:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <stdyDscr>
    <citation>
      <titlStmt>
        <titl>Study Title</titl>
        <IDNo>study-id-123</IDNo>
      </titlStmt>
    </citation>
    <stdyInfo>
      <abstract>Study abstract</abstract>
      <sumDscr>
        <timePrd>2024</timePrd>
        <nation>Country</nation>
        <anlyUnit>Analysis unit</anlyUnit>
        <universe>Study universe</universe>
        <dataKind>Survey Data</dataKind>
      </sumDscr>
    </stdyInfo>
  </stdyDscr>
  <dataDscr>
    <!-- Variable groups go here (if any) -->
    <varGrp ID="VG1" name="group1" type="grid" var="V1 V2">
      <concept>Group concept</concept>
      <txt>Group description</txt>
    </varGrp>

    <!-- Individual variables go here -->
    <var ID="V1" name="question1" intrvl="discrete">
      <concept>Question 1</concept>
      <qstn responseDomainType="category">
        <qstnLit>What is your answer?</qstnLit>
      </qstn>
      <catgry>
        <catValu>1</catValu>
        <labl>Yes</labl>
      </catgry>
      <catgry>
        <catValu>2</catValu>
        <labl>No</labl>
      </catgry>
      <varFormat type="numeric" schema="other"/>
    </var>

    <var ID="V2" name="question2" intrvl="discrete">
      <!-- ... -->
    </var>
  </dataDscr>
</codeBook>
```

## Supported Answer Types

### DDI to XLSForm Mapping

| DDI responseDomainType | XLSForm type | Notes |
|------------------------|--------------|-------|
| `numeric` | `integer` | Numeric input |
| `text` | `text` | Text input |
| `category` | `select_one <name>` | Single choice (list_name = variable name) |
| `multiple` | `select_multiple <name>` | Multiple choice (list_name = variable name) |

### XLSForm to DDI Mapping

| XLSForm type | DDI output | DDI responseDomainType | intrvl | varFormat.type |
|--------------|-----------|------------------------|--------|----------------|
| `integer`, `decimal`, `range` | `<var>` | `numeric` | `discrete` | `numeric` |
| `text`, `note` | `<var>` | `text` | `contin` | `character` |
| `select_one`, `matrix` | `<var>` + `<catgry>` | `category` | `discrete` | `numeric` |
| `select_multiple` | `<varGrp type="multipleResp">` + binary `<var>` per choice | `multiple` | `discrete` | `numeric` |

**Note on `select_multiple`:** Per DDI Codebook conventions, checkboxes are represented as a `<varGrp type="multipleResp">` with one binary `<var>` per choice option. Each binary variable has categories `0="Not mentioned"` and `1="Mentioned"`. The output is wrapped in a `<dataDscr>` element.

## Error Handling

Both endpoints return appropriate HTTP status codes:

- **200 OK**: Successful conversion
- **400 Bad Request**: Invalid input format or conversion error
- **413 Payload Too Large**: Request body exceeds 50MB limit

Error responses include a descriptive message:

```json
{
  "code": 400,
  "message": "Failed to convert DDI to XLSForm",
  "data": {
    "error": "input XML is neither a <var> nor a <varGrp> element"
  }
}
```

## Notes

- The conversion preserves the core question structure but may not retain all DDI metadata
- Generated DDI IDs follow the pattern `V_<name>` for variables and `VG_<name>` for groups
- XLSForm `hint` field maps to DDI `preQTxt` (pre-question text)
- Missing value categories (DDI `missing="Y"`) are excluded from XLSForm choices
- These endpoints are stateless and do not persist data to the database
