# q21-ecs2019: European Company Survey 2019, management questionnaire (Eurofound & Cedefop)

**Verdict: `ready`** (license: Eurofound "reproduction is authorised provided the source is acknowledged"; not CC; user confirms at merge)

## Source

- **Text source:** "European Company Survey 2019 – Online questionnaire for Management respondents", Eurofound,
  version for publication ("clean"), dated 14 July 2020 in the file name and PDF metadata. 21 pages, text layer, no
  scanned pages.
  - PDF: https://assets.eurofound.europa.eu/f/279033/9287637f70/ecs_2019_-_mm_questionnaire_for_publication_-_clean_14_july_2020.pdf
  - Linked from the ECS 2019 survey page: https://www.eurofound.europa.eu/en/surveys-and-data/surveys/european-company-survey/ecs-2019
  - Retrieved 2026-10-28. sha256 a3d528407d89866371ae01edee00fcdbd10df758e9dbb530bc9e02892d9813c7
    (`downloads/q21-ecs2019/ecs2019_mm_questionnaire.pdf`).
- **What this PDF is,** in its own "Information note": the English source (online) questionnaire, "adapted into an
  easier-to-read format".
  - Of the three wording versions (single independent companies, subsidiary sites of multi-site companies,
    headquarters), it shows the one for **subsidiary sites of multi-site companies**.
  - Question numbers and bracketed variable labels were added for the publication.
  - The routing instructions are in CAPITALS.
  - The definitions behind the (i) symbol are printed after the question stem.
- This is the latest (and only) published English version of the management questionnaire found. All wording comes from
  this one PDF.
- **Context only, not transcribed:** the overview report, Eurofound and Cedefop (2020), *European Company Survey 2019:
  Workplace practices unlocking employee potential* (EF20001), https://assets.eurofound.europa.eu/f/279033/3fe61d5438/ef20001en.pdf.
  - Retrieved 2026-10-28. sha256 194f6afc04a9e9afa3a71256856180bd7f6a0640e8f6919e533fd81315b8038d.
  - Used for the copyright statement and for the population definition.
- **Citation:** Eurofound & Cedefop. (2020). European Company Survey 2019: Online questionnaire for management
  respondents [Questionnaire, version for publication, 14 July 2020]. European Foundation for the Improvement of Living
  and Working Conditions. https://assets.eurofound.europa.eu/f/279033/9287637f70/ecs_2019_-_mm_questionnaire_for_publication_-_clean_14_july_2020.pdf
- **Population** (overview report): "all establishments with 10 or more employees in economic sectors engaged in what are
  termed ‘market activities’ in all 27 EU Member States and the United Kingdom"; fieldwork 2019. The respondent is the
  most senior manager in charge of personnel.
  - **Note for the user's use case:** German municipalities (public administration) were outside the ECS population.
    The items are written for "establishments", not public bodies.

## License / terms

- **The questionnaire PDF** carries no copyright or reuse notice.
- **ECS 2019 overview report (EF20001), copyright page (p. 2), verbatim:**
  > © European Foundation for the Improvement of Living and Working Conditions, 2020 and European Centre for the
  > Development of Vocational Training, 2020
  > Reproduction is authorised provided the source is acknowledged.
  > For any use or reproduction of photos or other material that is not under the Eurofound copyright, permission must
  > be sought directly from the copyright holders.
  > […]
  > Any queries on copyright must be addressed in writing to: copyright@eurofound.europa.eu
- **Eurofound "Legal information" page:** https://www.eurofound.europa.eu/en/legal-information (also served at
  /pages/legal-information and /about-eurofound/who-we-are/legal-information).
  - **Could not be fetched.** Every request returned a Vercel bot checkpoint (curl) or HTTP 429 (WebFetch). The
    Wayback Machine also returned 429.
  - Verbatim text, pasted by the user from the page on 2026-10-28 (section "Copyright"):
    > The reuse of any information of this website is authorised for commercial and non-commercial purposes, under the following conditions:
    > - The re-user is obliged to acknowledge the source of the document.
    > - The original meaning or the message of the documents should not be distorted.
    > - Eurofound cannot be held liable for any consequence stemming from the reuse.
    >
    > Reuse is understood as: ‘reproduction of textual data and multimedia items which are the property of Eurofound or of third parties and for which Eurofound holds the rights of use’.
    >
    > Where prior permission must be obtained for the reproduction or use of textual and multimedia information (sound, images, software, etc.), such permission shall cancel the above-mentioned general permission and shall clearly indicate any restrictions on use.
