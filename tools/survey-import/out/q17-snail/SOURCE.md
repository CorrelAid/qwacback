# q17-snail: SNAIL – Scale for the assessment of non-experts' AI literacy

**Verdict: `ready`**

## Sources (retrieved 2026-10-27)

- Article (open access), repository copy: https://hdl.handle.net/20.500.11811/13554 (bonndoc, Univ. Bonn; PDF `OA-2023-033.pdf`, publisher version). ScienceDirect (https://www.sciencedirect.com/science/article/pii/S2451958823000714) blocks scripted downloads; its Excel supplement (research data) could not be fetched and holds no item wording according to the article ("Research data will be published as Supplementary Material (Excel…)").
- **Table 1** (item wording, English) is an image in the PDF. Wording taken from a screenshot of Table 1 supplied by the user (`downloads/q17-snail/table1_user.png`; Fig. 4 path diagram: `fig4_user.png`).
- German version: Supplementary Material S1 of Laupichler, Aster, Perschewski & Schleiss (2023), Education Sciences 13(10), 978, https://doi.org/10.3390/educsci13100978. mdpi.com blocks scripted access (403); S1 was retrieved from the Wayback Machine copy of https://www.mdpi.com/2227-7102/13/10/978/s1 (snapshot 2024-01-17, zip with S1 PDF + S2 data set). Article PDF: https://hdl.handle.net/20.500.11811/13348.
- Cross-check source: Delphi study item table (Laupichler, Aster & Raupach 2023, Computers and Education: AI 4, 100126, https://doi.org/10.1016/j.caeai.2023.100126), as text in Laupichler's dissertation "Are they lit?" (https://hdl.handle.net/20.500.11811/12136, CC BY-NC-ND; used only for comparison).

## Citation

Laupichler, M. C., Aster, A., Haverkamp, N., & Raupach, T. (2023). Development of the "Scale for the assessment of non-experts' AI literacy" – An exploratory factor analysis. Computers in Human Behavior Reports, 12, 100338. https://doi.org/10.1016/j.chbr.2023.100338

German version: Laupichler, M. C., Aster, A., Perschewski, J.-O., & Schleiss, J. (2023). Evaluating AI courses: A valid and reliable instrument for assessing artificial-intelligence learning through comparative self-assessment. Education Sciences, 13(10), 978. https://doi.org/10.3390/educsci13100978

## License

- CHBR article, first page: "2451-9588/© 2023 The Authors. Published by Elsevier Ltd. This is an open access article under the CC BY license (http://creativecommons.org/licenses/by/4.0/)." Crossref also lists the VoR license as http://creativecommons.org/licenses/by/4.0/; bonndoc record: "Namensnennung 4.0 International", openAccess. The article adds: "Since SNAIL is freely available as an open access offering …".
- MDPI article (German version), first page: "Copyright: © 2023 by the authors. Licensee MDPI, Basel, Switzerland. This article is an open access article distributed under the terms and conditions of the Creative Commons Attribution (CC BY) license (https://creativecommons.org/licenses/by/4.0/)." S1 is the article's supplementary material; it carries no separate license (DOAJ lists the journal as CC BY-SA, OpenAlex/bonndoc as CC BY; the article itself says CC BY).

## Transcription

- 31 final items, V-number order, names `snail_v01` … `snail_v39` (gaps = eliminated V03, V07, V09, V10, V11, V26, V27, V38). One `field-list` group `snail`, all `select_one snail_agree7`.
- Base language Deutsch (de) from S1 ("German (as used in this study)"), English (en) from CHBR Table 1.
- Labels are full sentences "Ich kann …" / "I can …" + item, as the items were presented (S1 example: "Ich kann beschreiben, wie Machine Learning Modelle trainiert, validiert und getestet werden." / "I can describe how machine learning models are trained, validated, and tested."; CHBR: items "formulated as 'I can…' statements").
- Group label: scale name only (no instruction text is published for the original SNAIL; S1's instruction is specific to the retrospective course-evaluation version and was not used).
- Response scale: CHBR: "seven-point Likert scale from 'strongly disagree' (one) to 'strongly agree' (seven)". English labels 1 "strongly disagree", 2–6 numbers, 7 "strongly agree". German endpoint labels "Stimme überhaupt nicht zu" / "Stimme voll und ganz zu" are from S1 (see doubts).
- Not included: the two bogus/attention-check items (not published), sociodemographics, the S1 retrospective double rating, ATAI items, course questions.

### Factor membership (CHBR Table 1 / Table 2)

| Factor | Items |
|---|---|
| Technical Understanding (TU), 14 | V04, V05, V12, V13, V14, V15, V16, V17, V18, V22, V23, V30, V33, V36 |
| Critical Appraisal (CA), 10 | V06, V08, V20, V21, V25, V28, V32, V34, V35, V39 |
| Practical Application (PA), 7 | V01, V02, V19, V24, V29, V31, V37 |

Matches the German S1 factor lists (Technisches Verständnis / Kritische Beurteilung / Praktische Anwendung).

### Wording cross-check

- English Table 1 (image) = English column of MDPI S1 (text) for all 31 items, character for character (checked by script).
- Delphi table (CAEAI 2023): same wording except V04 has no final period, and typographic quotes (’) where Table 1 uses straight quotes ('…'). Table 1 kept.
- V17 has no final period in both Table 1 and S1 ("explain how deep learning relates to machine learning"); kept verbatim.
- German strings checked against S1 text by script; V24 "KI-unterstütztes" is hyphenated at a line break in S1 (layout extraction shows "KI-"), kept with hyphen.

## Validation

`build-codebook.mjs`: no errors, round trip 31 rows in / 31 out. `validate.sh`: **valid**.

## Doubts

- German response labels come from S1, whose scale is **six**-point (1 "Stimme überhaupt nicht zu" … 6 "Stimme voll und ganz zu"; English translation there: "Strongly disagree" / "Strongly agree"). They are used here as endpoints of the original seven-point scale. Alternative: number-only German labels as in q13.
- The German version was made for a course-evaluation adaptation (current + retrospective self-assessment); the item wordings themselves are unchanged statements and are labelled by the authors as the German SNAIL items.
- S1 obtained via the Wayback Machine (mdpi.com blocks scripted access); the content is the authors' supplement as published.

## Revision 2026-10-27 (user decision: only the literal text from the latest source)
- The latest source with item text is MDPI S1 (Laupichler, Aster, Perschewski & Schleiss 2023, *Education Sciences* 13(10), 978). The BMC 2024 study (Laupichler et al., *BMC Medical Education* 24, 401) used SNAIL but prints no item or scale text.
- Answer scale switched to S1's **6-point** scale (list `snail_agree6`). The endpoint labels are literal from S1: 1 = "Stimme überhaupt nicht zu" / "Strongly disagree", 6 = "Stimme voll und ganz zu" / "Strongly agree". S1 leaves points 2–5 unlabelled, so they carry their code number only.
- The original SNAIL paper (2023, CHBR) uses a 7-point scale; this codebook follows S1 instead.
- All 31 items were rechecked character by character against S1: German against the right column, English against the left column without the "(Vnn)" codes, each item read as "Ich kann …" / "I can …" plus the column text, as in S1's example item. All match. The script's apparent mismatches (V15, V21, V28, V33: page-break column merge; V24: "KI-" line break) were checked by hand.

## Titles (#64, 2026-10-28)

`titl` „Scale for the assessment of non-experts’ AI literacy“, `altTitl` „SNAIL“: CHBR article, „“Scale for the assessment of non-experts’ AI literacy” (SNAIL)“ (curly apostrophe as printed).
