# Answer Type Mapping: DDI Codebook ↔ XLSForm

This guide explains how answer types are converted between DDI Codebook XML and XLSForm JSON formats.

## Overview

DDI (Data Documentation Initiative) and XLSForm use different terminology and structures to represent survey questions. This converter handles the translation between these formats while preserving the semantic meaning of each question.

The XLSForm JSON mirrors the actual spreadsheet format with three sheets:
- **survey**: rows of questions/groups (type, name, label, hint, required, appearance, parameters)
- **choices**: rows of answer options (list_name, name, label)
- **settings**: form metadata (form_title, form_id, version)

## Quick Reference Tables

### DDI → XLSForm

| DDI `responseDomainType` | DDI `intrvl` | XLSForm `type` | Use Case |
|--------------------------|--------------|----------------|----------|
| `numeric` | `discrete` | `integer` | Age, quantity, count |
| `text` | `contin` | `text` | Comments, descriptions |
| `category` | `discrete` | `select_one <name>` | Radio buttons, dropdown |
| `category` (in grid) | `discrete` | `matrix` | Grid/table questions |
| `multiple` | `discrete` | `select_multiple <name>` | Checkboxes |

### XLSForm → DDI

| XLSForm `type` | DDI `responseDomainType` | DDI `intrvl` | DDI `varFormat.type` | Use Case |
|----------------|--------------------------|--------------|----------------------|----------|
| `integer` | `numeric` | `discrete` | `numeric` | Whole numbers |
| `decimal` | `numeric` | `discrete` | `numeric` | Decimal numbers |
| `range` | `numeric` | `discrete` | `numeric` | Slider, range |
| `text` | `text` | `contin` | `character` | Open text |
| `note` | `text` | `contin` | `character` | Display text |
| `select_one` | `category` | `discrete` | `numeric` | Single choice |
| `matrix` | `category` | `discrete` | `numeric` | Grid question |
| `select_multiple` | `multiple` | `discrete` | `numeric` | Multiple choice (`<varGrp type="multipleResp">` + binary `<var>` per option) |

## Field Mapping Details

### Survey Sheet Fields

| XLSForm Column | DDI Element/Attribute | Notes |
|----------------|----------------------|-------|
| `type` | `responseDomainType` (attr) | Determines answer type; includes list_name for select types |
| `name` | `name` (attr) | Variable identifier |
| `label` | `qstnLit` | Main question text |
| `label` | `concept` | Also copied to concept |
| `hint` | `preQTxt` | Pre-question text/hint |
| `parameters` | `ivuInstr` | `guidance_hint=...` maps to interviewer instructions |
| `required` | *(not mapped)* | Validation rule |
| `appearance` | *(not mapped)* | Display preference |

### Choices Sheet Fields

| XLSForm Column | DDI Element | Notes |
|----------------|-------------|-------|
| `list_name` | *(implicit)* | Links choices to survey row's type (e.g. `select_one <list_name>`) |
| `name` | `catValu` | Category value |
| `label` | `labl` | Category label |

### Auto-Generated Fields (XLSForm → DDI)

When converting from XLSForm to DDI, these fields are automatically generated:

| DDI Field | Value | Notes |
|-----------|-------|-------|
| `ID` (attr) | `V_<name>` or `VG_<name>` | Unique identifier |
| `intrvl` (attr) | `discrete` or `contin` | Based on answer type |
| `varFormat.type` (attr) | `numeric` or `character` | Data storage type |
| `varFormat.schema` (attr) | `other` | Standard DDI value |

## Detailed Examples

### Example 1: Single Choice Question (Select One)

#### XLSForm (JSON Input)
```json
{
  "survey": [
    {
      "type": "select_one gender",
      "name": "gender",
      "label": "What is your gender?",
      "hint": "Please select one option"
    }
  ],
  "choices": [
    {"list_name": "gender", "name": "1", "label": "Male"},
    {"list_name": "gender", "name": "2", "label": "Female"},
    {"list_name": "gender", "name": "3", "label": "Non-binary"}
  ],
  "settings": {}
}
```

#### DDI (XML Output)
```xml
<?xml version="1.0" encoding="UTF-8"?>
<var ID="V_gender" name="gender" intrvl="discrete">
  <concept>What is your gender?</concept>
  <qstn responseDomainType="category">
    <preQTxt>Please select one option</preQTxt>
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
  <catgry>
    <catValu>3</catValu>
    <labl>Non-binary</labl>
  </catgry>
  <varFormat type="numeric" schema="other"/>
</var>
```

