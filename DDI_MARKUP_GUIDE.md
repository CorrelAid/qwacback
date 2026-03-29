# DDI 2.5 Markup Guide

This guide defines the conventions for mapping survey questionnaires to DDI 2.5 Codebook XML in this project.

### Validation

This project validates DDI XML in two layers. The **DDI 2.5 XSD** defines structure and element ordering but is intentionally permissive — most elements are optional. The project-specific **Schematron rules** (`schematron/ddi_custom_rules.sch`) enforce stricter conventions on top: required elements (e.g. `concept` on every `var`/`varGrp`), forbidden elements (e.g. `labl` on `var`/`varGrp`), restricted `varGrp` types, category/label requirements, and consistency checks (e.g. `preQTxt` matching `varGrp/txt`). The rules in this guide reflect both layers combined.

### Examples

Each answer type has a complete example with both the **XLSForm input** and the **DDI Codebook output**, defined in `internal/examples/examples.go`. These are available at runtime via the API:

*   `GET /api/examples` — list all examples (XLSForm + DDI pairs)
*   `GET /api/examples/{answer_type}` — get a single example by type (e.g. `single_choice`, `multiple_choice`, `grid`)

---

## 1. Identifiers: `name`, `concept`, `labl`, and `ID`

### `name` vs `concept`

| | `name` (attribute) | `concept` (child element) |
|---|---|---|
| **Audience** | Machines / analysts | Humans / researchers |
| **Format** | `snake_case`, no spaces | Natural language prose |
| **Purpose** | Dataset column name — used in code, exports, database keys | Names what the variable or group measures |
| **Length** | Short (1–3 words) | Phrase or sentence fragment |
| **Example** | `institutional_trust` | `Trust in public institutions` |

Think of `name` as what you'd call the column in a CSV file, and `concept` as the codebook label that tells a researcher what the variable captures.

#### `name` rules

*   Name the **concept**, not the question text: `geschlecht` not `welches_geschlecht_haben_sie`.
*   Keep names short (1–3 words): `alter`, `schulabschluss`, `haushaltseinkommen`.
*   For grid/checkbox sub-items, prefix with the group name: `institutional_trust_parliament`, `device_ownership_smartphone`.

#### `concept` rules

`concept` is **required** on every `var` and `varGrp`. It names what the variable or group measures — this does not need to be a latent construct. For simple demographic or factual items, a plain descriptive phrase is correct.

*   **Standalone variables**: Describe what the question measures in plain language (e.g. `Gender`, `Age group`, `Perceived community safety (day)`).
*   **Grid sub-items**: Use a `Construct: Facet` pattern (e.g. `Trust: Parliament`, `Civic network: Council`).
*   **Variable groups**: Name the overarching topic or construct (e.g. `Trust in institutions`, `Device ownership`).
*   The `concept` element supports `vocab` and `vocabURI` attributes to link to controlled vocabularies.

### `labl`

`labl` is used **only on `catgry`** elements — it provides the human-readable label for a response value (e.g. "Strongly Agree", "Male"). It is **never** placed on `var` or `varGrp`; use `concept` there instead.

`labl` is **required** for `responseDomainType="category"` (single choice, grid) but **omitted** for `responseDomainType="multiple"` (checkboxes). Checkbox variables are binary 0/1 — the variable name and `qstnLit` already describe what the checkbox means.

### `notes`

`notes` is an optional free-text annotation on a `<var>`. Use it for methodology notes, source attribution, or other metadata that does not fit into the structured fields. Only **one** `notes` element per variable is allowed (enforced by Schematron). It must appear **after** `varFormat` (last child of `<var>`).

Example:
```xml
<var ID="V_geschlecht" name="geschlecht" intrvl="discrete">
  <qstn responseDomainType="category">
    <qstnLit>Was ist Ihr Geschlecht?</qstnLit>
  </qstn>
  <concept>Geschlecht</concept>
  <varFormat type="numeric" schema="other"/>
  <notes>Instrument nach Diethold (2023), Option B.</notes>
</var>
```

### `ID`

The `ID` attribute (`xs:ID`) is a document-wide unique identifier used for cross-referencing (e.g. `varGrp var="V_item1 V_item2"`). It does **not** need to be human-readable — use `name` for that. Convention: `V_<name>` for variables, `VG_<name>` for groups.

