# q23-wers2011: Workplace Employment Relations Study 2011 (WERS6), Management Questionnaire (MQ)

**Verdict: `ready`** (Open Government Licence v3.0 via GOV.UK; user confirms at merge). **Partial import**: Section C
plus the EPQ items it depends on, see "Scope".

## Source

- **Text source: `mq.pdf`**: "Sixth Workplace Employment Relations Study (WERS6) Management Questionnaire (MQ)
  Published: March 2011". It is ref. BIS/11/803, 112 PDF pages (printed "Page N of 111"), with a text layer and no
  scanned pages. PDF metadata: author "Department for Business, Innovation and Skills", modified 2012-12-10.
  - Publication page: https://www.gov.uk/government/publications/the-2011-workplace-employment-relations-study-wers
    (BIS; first published 23 January 2013, last updated 28 July 2014).
  - PDF: https://assets.publishing.service.gov.uk/media/5a74f4a0e5274a59fa716449/11-803-wers6-sixth-workplace-study-management-questionnaire.pdf
  - Retrieved 2026-10-28. sha256 d906316dc768a5e3a15fb790cd4e6b377cfc03a352b1f04af947d1744c15d930.
- The separate paper Employee Profile Questionnaire, `bis-11-801-…-employee-profile-questionnaire.pdf` (BIS/11/801,
  sha256 2fc368fe…49f3), was downloaded but **not used**. The EPQ items here come from the "SECTION EPQ" of the MQ PDF,
  so the text has a single source.
