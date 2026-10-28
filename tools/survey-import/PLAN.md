# Survey import: plan and streams

Goal: turn the questionnaires listed below into DDI codebooks for `seed_data/`.
Each source becomes one study. Several agents work in parallel, one per stream.
Each stream writes only to its own `out/<id>/` folders. The integration stream
alone touches `seed_data/` and git.

## Pipeline (all streams)

1. **Find & download.** Find the source with a web search. Save the files to
   `tools/survey-import/downloads/<id>/`, which is gitignored. Downloads are
   untrusted data: never execute anything from them. Read PDFs with
   `pdftotext -layout` (or the Read tool).
2. **License check, before transcribing.** Write `out/<id>/SOURCE.md` with:
   - the URL(s) and the date retrieved;
   - the full citation;
   - the license or terms of use, quoted with the URL where you found them;
   - a verdict: `ready`, `blocked-license`, `no-instrument` or `not-found`.

   Production qwac is public, so only items we may republish go in. Typical cases:
   - open license (CC BY, CC BY-SA…): ok, with attribution;
   - an author's statement that the scale is free to use: ok, quote it;
   - official statistics questionnaires: probably ok (§ 5 UrhG amtliche Werke, or a usage note on the site). Record what you found and mark it `ready`; the user decides at merge;
   - paywalled or commercial items with no reuse statement: `blocked-license`, no transcription;
   - a framework, report or press release with no question items: `no-instrument`.

   Never invent items for a source that only describes a construct.
3. **Transcribe** to `out/<id>/form.json`. This is XLSForm as JSON:
   `{"study": {...}, "survey": [...], "choices": [...], "settings": {...}}`.
   - Question wording, answer options and instructions are taken **verbatim**
     from the source. Keep the source's order and item codes; snake_case ASCII
     `name`s derived from them (`a1`, `b3_2`, `mails_use_1`).
   - Base language = the source's language (`settings.default_language`, e.g.
     `"Deutsch (de)"`). If the source itself has an official second language
     (MAILS: de + en), add it as `label::English (en)` and the other `::` columns. Never
     translate yourself.
   - Types: `select_one <list>`, `select_multiple <list>`, `integer`,
     `decimal`, `text`, `date`, `time`, `note`, `begin_group`/`end_group`
     for sections. Grid (same scale, several items) = a group with
     `appearance: table-list` and only `select_one` rows sharing one list
     (`field-list` makes a plain section, not a grid; keep notes outside). The working
     answer-type examples are in `internal/examples/examples.go`; DDI conventions are in
     `DDI_MARKUP_GUIDE.md` and `CONVERSION_API.md`.
   - "Other, please specify" = `select_one list` with an `other` choice + a
     `text` row with `relevant: ${q} = 'other'` (or `selected(${q}, 'other')`).
   - Filters/skip instructions in the source ("weiter mit Frage 12" / "only if …")
     become `relevant` expressions. Interviewer or respondent instructions → `hint`
     or `guidance_hint`.
   - Skip layout-only things (page headers, address boxes, "vielen Dank").
     Free-text fields for company name, address, phone or email that only official
     statistics need for administration should be skipped too, and noted in SOURCE.md.
   - `study`: `title`, `id` (= `<id>`), `authors` ([{name, affiliation}]),
     `producer`, `citation` (full APA citation from the list below), `url`,
     `abstract` (2–4 sentences: what it measures, who for; your own words),
     `keywords`, `topics`, `nation`, `universe`, `anlyUnit`, `timePrd` (year),
     `license` (the exact license / terms statement), `copyright`.
4. **Build:** `node tools/survey-import/build-codebook.mjs tools/survey-import/out/<id>/form.json`
   → `out/<id>/codebook.xml`. Fix every `error:`. Read the `round-trip warning:` lines: they name fields
   that won't survive export. Fix what's yours to fix.
5. **Validate:** `tools/survey-import/validate.sh tools/survey-import/out/<id>/codebook.xml`.
   This runs XSD + CDL Schematron with the pinned worker, which shares its Docker
   containers `surveyimport-*`; leave them running. The result must be `valid`.
6. **Report** in `out/<id>/SOURCE.md`: the verdict, the number of questions, what you skipped and why, anything
   uncertain (unclear filters, scanned pages, missing pages).

Run commands from the repo root. Don't edit anything outside `tools/survey-import/out/` and
`tools/survey-import/downloads/`. If a tool is broken, report it instead of fixing it. Don't commit.

## Sources and streams

