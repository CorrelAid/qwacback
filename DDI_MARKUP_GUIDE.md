# DDI 2.5 to XLSForm Markup Guide

This guide describes how to map common survey question types (XLSForm) to semantically accurate DDI 2.5 Codebook XML.

---

## General Conventions

### The `name` Attribute: Descriptive Variable Names

The `name` attribute on `var` and `varGrp` should be a **descriptive, snake_case identifier** that reflects what the variable captures. Think of it as the column name you would use in a dataset.

*   Keep names short but meaningful.
*   Use snake_case: `safety_day`, `contact_council`, `convo_diff_age`.
*   For grid sub-items, include the group context: `contact_community_groups`, `convo_diff_ethnicity`.
*   For demographics, use the concept directly: `age_group`, `gender`, `employment_status`.

The `ID` attribute (`xs:ID`) is for document-wide cross-referencing (e.g. `varGrp var="V4a V4b"`). It does **not** need to be human-readable — use `name` for that.

### The `concept` Element: What the Variable Measures

Use the `concept` element on `var` and `varGrp` to tag the **latent construct or topic** the item measures. This replaces using `labl` for construct names.

*   **Standalone variables**: The concept names the latent variable the question operationalises.
    *   `<concept>Perceived community safety (day)</concept>`
    *   `<concept>Self-efficacy (attitudes)</concept>`
*   **Grid sub-items** (`varGrp type="grid"`): Use a prefix pattern `Construct: Facet`.
    *   `<concept>Civic network: Council</concept>`
    *   `<concept>Bridging: Intergenerational</concept>`
*   **Variable groups**: The concept names the overarching latent construct.
    *   `<concept>Civic network awareness</concept>`
    *   `<concept>Bridging social capital</concept>`
*   **Self-evident variables**: Omit `concept` when no latent construct is being measured (e.g. demographics like age, gender, employment status, or open-ended feedback fields).

The `concept` element supports `vocab` and `vocabURI` attributes to link to controlled vocabularies when available.

XSD sequence note: on `var`, `concept` appears **after** `catgry` and **before** `varFormat`. On `varGrp`, it appears **after** `txt`.

### The `labl` Element

In our markup, `labl` is used **only on `catgry`** (category) elements to provide human-readable labels for response values (e.g. "Strongly Agree", "Yes", "Male"). It is **not** used on `var` or `varGrp` — use `concept` instead for construct tagging and `name` for the descriptive identifier.

### Variable Group Types

Only the following `varGrp` types are used:

| Type | Purpose |
| :--- | :--- |
| `grid` | Matrix / Likert grid — a set of items sharing the same scale and introductory text |
| `multipleResp` | Select-multiple (checkboxes) — each option becomes a binary 0/1 variable |

Do **not** use `type="section"` — section groups are structural containers that do not carry semantic meaning and are not stored in the database.

### Ordering

DDI 2.5 Codebook has **no explicit ordering attribute** on `var` or `varGrp`. The order of variables and groups is determined by their **document position** — i.e. the sequence in which they appear in the XML file.

When writing or editing a codebook, place elements in the intended questionnaire order:

```xml
<dataDscr>
  <!-- Variables appear in questionnaire order -->
  <var ID="V1" name="safety_day">...</var>
  <var ID="V2" name="safety_night">...</var>
  <var ID="V3" name="contact_council">...</var>
  <!-- Groups also appear in questionnaire order -->
  <varGrp ID="VG1" name="civic_network" type="grid" var="V3 V4">...</varGrp>
  <varGrp ID="VG2" name="social_interactions" type="grid" var="V5a V5b">...</varGrp>
</dataDscr>
```

On import, the application captures each element's position as a numeric `order` field in the database. On export, variables and groups are sorted by this `order` field, preserving the original document sequence through round trips.

> **Note**: The `qstn` element has a `seqNo` attribute in the XSD, but it is intended for question-flow numbering within instruments, not for variable ordering. Do not rely on it for ordering purposes.

### Template Placeholders

For template surveys (where question wording must be adapted per deployment), encode placeholders directly in `qstnLit` using angle brackets. Since `qstnLit` extends `simpleTextType` (which allows `xhtml:BlkNoForm.mix`), you can use XHTML to render placeholders in italics. This requires wrapping the text in `<xhtml:p>` since only block-level XHTML elements are valid direct children.

```xml
<qstnLit xmlns:xhtml="http://www.w3.org/1999/xhtml">
  <xhtml:p>I feel safe out and about in <xhtml:em>&lt;PROJECT AREA&gt;</xhtml:em> during the day.</xhtml:p>
</qstnLit>
```

Rendered meaning: I feel safe out and about in *&lt;PROJECT AREA&gt;* during the day.

Rules:
*   Use `&lt;` and `&gt;` to produce literal angle brackets in the output.
*   Wrap with `<xhtml:em>` for visual distinction.
*   Keep placeholder names short and uppercase: `<PROJECT AREA>`, `<TIME PERIOD>`, `<PROJECT/SPACE>`.
*   Do **not** put instructions in `qstnLit` (e.g. "insert name of specific project/space"). The literal question text should read naturally with the placeholder in place.

---

## Question Type Mappings

### 1. Integer

*   **XLSForm Type**: `integer`
*   **DDI Example**:
```xml
<var ID="V1" name="age" intrvl="discrete">
  <qstn ID="Q1" responseDomainType="numeric">
    <qstnLit>What is your age?</qstnLit>
  </qstn>
  <varFormat type="numeric" schema="other"/>
</var>
```