---

## 2. Structure Rules

### Variable group types

Only two `varGrp` types are used:

| Type | Purpose |
| :--- | :--- |
| `grid` | Matrix / Likert grid — items sharing the same scale and introductory text |
| `multipleResp` | Select-multiple (checkboxes) — each option becomes a binary 0/1 variable |

Do **not** use `type="section"` — section groups are structural containers with no semantic meaning.

### Child element ordering within `<var>`

The XSD enforces a strict sequence inside `<var>`. Simplified to the elements this project uses:

```
qstn → catgry* → concept → varFormat → notes?
```

`notes` is optional and must appear **after** `varFormat` (last child of `<var>`). Only one `notes` element per variable is allowed (enforced by Schematron).

### `<varGrp>` placement in `<dataDscr>`

All `<varGrp>` elements must appear **before** all `<var>` elements in `<dataDscr>`. This is an XSD ordering requirement.

### Document / questionnaire ordering

DDI 2.5 has no explicit ordering attribute. Order is determined by **document position** — elements are stored and restored in the sequence they appear in the XML file. Always place elements in questionnaire order (varGrps first, then vars).

On import, the application captures each element's position as a numeric `order` field in the database. On export, variables and groups are sorted by this field, preserving the document sequence through round trips.

> **Note**: The `qstn/@seqNo` attribute in the XSD is intended for question-flow numbering within instruments, not variable ordering. Do not rely on it.

---

## 3. Answer Type Mappings

### Base types (DB `answer_type`)

| `answer_type` | `intrvl` | `responseDomainType` | `varFormat/@type` | Container |
|--------|----------|----------------------|-------------------|-----------|
| `integer` | `contin` | `numeric` | `numeric` | `<var>` |
| `text` | `discrete` | `text` | `character` | `<var>` |
| `single_choice` | `discrete` | `category` | `numeric` | `<var>` + `<catgry>` per option |
| `multiple_choice` | `discrete` | `multiple` | `numeric` | `<varGrp type="multipleResp">` + binary `<var>` per option |
| `grid` | `discrete` | `category` | `numeric` | `<varGrp type="grid">` + `<var>` per item (categories repeated) |

### Subcategory flags

Two boolean fields on the `variables` collection modify the base type:

| Flag | Applies to | Effect |
|------|-----------|--------|
| `has_other` | `single_choice`, `multiple_choice` | A companion `_other` text variable exists for free-text specification |
| `has_long_list` | `single_choice`, `multiple_choice` | Categories come from an external code list via `concept/@vocab` instead of inline `<catgry>` |

### Semi-open questions (`_other` convention)

A semi-open (halb-offen) question provides a closed choice list plus an optional free-text field for respondents who select an "other" option. This produces **two variables** in DDI: the main categorical/checkbox variable, and a text variable named `<name>_other` for the free-text specification.

DDI Codebook has no concept of skip/relevance logic. The converter reconstructs XLSForm relevance from naming conventions during DDI→XLSForm conversion, ensuring lossless round-trips.

**Conventions for `_other` variables:**
*   The "other" category must use `catValu="other"`. The label can be localized (e.g. "Sonstiges").
*   The `_other` text variable must have `responseDomainType="text"`, `intrvl="contin"`, `varFormat type="character"`.
*   A matching base variable or group with the prefix name must exist.
*   For `multiple_choice_other`: the `_other` text variable must **not** be listed in the `varGrp/@var` attribute. No binary var is created for the "other" choice — the `_other` text var represents it outside the group.
*   **Round-trip**: DDI→XLSForm reconstructs relevance from naming convention: `${base} = 'other'` (single choice) or `selected(${base}, 'other')` (multiple choice). XLSForm→DDI drops relevance (reconstructable). For multiple choice, the converter synthesizes an `other` choice from the `_other` text var, and skips creating a binary var for it on the way back.

### Grid and checkbox group consistency

*   The `varGrp/txt` and each member variable's `qstn/preQTxt` must be identical.
*   `multipleResp` members must have `responseDomainType="multiple"`.
*   `grid` members must have `responseDomainType="category"`.
*   Grid categories **must be repeated** on every member variable (DDI 2.5 has no shared category reference).
*   Checkbox categories have **no `labl`** — the binary 0/1 values are self-explanatory.

