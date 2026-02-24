# DDI 2.5 Markup Guide

This guide defines the conventions used when mapping survey questionnaires to DDI 2.5 Codebook XML in this project. It covers naming, element semantics, structure rules, and all five question type mappings.

---

## 1. Key Concepts: `name`, `concept`, `labl`, and `ID`

These four identifiers are the most frequently confused. Each has a distinct role:

### `name` vs `concept`

| | `name` (attribute on `var`/`varGrp`) | `concept` (child element of `var`/`varGrp`) |
|---|---|---|
| **Audience** | Machines / analysts | Humans / researchers |
| **Format** | `snake_case`, no spaces | Natural language prose |
| **Purpose** | Dataset column name — used in code, exports, database keys | Names the latent construct or topic the question measures |
| **Length** | Short (1–3 words) | Phrase or sentence fragment |
| **Example** | `institutional_trust` | `Trust in public institutions` |

Think of `name` as what you'd call the column in a CSV file, and `concept` as the label you'd put in a codebook footnote explaining what the variable actually captures.

#### `name` rules

*   Name the **concept**, not the question text: `geschlecht` not `welches_geschlecht_haben_sie`.
*   Keep names short (1–3 words): `alter`, `schulabschluss`, `haushaltseinkommen`.
*   For grid/checkbox sub-items, prefix with the group name: `institutional_trust_parliament`, `device_ownership_smartphone`.

#### `concept` rules

*   **Standalone variables**: Name the latent variable the question operationalises.
    *   `<concept>Perceived community safety (day)</concept>`
    *   `<concept>Self-efficacy (attitudes)</concept>`
*   **Grid sub-items**: Use a `Construct: Facet` pattern.
    *   `<concept>Trust: Parliament</concept>`
    *   `<concept>Civic network: Council</concept>`
*   **Variable groups**: Name the overarching latent construct.
    *   `<concept>Trust in institutions</concept>`
    *   `<concept>Device ownership</concept>`
*   **Self-evident variables**: Omit `concept` for pure demographics or feedback fields (age, gender, open comments) where no latent construct is being operationalised.
*   The `concept` element supports `vocab` and `vocabURI` attributes to link to controlled vocabularies.

### `labl`

`labl` is used **only on `catgry`** elements — it provides the human-readable label for a response value (e.g. "Strongly Agree", "Male", "Not mentioned"). It is **never** placed on `var` or `varGrp`; use `concept` there instead.

### `ID`

The `ID` attribute (`xs:ID`) is a document-wide unique identifier used for cross-referencing (e.g. `varGrp var="V1 V2 V3"`). It does **not** need to be human-readable — use `name` for that. Convention: `V1`, `V2`, … for variables; `VG1`, `VG2`, … for groups.

---

## 2. Variable Group Types

Only two `varGrp` types are used:

| Type | Purpose |
| :--- | :--- |
| `grid` | Matrix / Likert grid — items sharing the same scale and introductory text |
| `multipleResp` | Select-multiple (checkboxes) — each option becomes a binary 0/1 variable |

Do **not** use `type="section"` — section groups are structural containers with no semantic meaning.

---

## 3. Structure Rules

### Child element ordering within `<var>`

The XSD enforces a strict sequence inside `<var>`:

```
qstn → catgry* → concept? → varFormat
```

### `<varGrp>` placement in `<dataDscr>`

All `<varGrp>` elements must appear **before** all `<var>` elements in `<dataDscr>`. This is an XSD ordering requirement.

```xml
<dataDscr>
  <varGrp ID="VG1" .../>  <!-- groups first -->
  <varGrp ID="VG2" .../>
  <var ID="V1" .../>      <!-- then variables -->
  <var ID="V2" .../>
</dataDscr>
```

### Document / questionnaire ordering

DDI 2.5 has no explicit ordering attribute. Order is determined by **document position** — elements are stored and restored in the sequence they appear in the XML file. Always place elements in questionnaire order within each group (varGrps first, then vars).

On import, the application captures each element's position as a numeric `order` field in the database. On export, variables and groups are sorted by this field, preserving the document sequence through round trips.