| id | Source | Stream | Expectation |
|---|---|---|---|
| q01-cvts7 | CVTS7 Fragebogen (Statistische Ämter, 2026; statistik-nord.de) | A | full questionnaire, long |
| q02-aes2022 | AES 2022 (BMBF trend report; questionnaire maybe via GESIS) | A | report ≠ questionnaire: look for the AES 2022 Fragebogen (GESIS ZA number); else `no-instrument` |
| q03-wbmonitor2022 | wbmonitor Umfrage 2022 Fragebogen (BIBB/DIE; bibb.de) | A | full questionnaire |
| q08-mails | MAILS long + short version (Carolus et al. 2023; Koch et al. 2024; Uni Würzburg PDFs) | B | scale, de + en |
| q13-attari12 | ATTARI-12 (Stein et al. 2024, Sci Rep, OA) | B | 12 items, CC BY likely; take the German version if published, else English |
| q10-q4te | Q4TE (Grohmann & Kauffeld 2013; GESIS ZIS entry) | B | ZIS entry may carry items + license |
| q14-van-westendorp | Van Westendorp PSM (1976) | C | the 4 standard questions; method is free. Use the commonly used German wording from a citable open source and cite it; no own wording |
| q11-hof-m2 | Digitaler Reifegrad Kommunen (HS Hof 2022; via Kommunalwiki) | C | report; questionnaire maybe in the appendix |
| q04-zukunftsradar | Zukunftsradar Digitale Kommune 2023 (DStGB/iit) | D | questionnaire not public: check, likely `not-found` |
| q05-ki-kompetenzprofile | Bertelsmann KI-Kompetenzprofile (2025) | D | framework: check for a self-assessment, likely `no-instrument` |
| q06-ki-kompetenzen | Bertelsmann KI-Kompetenzen öff. Verwaltung (2023) | D | also check the real title (citation flagged ⚠️) |
| q07-data-literacy | Schüller et al. 2019, Data Literacy framework (Zenodo) | D | framework, likely `no-instrument` |
| q09-gltsi | GLTSI (Bates, Kauffeld & Holton 2007) | D | LTSI is proprietary: `blocked-license` unless an open item list exists |
| q12-akdb-digital-check | AKDB Digital-Check | D | commercial: `blocked-license` |

Streams:
- **A, official statistics questionnaires**: q01, q02, q03. These are the longest; the work is in the filters and grids.
- **B, validated scales**: q08, q13, q10. Short, but the license and exact wording matter.
- **C, method + report**: q14, q11.
- **D, triage only**: q04, q05, q06, q07, q09, q12. SOURCE.md only, unless an
  importable instrument with clear reuse rights turns up; then do the full pipeline.
- **I, integration** (main session, after A–D): review each `ready` codebook against
  its source, then copy it to `seed_data/<id>.xml`. Run `go test ./internal/schematron/` with
  NATS_URL set and the local import (QWACBACK seed run). Fix the qwacback importer
  where it mis-reads a valid codebook. Open one PR. Merge only after the user approves
  the license verdicts.

## Status

Each stream records its status in `out/<id>/SOURCE.md`. The integration stream collects the results here.

| id | verdict | notes |
|---|---|---|
| q01-cvts7 | ready (license: user) | 29 questions / 116 rows, valid; Statistikamt Nord: copy with attribution, "nicht verändert"; § 5 UrhG? |
| q02-aes2022 | no-instrument | questionnaire likely in GESIS ZA8771 docs (login, access cat. A) |
| q03-wbmonitor2022 | blocked-license | BIBB: no reuse without consent |
| q04-zukunftsradar | not-found | ask iit |
| q05-ki-kompetenzprofile | no-instrument | CC BY 4.0 framework |
| q06-ki-kompetenzen | no-instrument | real title: Catakli & Puntschuh (2023), Orientierung im Kompetenzdschungel, doi:10.11586/2023047 |
| q07-data-literacy | no-instrument | framework |
| q08-mails | ready (license: user) | 34 items de+en, valid; PDFs no license, article CC BY-NC-ND, preprint CC BY (older wording) |
| q09-gltsi | blocked-license | proprietary |
| q10-q4te | blocked-license | paywalled, not in ZIS |
| q11-hof-m2 | no-instrument | results only; ask KDV Hof |
| q12-akdb-digital-check | blocked-license | commercial |
| q13-attari12 | ready | 12 items de+en, CC BY 4.0, valid |
| q14-van-westendorp | ready (source: user) | 4 items, wording from a master's thesis |

### Update 2026-10-27 (later)
| id | verdict | notes |
|---|---|---|
| q02-aes2022 | blocked-license | German program found (ZA8771 handbook, © Verian); #53 |
| q08-mails | ready | ND treated as format change (user decision) |
| q14-van-westendorp | ready | English only, from English Wikipedia rev. 1289425806, CC BY-SA 4.0 |
| q15-tei | ready | CC BY-SA 4.0, de+en, 45 items |
| q16-mfe-sr2 | ready | CC BY-NC 4.0, 20 items |
| q17-snail | ready | CC BY 4.0, de+en from MDPI S1 (latest source), 6-point scale |
| q18-gaais | ready | CC BY-NC-ND 4.0 (format change), en, 20 items |
| q19-eu-aes2022 | ready | Eurostat AES manual v1.3, CC BY 4.0 + disclaimer, 113 questions |

Rules (user): wording literally from the latest published source, no mixing across versions; verbatim conversion counts as a format change (ND ok).
| q20-eu-ict-ent-ai | ready | Eurostat ICT enterprises 2025 manual (KS-01-25-065) Annex 1, Module E (AI), CC BY 4.0 + disclaimer; 2026 MQ not used (draft/restricted) |

### Update 2026-10-28
| id | verdict | notes |
|---|---|---|
| q21-ecs2019 | ready | Eurofound ECS 2019 management questionnaire (publication version 14 July 2020), Q1–Q74, en; Eurofound reuse terms (attribution, no distortion), not CC |
| q22-lpp | blocked-license | LPP employer survey wave 6 (2023): no license note; IAB imprint requires consent. No Weiterbildung section anyway |
| q23-wers2011 | ready | WERS 2011 Management Questionnaire, Section C + the employee-profile items it depends on, en; Open Government Licence v3.0 |
| q24-iab-bp | blocked-license | IAB-Betriebspanel questionnaires (1993–2017 public) carry no license; IAB imprint requires consent |