### External code lists (`_long` types)

When a question draws from a large external code list (e.g. ISO 3166-1 country codes), inline `<catgry>` elements are replaced by a `concept/@vocab` reference that identifies the standard (e.g. `vocab="iso_3166_1"`).

In XLSForm, this corresponds to `select_one_from_file` or `select_multiple_from_file`.

#### single_choice_long

```xml
<var ID="V_geburtsland" name="geburtsland" intrvl="discrete">
  <qstn responseDomainType="category">
    <qstnLit>In welchem Land wurden Sie geboren?</qstnLit>
  </qstn>
  <concept vocab="iso_3166_1">In welchem Land wurden Sie geboren?</concept>
  <varFormat type="numeric" schema="other"/>
</var>
```

#### multiple_choice_long

Same structure as `single_choice_long` but with `responseDomainType="multiple"`. This is a **standalone `<var>`**, not a `varGrp` with binary member variables — the external file replaces the need for inline expansion.

```xml
<var ID="V_herkunftslaender" name="herkunftslaender" intrvl="discrete">
  <qstn responseDomainType="multiple">
    <qstnLit>Aus welchen Ländern stammen die Menschen, die Ihre Angebote nutzen? Mehrere Antworten möglich.</qstnLit>
  </qstn>
  <concept vocab="iso_3166_1">Herkunftsländer der Nutzer*innen</concept>
  <varFormat type="numeric" schema="other"/>
</var>
```

> **Note**: The schematron rules exempt variables with `concept/@vocab` from the requirement to have inline `catgry` elements.

---

## 4. Study-Level Date Elements

Three date elements serve distinct purposes:

| Element | Location | Purpose |
|---------|----------|---------|
| `prodDate` | `citation/prodStmt` | When the **publication** was produced/released (not distributed or archived) |
| `timePrd` | `stdyInfo/sumDscr` | The **time period the data refer to** — not the dates of coding or collection |
| `collDate` | `stdyInfo/sumDscr` | When the **data were actually collected** (e.g. fieldwork dates) |

### Key distinctions

- `timePrd` ≠ `collDate`: A survey collected in Jan–Mar 2020 (`collDate`) may ask about activities in 2019 (`timePrd`). For a project running Oct 2018–Dec 2020, `timePrd` covers that full span.
- `prodDate` is about the publication artifact, not the data.
- `collDate` belongs in `sumDscr`, **not** in `method/dataColl` (the XSD does not allow it there).

### Event attribute

Both `timePrd` and `collDate` use the `event` attribute: `"start"`, `"end"`, or `"single"`. For date ranges, use two elements with `start` and `end`.

---

## 5. Template Placeholders

For template surveys where question wording is adapted per deployment, encode placeholders directly in `qstnLit` using angle brackets. Since `qstnLit` extends `simpleTextType` (which allows `xhtml:BlkNoForm.mix`), use XHTML to render placeholders in italics — wrap in `<xhtml:p>` since only block-level XHTML elements are valid direct children.

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
| `catgry` | `var` | Response category; contains `catValu` and optionally `labl` |
| `labl` | `catgry` | Human-readable label for a category value (required for `category`, omitted for `multiple`) |
| `concept` | `var`, `varGrp` | Human-readable name of what the variable or group measures (required) |
| `txt` | `varGrp` | Shared introductory question text for grid/checkbox groups |
| `varFormat` | `var` | Technical data format (must appear second-to-last inside `var`) |
| `notes` | `var` | Optional free-text annotation (max one per variable, must be last child) |

### Attributes

| Attribute | Element | Purpose |
| :--- | :--- | :--- |
| `ID` | `var`, `varGrp` | Document-wide unique identifier for cross-referencing (`xs:ID`) |
| `name` | `var`, `varGrp` | Abstract snake_case column name / machine identifier |
| `intrvl` | `var` | Measurement level: `discrete` or `contin` |
| `type` | `varGrp` | Group semantics: `grid` or `multipleResp` |
| `var` | `varGrp` | Space-separated list of member variable IDs |
| `responseDomainType` | `qstn` | Response type: `numeric`, `text`, `category`, or `multiple` |
