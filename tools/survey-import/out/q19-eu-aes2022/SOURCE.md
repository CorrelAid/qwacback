# q19-eu-aes2022: EU Adult Education Survey 2022, model questions (AES manual v1.3)

**Verdict: `ready`** (CC BY 4.0 / Eurostat reuse policy; user confirms at merge)

## Source

- **Text source (latest version): `1 - 2022 AES MANUAL V1.3.docx`**: Eurostat, 2022 AES manual, Version 1.3,
  30 March 2023, Section 3 "Explanatory notes", per variable "Model question", "Filter", "Codes or values".
  Copied from the user's Downloads (sha256 4e2d597f…dfdff). The same file is in the CIRCABC folder "AES 2022"
  (https://circabc.europa.eu/w/browse/a81f03f5-b8e6-4b84-a93d-ef847ef85db0, node b0e7dcc6-3990-426a-a7b6-4a242c0ddeaa).
- **Structure only: `2b - 2022 AES QUESTIONNAIRE v1.2 (2022-08-09).xls`**: Section 2, suggested European questionnaire.
  It is older than the manual, so none of its wording is used. Retrieved 2026-10-27 as guest from CIRCABC:
  https://circabc.europa.eu/d/a/workspace/SpacesStore/4a0cd68c-f359-4e01-9560-233e9088e132/2b%20-%202022%20AES%20QUESTIONNAIRE%20v1.2%20(2022-08-09).xls
  (sha256 b7f63875…6741).
- `2a - 2022 AES MANUAL - Annexes v1.3.1 (2023-06-26).pdf` (CIRCABC): only checked for copyright notes.
- Citation: Eurostat. (2023). 2022 AES manual: Methodological guidelines for the adult education survey (Version 1.3,
  30 March 2023). European Commission, Eurostat, Unit F-3.
- Context: legal basis Regulation (EU) 2019/1700 and Commission Implementing Regulation (EU) 2021/861. Population:
  people aged 18–69 in private households. Fieldwork: July 2022–March 2023.

## License / terms

- Eurostat "Copyright notice and free re-use of data" (https://ec.europa.eu/eurostat/help/copyright-notice, retrieved 2026-10-27):
  > The copyright for the editorial content of this website, which is owned by the EU, is licensed under the Creative
  > Commons Attribution 4.0 International licence. This means that you can re-use the content provided you acknowledge
  > the source and indicate any changes you have made. […]
  > Reuse of statistical data, metadata, publications, and other dissemination tools published on this website for
  > commercial or non-commercial purposes is authorised provided the source is acknowledged. The reuse policy of the
  > European Commission is implemented by the Decision of 12 December 2011.
  > These general principles can be subject to conditions which may be specified in individual copyright notices. […]
  > When reuse involves translations of publications or modifications to the data or text, this must be stated clearly
  > to the end user of the information. A disclaimer regarding the non-responsibility of Eurostat shall be included.
- The Decision of 12 December 2011 is Commission Decision 2011/833/EU.
- The notice prescribes disclaimer wording only for translations, not for modifications. `study.license` therefore
  quotes the requirement exactly and adds our own sentence: "Disclaimer: this DDI version was made by the qwac project
  (CorrelAid), not by Eurostat; Eurostat bears no responsibility for it." It also states the change: **a format
  conversion to DDI**, wording unchanged, some items omitted (listed below).
- The manual, the annexes and the Excel carry no copyright note or restriction of their own.
- The files sit on CIRCABC, not on the Eurostat website that the notice covers. They are Eurostat documents, so the
  Commission reuse policy applies.
- GALI originates from the Euro-REVES project (manual). It is a standardised EU item, and no separate notice was found.

## Transcription

- **Rule:** every label, hint and choice label is literal text from manual v1.3, Section 3. Nothing comes from the
  Excel and nothing is our own wording.
  - Question labels come from "Model question".
  - Choice labels come from "Codes or values".
  - Hints come from instruction sentences printed under "Model question" or "Filter", or from code-table labels such as
    "Total number of instruction hours" and "In euros".
  - Group labels are the manual's "Detailed topic" entries from each variable's specification table. Example:
    "Most recent formal education activity - details (12 months)".
- Placeholders are kept exactly as printed: `<FEDNAME>`, `<NFEACTxx_NAME>`, `<NFERAND1_NAME>`, `<<month, year>>`,
  `<any of the following formal education or training programmes>`, `[net]`, `[1st cut-off value]`.
- **Count:** 113 questions plus 1 note, in 22 groups, giving 184 DDI variables. Types: 70 select_one, 11 select_multiple,
  14 integer, 16 text, 2 decimal.
  - Build: no errors and no warnings. The round trip keeps all 114 rows (114 in, 114 out).
  - Validation: **valid**.
  - The previous Excel-based version had 133 questions.
- Names are the manual's variable identifiers in lower case. Choice codes are the manual's codes.
- **Filters → `relevant`:**
  - All filters are taken from the manual's "Filter" fields: EMP12M, the job block (MAINSTAT = 10), PERMJOB,
    DROPEDUCLEVEL, GUIDESOURCE/GUIDEINTER, FEDNUM, FED details (FEDNUM ≥ 1), FEDFIELD, FEDREASONMAIN, FEDPAIDBY,
    FEDOUTCOMEMAIN, NFEGUIDEJT (MAINSTAT = 10 or EMP12M = 1), NFENUM, the NFE blocks, NFEINITIA1, NFEPAIDBY1,
    NFEPAIDVAL1, NONEED, DIFFTYPE, DIFFMAIN, INFPURP, LANGBEST1, LANGLEVEL1 and GALI QUESTION_2.
  - Derived variables used in filters (GUIDE, NFE, INF, WANT, NONEED, LANGUSED) are written out over the items they come from.
  - Two filters apply to single answer options: "Not working at that time" for FEDWORKTIME and NFEACTxx_WORKTIME, and the
    job-related reasons of NFEREASON1. These are hints quoting the manual's filter text, not choice filters.
  - The exclusive "No outcome yet" rule is a constraint, with the manual sentence as hint.
- **Where the Excel was used (structure only):**
  - Item order. The manual lists variables by topic, so the Excel's questionnaire order is used. HHINCOME is moved to the
    end, as the Excel recommends.
  - The routing condition for WANT_B (no FED and no NFE).
  - Splitting LOCSIZEFIRM "1-9" into a category plus a follow-up exact number.
  - Grouping NFEACTxx and NFERAND1 into separate blocks.
- **Item modelling:**
  - WANT_A and WANT_B are select_one items with "Yes"/"No", coded to the WANT codes the manual gives (2/1 and 4/3).
  - DIFFTYPE has two printed wordings: `difftype` (WANT = 2 or 4) and `difftype_noneed` (NONEED = 2). They share one
    list, and DIFFMAIN covers both.
  - INF is a single checklist `select_multiple`. The manual codes each item "1 if selected, 2 if not selected" and
    recommends a single checklist. Both suggested introductions are kept as a note.
  - GUIDE_1–3 are printed in the model question as three separate questions, so they are three select_one items with
    the manual's own GUIDE codes: 1 = "selected", 2 = "not selected" (from "Each variable is coded: 1 if selected,
    2 if not selected"). The question texts are the model question's GUIDE_1/2/3 wording.
  - The "Country of …" codes of CITIZEN, BIRTHPLACE, BIRTHFATHER and BIRTHMOTHER, the RESTIME "01-69" code and the
    LOCSIZEFIRM "1-9" code each get a follow-up field. Its label is the manual's code label, e.g.
    "Country of birth (SCL GEO alpha-2 code); see annex 6".
  - RESTIME keeps the manual's native-born / foreign-born instruction as one label. Country of birth is a free code,
    so that routing cannot be expressed.
  - Alternative wordings printed as "should be asked in situations …" (BIRTHPLACE, BIRTHFATHER, BIRTHMOTHER) are hints.
- **Classifications are not inlined.** JOBISCO, LOCNACE, countries and languages are text fields whose label or hint
  is the manual's code label naming the classification (ISCO-08, NACE Rev. 2, SCL GEO alpha-2, ISO 639 alpha-3).
  ISCED levels and fields are inlined because the manual lists them.

## Character-by-character check

- `verify-labels.py` converts the manual docx to HTML with pandoc and collects every paragraph and table cell as plain
  text. Each label, hint and choice label of `form.json` is split at line breaks, and every line must occur in that
  text exactly. Only runs of whitespace are collapsed. Quotes, dashes, case and punctuation are not normalised.
- Run: `python3 -I tools/survey-import/out/q19-eu-aes2022/verify-labels.py tools/survey-import/out/q19-eu-aes2022/form.json "tools/survey-import/downloads/q19-eu-aes2022/1 - 2022 AES MANUAL V1.3.docx"`
- Result, 2026-10-27 (after the GUIDE fix): **481 text lines from 453 texts checked, 0 not found.**
- Negative test: three Excel v1.2 wordings were put back into a copy of the form, and all 5 resulting lines were flagged.
  The wordings were "getting a job, or changing", "Do no know precisely but less …" and "(FEDNAME) start".
- Limit: the check finds each line anywhere in the manual, not necessarily in the same variable's section. Short labels
  such as "Yes" or "Health" pass trivially. The pairing of text with variable was checked by reading the dump against
  the generator.

## Dropped

- **No model question in the manual** (the "Model question" field is only guidance or "-"):
  - HATLEVEL, HATFIELD, HATYEAR;
  - NFERAND1 and NFERAND1_TYPE ("-").
- **Not in the manual at all** (they were Excel-only items):
  - FEDNAME, NFEACTxx_NAME, NFERAND1_NAME;
  - NFECOURSENUM, NFEWORKSHOPNUM, NFEGUIDEDJTNUM, NFELESSONNUM ("How many?");
  - FEDNBWEEKS, FEDDURPERWEEK, NFENBWEEKS1, NFEDURPERWEEK1 (optional helpers);
  - LANGBEST2, LANGLEVEL2.
- **AGE:** the manual's model question is "What is your year of birth? Have you already had your birthday this year?".
  It is transcribed as BIRTHYEAR and BIRTHPASS. AGE itself is derived and has no question. The alternative "What is
  your date of birth?" is not included.
- **Texts dropped:**
  - the Excel's intro and instruction texts and its section titles (not in the manual);
  - the manual's "Please note" paragraphs addressed to countries (DSS notes; national adaptation of FEDMAINSTAT,
    FEDLEVEL, DROPEDUC, LANGUSED, PERMJOB; the GALI background);
  - the manual's description of the HHINCOME model question. The quoted questions and bands are kept.
- **Technical variables:** COUNTRY, REGION, DEG_URB, REFYEAR, REFMONTH, REFDAY, RESPID, INTMETHOD, INTPART.
- **Derived global variables** without their own question: NFE, INF, GUIDE, LANGUSED (number of languages), WANT,
  NONEED, and the transmitted HHINCOME quintiles 1–5.
- **Codes not offered as answers:** "Not stated" (-1), plain "Not applicable" (-2) and "No answer (WANT = -1)". These
  are transmission codes. "Not applicable (father unknown)" and "Not applicable (mother unknown)" are kept because they
  are substantive answers. The "Not stated." line in the HHINCOME band list is dropped for the same reason.
- **Activity loops:** the per-activity block NFEACTxx (for up to 5 activities) is transcribed once, with variable names
  `nfeact01_*` and labels keeping `<NFEACTxx_NAME>`. The details block for the randomly selected activities is
  transcribed once, for the 1st activity (NFERAND1); the manual only defines that one. NFEINITIA1's filter
  "NFEACTxx_PURP = 1" is bound to `nfeact01_purp`.
- **MARSTADEFACTO:** the footnote marker "1" after "partner" is dropped from the choice labels. The footnote text
  "Husband/wife/civil partner/partner/cohabitee" is the hint.

## Doubts

- GUIDE_1–3 use the manual's codes "selected / not selected" as choice labels on three separate questions. A single
  checklist with the code-table labels would need a stem question that the manual does not print, so it was not used.
- The two NFE blocks (NFEACTxx and NFERAND1) share the group label "Non-formal education activities - details (12 months)",
  because that is the manual's detailed topic for both.
- `study.license` quotes the notice's disclaimer requirement, but the disclaimer sentence itself is ours. The notice
  gives no wording for modified reuse.