> **Note**: The `qstn/@seqNo` attribute in the XSD is intended for question-flow numbering within instruments, not variable ordering. Do not rely on it.

---

## 4. Question Type Mappings

The pipeline uses five question formats. For each, the table below shows the DDI encoding and a full XML example.

| Format | `intrvl` | `responseDomainType` | `varFormat/@type` | Container |
|--------|----------|----------------------|-------------------|-----------|
| `open_number` | `discrete` | `numeric` | `numeric` | `<var>` |
| `open_text` | `contin` | `text` | `character` | `<var>` |
| `single_choice` | `discrete` | `category` | `numeric` | `<var>` + `<catgry>` per option |
| `checkboxes` | `discrete` | `multiple` | `numeric` | `<varGrp type="multipleResp">` + binary `<var>` per option |
| `grid` | `discrete` | `category` | `numeric` | `<varGrp type="grid">` + `<var>` per item (categories repeated) |

### open_number — Numeric free entry

```xml
<var ID="V1" name="alter" intrvl="discrete">
  <qstn ID="Q1" responseDomainType="numeric">
    <qstnLit>What is your age?</qstnLit>
  </qstn>
  <varFormat type="numeric" schema="other"/>
</var>
```

### open_text — Free-text entry

```xml
<var ID="V2" name="open_comments" intrvl="contin">
  <qstn ID="Q2" responseDomainType="text">
    <qstnLit>Any other comments?</qstnLit>
  </qstn>
  <varFormat type="character" schema="other"/>
</var>
```

### single_choice — Pick exactly one option

```xml
<var ID="V3" name="geschlecht" intrvl="discrete">
  <qstn ID="Q3" responseDomainType="category">
    <qstnLit>What is your gender?</qstnLit>
  </qstn>
  <catgry><catValu>1</catValu><labl>Male</labl></catgry>
  <catgry><catValu>2</catValu><labl>Female</labl></catgry>
  <catgry><catValu>3</catValu><labl>Other</labl></catgry>
  <varFormat type="numeric" schema="other"/>
</var>
```

### checkboxes — Select all that apply

Each option becomes a separate binary variable (0 = not mentioned, 1 = mentioned). The `varGrp/txt` and each member variable's `qstn/preQTxt` must be identical.

```xml
<varGrp ID="VG1" name="device_ownership" type="multipleResp" var="V4 V5">
  <txt>Which of these devices do you own? Please check all that apply.</txt>
  <concept>Device ownership</concept>
</varGrp>

<var ID="V4" name="device_ownership_smartphone" intrvl="discrete">
  <qstn ID="Q4" responseDomainType="multiple">
    <preQTxt>Which of these devices do you own? Please check all that apply.</preQTxt>
    <qstnLit>Smartphone</qstnLit>
  </qstn>
  <catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
  <catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
  <concept>Device ownership: Smartphone</concept>
  <varFormat type="numeric" schema="other"/>
</var>

<var ID="V5" name="device_ownership_laptop" intrvl="discrete">
  <qstn ID="Q5" responseDomainType="multiple">
    <preQTxt>Which of these devices do you own? Please check all that apply.</preQTxt>
    <qstnLit>Laptop</qstnLit>
  </qstn>
  <catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
  <catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
  <concept>Device ownership: Laptop</concept>
  <varFormat type="numeric" schema="other"/>
</var>
```

### grid — Matrix / Likert scale

A set of items sharing the same scale and introductory stem. Categories **must be repeated** on every member variable (DDI 2.5 has no shared category reference). The `varGrp/txt` and each member's `qstn/preQTxt` must be identical.