### 2. Text (Open-Ended)

*   **XLSForm Type**: `text`
*   **DDI Example**:
```xml
<var ID="V2" name="other_comments" intrvl="contin">
  <qstn ID="Q2" responseDomainType="text">
    <qstnLit>Any Other Comments?</qstnLit>
  </qstn>
  <varFormat type="character" schema="other"/>
</var>
```

### 3. Select One

*   **XLSForm Type**: `select_one [list_name]`
*   **DDI Example**:
```xml
<var ID="V3" name="gender" intrvl="discrete">
  <qstn ID="Q3" responseDomainType="category">
    <qstnLit>What is your gender?</qstnLit>
  </qstn>
  <catgry><catValu>1</catValu><labl>Male</labl></catgry>
  <catgry><catValu>2</catValu><labl>Female</labl></catgry>
  <varFormat type="numeric" schema="other"/>
</var>
```

### 4. Select Multiple (Checkboxes)

*   **XLSForm Type**: `select_multiple [list_name]`
*   **DDI Mapping**: Use a `varGrp` with `type="multipleResp"`. Each option becomes a binary variable (0/1).
*   **DDI Example**:
```xml
<varGrp ID="VG1" name="devices_owned" type="multipleResp" var="V4_1 V4_2">
  <txt>Which of these devices do you own? Please check all that apply.</txt>
  <concept>Device ownership</concept>
</varGrp>

<var ID="V4_1" name="device_phone" intrvl="discrete">
  <qstn ID="Q4_1" responseDomainType="multiple">
    <preQTxt>Which of these devices do you own?...</preQTxt>
    <qstnLit>Smartphone</qstnLit>
  </qstn>
  <catgry><catValu>0</catValu><labl>Not mentioned</labl></catgry>
  <catgry><catValu>1</catValu><labl>Mentioned</labl></catgry>
  <concept>Device: Smartphone</concept>
  <varFormat type="numeric" schema="other"/>
</var>
```

### 5. Matrix / Likert Grid

Used for a group of questions that share the same scale and introductory text.

*   **XLSForm Type**: `begin_kobomatrix` / `begin_group` with appearance `grid`.
*   **DDI Mapping**: Use a `varGrp` with `type="grid"`.
*   **Consistency Rule**: The `varGrp/txt` and the `qstn/preQTxt` of every variable in the group SHOULD be identical. This ensures the question context is preserved at both the group and variable levels.
*   **DDI Example**:
```xml
<varGrp ID="VG2" name="trust_institutions" type="grid" var="V5_1 V5_2">
  <txt>How much do you trust each of the following institutions on a scale from 1 to 5?</txt>
  <concept>Trust in institutions</concept>
</varGrp>

<var ID="V5_1" name="trust_parliament" intrvl="discrete">
  <qstn ID="Q5_1" responseDomainType="category">
    <preQTxt>How much do you trust each of the following institutions...?</preQTxt>
    <qstnLit>The Parliament</qstnLit>
  </qstn>
  <!-- DDI 2.5 NOTE: Categories MUST be repeated for every variable in the grid -->
  <catgry><catValu>1</catValu><labl>Not at all</labl></catgry>
  <catgry><catValu>5</catValu><labl>Completely</labl></catgry>
  <concept>Trust: Parliament</concept>
  <varFormat type="numeric" schema="other"/>
</var>
```

---

## Element Quick Reference

| Element | Parent | Purpose |
| :--- | :--- | :--- |
| `concept` | `var`, `varGrp` | Tags the latent construct / topic the variable measures |
| `labl` | `catgry` | Human-readable label for a category value |
| `txt` | `varGrp` | Longer description; for grids this holds the shared introductory question |
| `qstnLit` | `qstn` | The literal question text as presented to the respondent |
| `preQTxt` | `qstn` | Introductory/context text shown before the question (duplicates `varGrp/txt` for grid items) |
| `postQTxt` | `qstn` | Text shown after the question |
| `ivuInstr` | `qstn` | Interviewer instructions (not shown to respondent) |

## Attribute Quick Reference

| Attribute | Element | Purpose |
| :--- | :--- | :--- |
| `ID` | `var`, `varGrp` | Document-wide unique identifier for cross-referencing (`xs:ID`) |
| `name` | `var`, `varGrp` | Descriptive identifier / data file column name (`xs:string`) |
| `intrvl` | `var` | Measurement level: `discrete` or `contin` |
| `type` | `varGrp` | Group semantics: `grid` or `multipleResp` (not `section`) |
| `var` | `varGrp` | Space-separated list of member variable IDs |
| `responseDomainType` | `qstn` | Response type: `numeric`, `text`, `category`, `multiple` |

## Summary Table

| XLSForm Type | DDI Interval | Response Domain | Container | Item literal location |
| :--- | :--- | :--- | :--- | :--- |
| `integer` | `discrete` | `numeric` | `var` | `qstnLit` |
| `text` | `contin` | `text` | `var` | `qstnLit` |
| `select_one` | `discrete` | `category` | `var` | `qstnLit` |
| `select_multiple` | `discrete` | `multiple` | `varGrp(multipleResp)` | `qstnLit` (w/ `preQTxt`) |
| `matrix` | `discrete` | `category` | `varGrp(grid)` | `qstnLit` (w/ `preQTxt`) |