#### Mapping Breakdown
```
XLSForm                           DDI
────────────────────────────────  ──────────────────────────────────
survey[0].type: "select_one"   →  responseDomainType="category"
                               →  intrvl="discrete"
                               →  varFormat type="numeric"

survey[0].name: "gender"       →  name="gender"
                               →  ID="V_gender" (auto-generated)

survey[0].label: "What is..."  →  <concept>What is...</concept>
                               →  <qstnLit>What is...</qstnLit>

survey[0].hint: "Please..."    →  <preQTxt>Please...</preQTxt>

choices (list_name="gender")   →  <catgry> elements:
  {name:"1", label:"Male"}     →    <catValu>1</catValu><labl>Male</labl>
```

---

### Example 2: Numeric Question (Integer)

#### XLSForm (JSON Input)
```json
{
  "survey": [
    {
      "type": "integer",
      "name": "age",
      "label": "What is your age in years?",
      "hint": "Enter a number between 0 and 120"
    }
  ],
  "choices": [],
  "settings": {}
}
```

#### DDI (XML Output)
```xml
<?xml version="1.0" encoding="UTF-8"?>
<var ID="V_age" name="age" intrvl="discrete">
  <concept>What is your age in years?</concept>
  <qstn responseDomainType="numeric">
    <preQTxt>Enter a number between 0 and 120</preQTxt>
    <qstnLit>What is your age in years?</qstnLit>
  </qstn>
  <varFormat type="numeric" schema="other"/>
</var>
```

---

### Example 3: Open Text Question

#### XLSForm (JSON Input)
```json
{
  "survey": [
    {
      "type": "text",
      "name": "comments",
      "label": "Please provide any additional comments",
      "hint": "Be as specific as possible"
    }
  ],
  "choices": [],
  "settings": {}
}
```

#### DDI (XML Output)
```xml
<?xml version="1.0" encoding="UTF-8"?>
<var ID="V_comments" name="comments" intrvl="contin">
  <concept>Please provide any additional comments</concept>
  <qstn responseDomainType="text">
    <preQTxt>Be as specific as possible</preQTxt>
    <qstnLit>Please provide any additional comments</qstnLit>
  </qstn>
  <varFormat type="character" schema="other"/>
</var>
```

---

### Example 4: Multiple Choice Question (Checkboxes)

Per DDI Codebook conventions, `select_multiple` (checkboxes) produces a `<varGrp type="multipleResp">` with one binary `<var>` per choice option (0 = Not mentioned, 1 = Mentioned).

#### XLSForm (JSON Input)
```json
{
  "survey": [
    {
      "type": "select_multiple hobbies",
      "name": "hobbies",
      "label": "What are your hobbies? (Select all that apply)"
    }
  ],
  "choices": [
    {"list_name": "hobbies", "name": "reading", "label": "Reading"},
    {"list_name": "hobbies", "name": "sports", "label": "Sports"},
    {"list_name": "hobbies", "name": "music", "label": "Music"}
  ],
  "settings": {}
}
```

#### DDI (XML Output)
```xml
<?xml version="1.0" encoding="UTF-8"?>
<dataDscr>
  <varGrp ID="VG_hobbies" name="hobbies" type="multipleResp" var="V_hobbies_reading V_hobbies_sports V_hobbies_music">
    <txt>What are your hobbies? (Select all that apply)</txt>
    <concept>What are your hobbies? (Select all that apply)</concept>
  </varGrp>
  <var ID="V_hobbies_reading" name="hobbies_reading" intrvl="discrete">
    <qstn responseDomainType="multiple">
      <preQTxt>What are your hobbies? (Select all that apply)</preQTxt>
      <qstnLit>Reading</qstnLit>
    </qstn>
    <catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
    <catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
    <concept>What are your hobbies? (Select all that apply): Reading</concept>
    <varFormat type="numeric" schema="other"/>
  </var>
  <var ID="V_hobbies_sports" name="hobbies_sports" intrvl="discrete">
    <qstn responseDomainType="multiple">
      <preQTxt>What are your hobbies? (Select all that apply)</preQTxt>
      <qstnLit>Sports</qstnLit>
    </qstn>
    <catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
    <catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
    <concept>What are your hobbies? (Select all that apply): Sports</concept>
    <varFormat type="numeric" schema="other"/>
  </var>
  <var ID="V_hobbies_music" name="hobbies_music" intrvl="discrete">
    <qstn responseDomainType="multiple">
      <preQTxt>What are your hobbies? (Select all that apply)</preQTxt>
      <qstnLit>Music</qstnLit>
    </qstn>
    <catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
    <catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
    <concept>What are your hobbies? (Select all that apply): Music</concept>
    <varFormat type="numeric" schema="other"/>
  </var>
</dataDscr>
```