```xml
<varGrp ID="VG2" name="institutional_trust" type="grid" var="V6 V7">
  <txt>How much do you trust each of the following institutions on a scale from 1 to 5?</txt>
  <concept>Trust in institutions</concept>
</varGrp>

<var ID="V6" name="institutional_trust_parliament" intrvl="discrete">
  <qstn ID="Q6" responseDomainType="category">
    <preQTxt>How much do you trust each of the following institutions on a scale from 1 to 5?</preQTxt>
    <qstnLit>The Parliament</qstnLit>
  </qstn>
  <catgry><catValu>1</catValu><labl>Not at all</labl></catgry>
  <catgry><catValu>2</catValu><labl>2</labl></catgry>
  <catgry><catValu>3</catValu><labl>3</labl></catgry>
  <catgry><catValu>4</catValu><labl>4</labl></catgry>
  <catgry><catValu>5</catValu><labl>Completely</labl></catgry>
  <concept>Trust: Parliament</concept>
  <varFormat type="numeric" schema="other"/>
</var>

<var ID="V7" name="institutional_trust_police" intrvl="discrete">
  <qstn ID="Q7" responseDomainType="category">
    <preQTxt>How much do you trust each of the following institutions on a scale from 1 to 5?</preQTxt>
    <qstnLit>The Police</qstnLit>
  </qstn>
  <catgry><catValu>1</catValu><labl>Not at all</labl></catgry>
  <catgry><catValu>2</catValu><labl>2</labl></catgry>
  <catgry><catValu>3</catValu><labl>3</labl></catgry>
  <catgry><catValu>4</catValu><labl>4</labl></catgry>
  <catgry><catValu>5</catValu><labl>Completely</labl></catgry>
  <concept>Trust: Police</concept>
  <varFormat type="numeric" schema="other"/>
</var>
```

---

## 5. Template Placeholders

For template surveys where question wording is adapted per deployment, encode placeholders directly in `qstnLit` using angle brackets. Since `qstnLit` extends `simpleTextType` (which allows `xhtml:BlkNoForm.mix`), use XHTML to render placeholders in italics — wrap in `<xhtml:p>` since only block-level XHTML elements are valid direct children.

```xml
<qstnLit xmlns:xhtml="http://www.w3.org/1999/xhtml">
  <xhtml:p>I feel safe out and about in <xhtml:em>&lt;PROJECT AREA&gt;</xhtml:em> during the day.</xhtml:p>
</qstnLit>
```

Rendered meaning: I feel safe out and about in *&lt;PROJECT AREA&gt;* during the day.

Rules:
*   Use `&lt;` and `&gt;` for literal angle brackets.
*   Wrap with `<xhtml:em>` for visual distinction.
*   Keep placeholder names short and uppercase: `<PROJECT AREA>`, `<TIME PERIOD>`, `<PROJECT/SPACE>`.
*   Do **not** embed instructions in `qstnLit`. The text should read naturally with the placeholder in place.

---

## 6. Quick Reference

### Elements

| Element | Parent | Purpose |
| :--- | :--- | :--- |
| `qstnLit` | `qstn` | Literal question text as presented to the respondent |
| `preQTxt` | `qstn` | Introductory context shown before the question (must match `varGrp/txt` for grid/checkbox items) |
| `postQTxt` | `qstn` | Text shown after the question |
| `ivuInstr` | `qstn` | Interviewer instructions (not shown to respondent) |
| `catgry` | `var` | Response category; contains `catValu` and `labl` |
| `labl` | `catgry` | Human-readable label for a category value |
| `concept` | `var`, `varGrp` | Human-readable name of the latent construct measured |
| `txt` | `varGrp` | Shared introductory question text for grid/checkbox groups |
| `varFormat` | `var` | Technical data format (must appear last inside `var`) |

### Attributes

| Attribute | Element | Purpose |
| :--- | :--- | :--- |
| `ID` | `var`, `varGrp` | Document-wide unique identifier for cross-referencing (`xs:ID`) |
| `name` | `var`, `varGrp` | Abstract snake_case column name / machine identifier |
| `intrvl` | `var` | Measurement level: `discrete` or `contin` |
| `type` | `varGrp` | Group semantics: `grid` or `multipleResp` |
| `var` | `varGrp` | Space-separated list of member variable IDs |
| `responseDomainType` | `qstn` | Response type: `numeric`, `text`, `category`, or `multiple` |