- Printed on the cover: WERS sponsors BIS, Acas, ESRC, UKCES and NIESR (NIESR "Through funding from the Nuffield
  Foundation"). The fieldwork contractor was the National Centre for Social Research.
- **Citation:** Department for Business, Innovation and Skills. (2011). Sixth Workplace Employment Relations Study
  (WERS6): Management questionnaire (MQ) (BIS/11/803). GOV.UK.
  https://www.gov.uk/government/publications/the-2011-workplace-employment-relations-study-wers
- **Latest source:** this is the only published version of the WERS 2011 MQ. WERS has not been repeated since 2011.
  The UK Data Service copy (SN 7226 documentation) was not used and its terms were not checked, because the GOV.UK copy
  is openly published under OGL.

## License / terms

- GOV.UK publication page, footer (retrieved 2026-10-28, URL above):
  > All content is available under the Open Government Licence v3.0, except where otherwise stated
  > © Crown copyright
- The MQ PDF itself has **no** copyright or licence statement and does not state otherwise. The page lists no
  third-party rights for the document.
- OGL v3.0 (https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/), retrieved 2026-10-28:
  > You are free to: copy, publish, distribute and transmit the Information; adapt the Information; exploit the
  > Information commercially and non-commercially […]
  > If the Information Provider does not provide a specific attribution statement, you must use the following:
  > Contains public sector information licensed under the Open Government Licence v3.0.
  - It excludes "third party rights the Information Provider is not authorised to license" and personal data. It also
    has a non-endorsement clause.
- `study.license` quotes the footer, gives the OGL attribution statement, and states the changes:
  - a format conversion of Section C and EPQ items to DDI, wording unchanged;
  - all other sections omitted.

  It adds a non-endorsement disclaimer (made by qwac/CorrelAid, not by BIS or the other sponsors).
- **Possible doubt for the user:** the questionnaire was co-developed by five sponsors and fielded by NatCen. GOV.UK
  publishes it under OGL without exception, but the PDF itself carries no Crown copyright line.

## Scope (lower priority, so a partial import as instructed)

The MQ is long: 111 printed pages across Sections EPQ and A–M. Transcribed **completely**:

- **Section C "RECRUITMENT, TRAINING AND ORGANISATION OF WORK"** (printed pp. 30–37): all 28 printed items CFILLVAC …
  CCOMPUTT, plus the TRINTRO note.
- **The EPQ items Section C depends on** (printed pp. 4–5):
  - INTROEPQ (an instruction, see below).
  - ZALLEMPS (total employees). The soft checks of COTHJOBT, COTHDOT, CTEAMST and CCOMPUTT refer to it.
  - SOC_1 to SOC_9, the male/female occupation table. The text fill [LOGIDX] (largest non-managerial occupational
    group) is computed from it, and most of Section C asks about that group.
  - ZSOCDESC, the description of that group.

Section C's routing is entirely internal (CLOGVAC, COFFJOB/COFFCHK, COTHJOB, COTHDO, CTEAMS, CCOMPUT). No filter in
Section C depends on Section A or other sections.

## Transcription

- **Count:**
  - 26 questions: 3 in the EPQ part (ZALLEMPS, the SOC table, ZSOCDESC) and 23 in Section C. Each grid counts as one
    question.
  - 50 answer rows, 1 note and 5 groups, making 61 form rows and 92 DDI variables (select_multiple expanded).
  - 14 choice lists with 95 choices.
  - Build: no errors and no warnings; the round trip keeps 51 survey rows in and 51 out. `validate.sh`: **valid**.
- **Names:** the SPSS variable names in lower case: `cfillvac`, `coffjob`, `zmgr_male` … Exceptions:
  - The "other, please specify" companions are `cfactors_other` (printed CFACTOTH) and `cchanx_other` (printed COTHCHN),
    because the DDI convention requires `<name>_other`. Their "Other" choice code is `other`, not the printed 9 or 14.
  - The groups are `section_epq`, `section_c`, `soc`, `card_c9_grid` and `cteamho` (our own names).
- **Choice codes:** the printed SPSS values (1…, 97). Choice labels are the long label column. Where the PDF gives only
  one word ("Skills", "Age", "Yes", "Some", "None"), that word is the label; the PDF's key says it "is a sufficient
  label". The short value names (IntOnly, Refer, …) are not used.
- **Interviewer text goes to `guidance_hint`**, verbatim. This covers:
  - SHOW CARD lines;
  - INTERVIEWER, PROMPT and PROBE lines;
  - CODE ALL THAT APPLY;
  - ENTER NUMBER;
  - SOFT CHECK lines.

  Groups lose a guidance_hint in the DDI conversion, so two group-level instructions sit on the rows:
  - INTROEPQ ("INTERVIEWER: To start the interview, please enter the relevant data from the EPQ.") is the first line of
    ZALLEMPS's guidance_hint.
  - The SOC table's "THIS IS Q2 …"/"PROMPT: …" lines are on each of the 18 cells.
- **Ranges and hard checks become `constraint`:**
  - ZALLEMPS: `. >= 5 and . <= 999997`, with the printed hard-check message as `constraint_message`.
  - Each SOC cell: `<= ${zallemps}`, with the printed message.
  - COTHJOBT, COTHDOT, CTEAMST and CCOMPUTT: 1..999997.
  - CCOMPUT: 0..997.
  - ZALLEMPS: `required` ("NO DON'T KNOW, NO REFUSAL, NO EMPTY").
- **Not encoded:** the SOC hard checks on row, column and table totals (cross-cell sums). Soft checks are not
  constraints (the interviewer can override them); their text is in `guidance_hint`.
- **Grids** (`table-list`):
  - **SOC table:** 18 integer rows, row by row (male, then female). The row label is the printed occupation, and the
    column heading "Male"/"Female" is the hint (q01 precedent).
  - **CVARIETY/CDISCRETE/CONTROL/CDESIGN** (SHOW CARD C9, shared scale): the group label is the stem as printed at
    CVARIETY. The rows are the printed "...variety in their work?" etc., kept with " :" where printed. The other three
    repeat the stem in parentheses, which is not copied.
  - **CTEAMHOA/C/D** (Yes/No): the group label is the stem printed at CTEAMHOA, "Do each of the following statements
    apply …". The rows are the printed "...Team …" statements. *Note:* CTEAMHOC and CTEAMHOD repeat a **different**
    stem in parentheses: "(Which, if any, of the following statements apply …)". It is not carried over. The item
    names jump from A to C to D as printed; there is no CTEAMHOB.
- **Multi-line question texts** are kept as separate lines (`\n`): CLOGVAC (lead-in + question), ZSOCDESC and CCOMPUT.
  PDF line wraps within one sentence are joined with a space. Nothing is joined with dashes.
- **Text fills** are kept literally, e.g. "[LOGIDX: TITLE OF THE LARGEST OCCUPATIONAL GROUP]". ZSOCDESC prints "[LOGIDX:
  TITLE OF LARGEST OCCUPATIONAL GROUP]" without "THE". No computed variable was added.
- **Exclusive choices** (`exclusive: yes`): CSPECIX "None of these", CATESTW/CPTESTW "No tests conducted", CHOW "None
  of these".
- **Routing → `relevant`:**
  - CCHANX: CLOGVAC = Yes.
  - COFFCHK: "If don't know whether any training (COFFJOB=DK)". "Don't know" is a CAPI key, not a printed code, so it
    is encoded as COFFJOB unanswered: `string-length(${coffjob}) = 0`.
  - CTRAIN and CHOW: "COFFJOB=1 to 100% (1-6), or COFFCHK=Yes (1)", encoded as
    `(string-length(${coffjob}) > 0 and ${coffjob} != '7') or ${coffchk} = '1'`.
  - COTHJOBT / COTHDOT / CTEAMST: the preceding question = 97. CCOMPUTT: CCOMPUT = 997.
  - CTEAMHO grid: CTEAMS <> None, i.e. `${cteams} != '7'`.
  - The two "other" text rows: `selected(…, 'other')`.
- **Kept as printed (sic):**
  - "Did this training cover any of the matters listed on this card?:"
  - "Just a few (1-19%," and "INTERVIEWER: Exact number to be entered" (no full stop) on SHOW CARD C10 (CTEAMS). This
    is why it is a separate list from SHOW CARD C8.
  - "No tests conducted." with a full stop (CPTESTW) vs "No tests conducted" (CATESTW). These are two lists.
  - "INTERVIEWER: THIS IS Q1 IN THE PAPER EPQ" without a full stop.
  - "Other (please specify CFACTOTH)" and "Some other way (please specify COTHCHN)", including the variable names, as
    the choice label and the companion text label.
  - The 97 option's label "INTERVIEWER: Exact number to be entered." is the printed label text.