---

### Example 5: Question Group

#### XLSForm (JSON Input)
```json
{
  "survey": [
    {"type": "begin_group", "name": "satisfaction_group", "label": "Satisfaction Questions"},
    {"type": "end_group", "name": ""}
  ],
  "choices": [],
  "settings": {}
}
```

#### DDI (XML Output)
```xml
<?xml version="1.0" encoding="UTF-8"?>
<varGrp ID="VG_satisfaction_group" name="satisfaction_group" type="multipleResp" var="">
  <concept>Satisfaction Questions</concept>
</varGrp>
```

#### Mapping Breakdown
```
XLSForm                              DDI
──────────────────────────────────  ──────────────────────────────────
survey[0].type: "begin_group"    →  <varGrp> element
                                 →  type="multipleResp" (inferred)

survey[0].name: "satisfaction.." →  name="satisfaction_group"
                                 →  ID="VG_satisfaction_group"

survey[0].label: "Satisfaction"  →  <concept>Satisfaction Questions</concept>

survey[1].type: "end_group"      →  (closes the group)
                                 →  var="" (would list variable IDs)
```

## Special Cases

### Missing Value Categories

DDI supports marking categories as "missing values" (e.g., "Don't know", "Refused to answer"):

```xml
<catgry missing="Y">
  <catValu>99</catValu>
  <labl>Don't know</labl>
</catgry>
```

When converting DDI → XLSForm, categories with `missing="Y"` are **excluded** from the choices sheet since XLSForm doesn't have an equivalent concept.

### Grid/Matrix Questions

In DDI, matrix questions use a `<varGrp type="grid">` containing multiple variables:

```xml
<varGrp ID="VG1" name="service_rating" type="grid" var="V1 V2 V3">
  <concept>Service Quality Assessment</concept>
  <txt>Please rate the following aspects:</txt>
</varGrp>

<var ID="V1" name="service_speed" intrvl="discrete">
  <concept>Speed of service</concept>
  <qstn responseDomainType="category">
    <qstnLit>Speed of service</qstnLit>
  </qstn>
  <!-- Same categories for all grid items -->
</var>
```

In XLSForm, this becomes begin_group/end_group rows wrapping individual question rows.

### Interviewer Instructions

Interviewer-facing instructions use the `parameters` column with `guidance_hint=`:

```json
{
  "survey": [
    {
      "type": "select_one gender",
      "name": "gender",
      "label": "What is your gender?",
      "parameters": "guidance_hint=Show response card to participant"
    }
  ]
}
```

Maps to:

```xml
<qstn>
  <ivuInstr>Show response card to participant</ivuInstr>
</qstn>
```

## Technical Notes

### Data Types (intrvl and varFormat.type)

DDI uses two attributes to describe data type:

1. **intrvl** (interval): Measurement scale
   - `discrete`: Countable values (1, 2, 3, ...)
   - `contin`: Continuous values (text, measurements)

2. **varFormat.type**: Storage type
   - `numeric`: Numbers (stored as integers/floats)
   - `character`: Text (stored as strings)

**Mapping rules:**
- Numbers → `intrvl="discrete"` + `varFormat.type="numeric"`
- Text → `intrvl="contin"` + `varFormat.type="character"`
- Categories → `intrvl="discrete"` + `varFormat.type="numeric"` (categories stored as numbers)

### ID Generation

When converting XLSForm → DDI, IDs are auto-generated:
- Variables: `V_` + variable name (e.g., `V_age`, `V_gender`)
- Groups: `VG_` + group name (e.g., `VG_demographics`)

Spaces in names are replaced with underscores.

## API Usage

See [CONVERSION_API.md](CONVERSION_API.md) for API endpoint documentation and curl examples.

## Limitations

1. **One-way loss**: Some DDI metadata (study-level info, full codebook structure) is not present in individual XLSForm questions
2. **Group variables**: The `var` attribute (listing variable IDs) is empty when converting XLSForm groups without nested questions to DDI
3. **Missing values**: DDI missing value categories are dropped during DDI → XLSForm conversion
4. **Advanced features**: Some XLSForm features (skip logic, constraints, calculations) are not represented in DDI

## See Also

- [CONVERSION_API.md](CONVERSION_API.md) - API endpoint documentation
- [DDI_MARKUP_GUIDE.md](DDI_MARKUP_GUIDE.md) - DDI Codebook conventions
- [examples/single_question_example.xml](examples/single_question_example.xml) - Complete DDI examples