- **Assessment.** The PDF is published by Eurofound for documentation. Eurofound's general terms allow reuse with
  attribution and without distortion. The import is a verbatim format conversion with attribution, which fits those
  terms (cf. the user's "format change" decision for ND).
  - The questionnaire was developed by Eurofound and Cedefop, with Ipsos as the fieldwork contractor. No third-party
    rights are claimed in the PDF.
- `study.license` states:
  - the terms;
  - the changes: a format conversion, wording unchanged, with the welcome/closing texts, the ER contact block, the
    consent questions and the layout omitted;
  - a non-responsibility line, as in q19/q20.
- **German version:** not added. Eurofound's ECS 2019 page reportedly offers national-language questionnaires, but the
  page could not be loaded (bot checkpoint). Any German version would also be a separate translation, possibly of a
  different wording variant. English only.

## Transcription

- **Rule:** every label, hint, guidance_hint and choice label is literal text from the PDF.
  - Lines that the layout wraps are joined with one space. Separate printed paragraphs stay separate lines (`\n`).
    Nothing is joined with dashes.
  - **Kept as printed (sic):**
    - "it refers the local site" (Q1);
    - "No, we made loss" (Q69);
    - "The term establishment' refers …" (Q50, missing opening quote);
    - "1-DIFIT CODE" (Q72 placeholder);
    - the mixed quotes in Q7 (‘Not applicable’ vs 'We do not engage in export (0%)').
  - Bracketed variable labels ([MANAGER], [DISCHELP] …) are not text. They give the `name`s.
- **Text roles:**
  - Hints: the italic question-type lines ("Tick one box only.", "For each statement, Tick one box only.", "Tick all
    that apply.") and the (i) definitions.
  - Guidance hints: the CAPITAL routing lines ("Yes ➔ GO TO QUESTION 17", "(ASK IF QUESTION 22 = YES OR QUESTION 23 =
    YES)") and the placeholders "[COUNTRY SPECIFIC LIST]".
- **Structure:**
  - **Single-choice questions:** `select_one`. Codes are 1..n in printed order. The PDF prints no codes; these are ours.
  - **Item batteries with one scale:** `table-list` grids.
    - Q27 `disc`, Q28 `moti`, Q41 `tr`, Q53 `invol`, Q57 `mmepin`, Q58 `mmerin`.
    - Item names come from the printed labels, e.g. `disc_help` for [DISCHELP] and `invol_regmee` for [REGMEE].
    - Q53 has no question label of its own, so it is named `invol`.
    - Q58 reuses the item labels [MMEPINORG]…[MMEPINPAY] of Q57 in the PDF; here they are `mmerin_org` … `mmerin_pay`.
  - **"Number or percentage" questions:** "Enter the exact number of employees or tick one of the percentage
    categories".
    - Covers Q13–15, 19, 29–31, 34, 35, 37, 38 and 43.
    - The question is printed once above two alternative fields, "Number" and "Percentage". A grid takes only one row
      type, so each question becomes **two one-row `table-list` grids** that share the printed question as label:
      - `<code>_number` with the integer row `<code>_n` "Number";
      - `<code>_percentage` with the `select_one pct` row `<code>_pct` "Percentage".
    - This keeps the question as `preQTxt` on both variables. A plain section would drop it, and the bank would show
      questions called just "Number".
    - The printed definition and instruction are the hint of both grids.
    - Q46 (variable pay, 4 items) works the same way: the grid `vp_number` has 4 integer rows and the grid
      `vp_percentage` has 4 percentage rows, labelled with the item texts. "Number"/"Percentage" are in the hint.
  - **Ranking questions:** Q36 `trainlearn` (1–3), Q45 `hir` (1–4) and Q65 `pmstrat` (1–4) are number grids (integer
    rows) with a range constraint. Q36's items have no printed labels, so they are named `trainlearn_1..3`.
  - **Q32 SKILL:** two alternative number tables, `skill_pct` (percentages) and `skill_n` (exact numbers).
    - Both carry the printed stem as label. The "QUESTION USING …" heading and its paragraph are the hint.
    - The printed totals ("100 %", "[NUMBER]") are layout and are not transcribed.
    - There is no routing between the two; both are shown.
  - **Q1** has an integer field plus a "Don’t know" tick box. The box is the row `wpsize_mm_n_dk` (`select_multiple`
    with one option "Don’t know"), labelled with the printed instruction "Indicate a number or tick the ‘Don’t know’
    box."
  - **Q58** has the tick box "There is no employee representation at this establishment", which becomes the row
    `mmerin_noer`. Its label is the printed sentence "If there is no employee representation at your establishment,
    tick the box below and continue with the next question."
  - **Q2/Q72 (NACE lists):** the list is not printed, only a placeholder. The qwac build accepts no NACE vocabulary
    (`select_one_from_file` only knows ISO 3166), so the choices are the printed placeholder line plus "Don’t know" and
    "No answer", with codes `nace`/`dk`/`na`.
  - **Notes:** three printed lead-in sentences (before Q55, before Q59, and before Q73).
- **Routing → `relevant`:**
  - Q17, Q18: TEAMEX = Yes.
  - Q20–22: not (ICTCOMP number = 0 or percentage = "None at all").
  - Q24: Q22 or Q23 = Yes.
  - Q40, Q41: TRAINATALL = Yes.
  - Q43: NEWLYHIR ≥ 1.
  - Q50–52: "There is no official employee representation" not ticked in Q49. That option is exclusive (constraint).
  - Q58 grid: the "no employee representation" box not ticked.
  - Q70: PROFIT ≠ "Not applicable …".
  - Q72: MAINACT = the NACE placeholder, which stands for "a code was selected".
  - Constraints that are not text: years have 4 digits; numbers are ≥ 0; percentages are 0–100; ranks are 1–3 or 1–4.
- **Count:**
  - **74 printed questions (Q1–Q74), 127 answer rows** (86 `select_one`, 37 `integer`, 4 `select_multiple`), 3 notes
    and 37 groups, giving 204 form rows. There are 37 choice lists with 142 options.
  - Build: no errors and no warnings. The round trip keeps 130 rows in and 130 out.
  - **Validation (`validate.sh`): valid.**

## Character-by-character check

- `verify-labels.py` (in this folder) checks every label, hint, guidance_hint and choice label, line by line, against
  the PDF text.
  - It builds two corpora: `pdftotext -raw`, which keeps the two-column grids item by item but glues some words, and
    `pdftotext -layout`, which keeps the spaces.
  - It removes only the tick-box glyph ☐ and collapses whitespace.
  - Line-end hyphens may be skipped or kept. Quotes, apostrophes, dashes, arrows and case must match exactly.
- **Run:** `python3 -I tools/survey-import/out/q21-ecs2019/verify-labels.py tools/survey-import/out/q21-ecs2019/form.json tools/survey-import/downloads/q21-ecs2019/ecs2019_mm_questionnaire.pdf`
- **Result (2026-10-28): 443 text lines (295 distinct) from 407 texts, 0 not found.**
  - One line matched only when whitespace is ignored: the Q46 item "Variable extra pay linked to the results of the
    company or establishment (profit sharing scheme)". The raw layer has "extrapaylinked … resultsofthe", and it sits in
    a two-column cell. Its spacing was checked by hand against the `-layout` text.
- **Negative test:** five planted errors in a copy of the form were all flagged:
  - "Don't" with a straight apostrophe;
  - "customized";
  - "demonstration";
  - "Very Bad";
  - an extra comma.
- **Limits:**
  - A line is found anywhere in the PDF, not necessarily under the same question. The pairing was checked by reading the
    PDF against `gen-form.mjs`.
  - Short texts ("Yes", "Number") pass trivially.

## Skipped

- **Layout and framing:** the "Information note" (pp. 1), "Welcome to the European Company Survey 2019.", "Thank you for
  your participation …".
  - The Q25 diagrams "Example of 3 levels / Example of 4 levels" are pictures; only the question and "Indicate the
    number of hierarchical levels." are kept.
- **Administrative and contact block after Q74** (personal data and survey administration only):
  - [ER_NAME] and [ER_EMAIL], the name and email of the employee representative, with their intro text;
  - [REPORTMM1/2], the summary-report offer and retention of contact details;
  - [FOLLOWMM1/2], consent to share contact details and to follow-up contact.
- **Not part of this PDF:** the telephone screener and the employee-representative questionnaire (separate instrument).
  The other two wording versions (single independent companies, headquarters) are not published in this format.

## Doubts

- **License page:** now quoted verbatim (pasted by the user, 2026-10-28); `study.license` quotes the reuse conditions exactly.
- **Q2 and Q72 lists:** no category list is printed. The placeholder is kept as a choice, so these two variables are not
  usable as real NACE variables.
- **Number/percentage modelling** (two one-row grids per question) is our structure. On paper it is one question with
  two alternative fields. No printed rule says whether both may be filled, so neither field is made conditional on the
  other.
- **Q19 routing to Q23:** "None at all" or 0 routes to Q23. The expression is
  `not(${ictcomp_n} = 0 or ${ictcomp_pct} = '1')`.
- **Q49:** the bracket "GO TO QUESTION 50" covers the first eight options; only the last option routes to Q53.
- **Q69:** "GO TO QUESTION 70" covers the first three options.
- **Q24** has its own "This establishment does not use data analytics" option, although the filter already requires
  data analytics use. It is kept as printed.
- **Reference period:** the questions refer to "since the beginning of 2016" and "In 2018"; the fieldwork was in 2019
  (`timePrd` 2019).

## Titles (#64, 2026-10-28)

`titl` „European Company Survey 2019“, `subTitl` „Online questionnaire for Management respondents“, `altTitl` „ECS 2019“: page 1 of the questionnaire PDF (heading and „The European Company Survey 2019 (ECS 2019) …“).

## Group labels (formtransform 0.9.0, 2026-10-29)

None. Rule (user decision): `label_short` only for a printed scale or instrument name; no question numbers, no section designators. The questionnaire prints question numbers ("13. [MANAGER]"), but no scale names.