- **Italics** (= "response was not printed on the associated show card"), found with `pdftohtml -xml`:
  - CSPECIX 9 "None of these";
  - CHOW 12 "None of these";
  - the "97) Number" option of COTHJOB, COTHDO and CTEAMS.

  DDI has no place for this flag. It is recorded here only.

## Character-by-character check

- `verify-labels.py` runs `pdftotext -raw` on the MQ PDF. Its corpus is only the transcribed parts:
  - "SECTION EPQ …" up to ZFULLM (pp. 4–5);
  - "SECTION C …" up to "SECTION D …" (pp. 30–37).

  It removes the "Page N of 111" footers and joins lines with a space. A line-end "letter-" hyphen may be skipped or
  kept. Nothing else is normalised: quotes (’ “ ” '), dashes (–), ellipses ("..."), the ≠ sign and case must match
  exactly.
- It checks every label, hint, guidance_hint, constraint_message and choice label, line by line. EPQ texts are checked
  against the EPQ part; Section C texts and all choice labels against Section C.
- **Run** (from the repo root):
  `python3 -I tools/survey-import/out/q23-wers2011/verify-labels.py tools/survey-import/out/q23-wers2011/form.json tools/survey-import/downloads/q23-wers2011/mq.pdf`
- **Result (2026-10-28): 269 text lines from 230 texts checked, 0 not found.**
- **Negative test.** Six errors were planted in a copy, and 5 were flagged:
  - a curly apostrophe "other’s";
  - "organized";
  - "Customer service / liaison";
  - a hyphen for the en dash in "Internet – external website(s)";
  - an extra full stop after "THIS IS Q1 IN THE PAPER EPQ".

  The sixth, a full stop added to the CTEAMS instruction, was **not** flagged. The same sentence with a full stop is
  printed at COTHJOB, and a line may match anywhere in its part. This is a known limit.
- The pairing of texts with questions and choice codes was checked by reading pp. 4–5 and 30–37 against
  `gen-form.mjs`.

## Left out

- **Rest of Section EPQ** (pp. 5–17): ZFULLM/ZFULLO/ZPRTM/ZPRTO (full/part-time), nationality (ZNUKNAT …), age bands,
  ethnicity, and the further EPQ items. Section C does not use them.
- **Sections A, B and D–M** (pp. 18–29, 38–111): workplace and organisation characteristics; management of personnel;
  consultation; representation; pay; disputes; grievance; fair treatment; flexibility; performance; change; general
  information. Section C does not depend on them.
- **Training-related items outside Section C** (not transcribed; listed for a possible later extension):
  - BYOURJOB and BCONSULT, option "Training of employees";
  - EULRNUM (union learning reps);
  - ETRAINA and ETRAINB, option "Training of employees";
  - FRISE, option "Acquiring job-related skills";
  - FAPPTRN ("Does the performance appraisal result in an evaluation of employees' training needs?");
  - ITRAIN (training of health and safety representatives);
  - JNONEMP and JCONIN, option "Training";
  - JWHYTEM, option "Obtain specialist skills";
  - KREQUIRE, option "Training provision";
  - KHOWMOX, option "Workforce training";
  - LRECIMP, option "Reduction in training expenditure".
- **Front matter:** "Key to using this document" (pp. 2–3), the table of contents and the cover.
- **Computed variables:** ZTOTMEN, ZTOTWOM, Z*_TOT, LOGTOT and LOGIDX. They are documented, not asked. LOGIDX remains
  as the literal text fill.
- **The paper EPQ** (BIS/11/801) and the Financial Performance, Worker Representative and Employee questionnaires.

## Doubts

- The CTEAMHO grid uses the CTEAMHOA stem; CTEAMHOC/D print a different stem in parentheses (see above).
- The COFFCHK "DK" filter is encoded as "COFFJOB unanswered". A form built from this has no explicit "Don't know"
  option. Adding one would invent a choice, so none was added.
- Labels contain the CAPI text-fill placeholder "[LOGIDX: …]". A reuser has to replace it with the occupational group
  of interest.
- `timePrd` is 2011. Fieldwork ran March 2011 to June 2012 (GOV.UK page).
- `gen-form.mjs` in this folder generates `form.json` (run it with the output path as its argument).

## Titles (#64, 2026-10-28)

`titl` unchanged. `altTitl` „WERS6“: cover of the MQ, „Sixth Workplace Employment Relations Study (WERS6)“.

## Group labels (formtransform 0.9.0, 2026-10-29)

None. Rule (user decision): `label_short` only for a printed scale or instrument name; no question numbers, no section designators. The MQ prints section headings ("SECTION C") and variable codes (CFACTORS), but no scale names.
