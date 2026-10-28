# q20-eu-ict-ent-ai: EU survey on ICT usage and e-commerce in enterprises 2025, Module E "Artificial Intelligence"

**Verdict: `ready`** (CC BY 4.0, stated in the manual itself; user confirms at merge)

## Source

- **Text source: `KS-01-25-065-EN-N.pdf`**, Eurostat, "European businesses statistics compilers' manual for ICT usage and
  e-commerce in enterprises, 2025 survey", 2025 edition, Manuals and guidelines, catalogue number KS-01-25-065.
  - Published 17 December 2025; manuscript completed November 2025. Publications Office of the European Union.
  - Page: https://ec.europa.eu/eurostat/web/products-manuals-and-guidelines/w/ks-01-25-065
  - PDF: https://ec.europa.eu/eurostat/documents/3859598/22759373/KS-01-25-065-EN-N.pdf/b8659ccb-6ca6-1ea0-d9cd-a1321b021a77?version=1.0&t=1765876089895&download=true
  - Retrieved 2026-10-27. sha256 572087602ae862e781a4a2b7c7e1637065aba963b67a6e35268d83b0094d43b6.
- **Part used:** Annex 1 – Model questionnaire, which prints "ICT-ENT 2025 - Model Questionnaire V1.5"
  ("MODEL QUESTIONNAIRE VERSION 1.5"). The part is Module E "Artificial Intelligence" (manual pp. 275–277), plus the AI
  entries of Annex 1's glossary (pp. 280 ff.).
- **Why Annex 1, not the explanatory notes.** Within the same manual, Section 2.3.2.6 ("Module E: Artificial
  Intelligence", explanatory notes) quotes the questions with different wording:
  - "artificial intelligence" in lower case;
  - a newer AI definition ("gather, use and/or produce data", "systems that create content (generative AI)",
    "software-based");
  - "The costs seemed too high".

  Annex 1 is the questionnaire itself, so all wording, numbering and routing come from Annex 1 alone. The notes were
  read only to confirm scopes and the mandatory/optional status.
- **Citation:** Eurostat. (2025, December 17). European businesses statistics compilers' manual for ICT usage and
  e-commerce in enterprises, 2025 survey (2025 edition; Manuals and guidelines, KS-01-25-065). Publications Office of
  the European Union. https://ec.europa.eu/eurostat/web/products-manuals-and-guidelines/w/ks-01-25-065
- **Population, as printed in Annex 1:**
  - Sampling unit: "Enterprise".
  - NACE Rev. 2 sections C, D, E, F, G, H, I, J, L, M, N and group 95.1.
  - "Enterprises with 10 or more employees and self-employed persons"; optional 0–9.
  - Module E: "(Scope: enterprises with access to the internet, i.e. if A1>0)".
  - Reference: the current situation (survey period in 2025). Recommended survey period: first quarter 2025.

### Why the 2026 model questionnaire was not used (user decision)

"ICT-ENT 2026 - Model Questionnaire V1.1.pdf" is newer, from the CIRCABC folder "ICT enterprise survey MQ 2026",
uploaded 2026-10-09 (sha256 eac31057…b98a). Its own metadata table says "Status: Draft" and "Rights: Restricted access
to members of the TF and the ISS WG". The user therefore chose the latest officially published source, the 2025 manual.
The 2026 PDF is still in `downloads/` and is used only for the comparison below. None of its text is in `form.json`.

## License / terms

- The manual's copyright page (p. 2):
  > © European Union, 2025
  > The reuse policy of European Commission documents is implemented based on Commission Decision 2011/833/EU of
  > 12 December 2011 on the reuse of Commission documents (OJ L 330, 14.12.2011, p. 39).
  > http://data.europa.eu/eli/dec/2011/833/oj) Except otherwise noted, the reuse of this document is authorised under a
  > Creative Commons Attribution 4.0 International (CC-BY 4.0) licence https://creativecommons.org/licenses/by/4.0/).
  > This means that reuse is allowed provided appropriate credit is given and any changes are indicated.
  - The only third-party element it names is the cover photo (© Gorodenkoff/AdobeStock), which is not used.
- The Eurostat "Copyright notice and free re-use of data" (https://ec.europa.eu/eurostat/help/copyright-notice) is quoted
  as in q19. It includes "When reuse involves translations of publications or modifications to the data or text, this
  must be stated clearly to the end user of the information. A disclaimer regarding the non-responsibility of Eurostat
  shall be included."
- `study.license` quotes both and states the changes:
  - a format conversion of Module E (Annex 1) to DDI, wording unchanged;
  - other modules, explanatory notes and layout omitted;
  - glossary definitions attached to the E1 items as hints.

  It adds the q19 disclaimer: "this DDI version was made by the qwac project (CorrelAid), not by Eurostat; Eurostat bears
  no responsibility for it."

## What the AI module contains (2025, Annex 1)

The module has 5 questions, E1–E5, with 29 answer items: 15 mandatory (E1 a–h, E2 a–g) and 14 optional (E3 a–e, E4,
E5 a–h). This matches Annex 1's own count: "E Artificial Intelligence 15 / 14".

| Q | Content | Who gets it | Mand./opt. |
|---|---|---|---|
| E1 a–h | Use of 8 AI technologies (text mining, speech recognition, natural language generation, generating pictures/videos/audio, image recognition, machine learning, AI-based robotic process automation, autonomous machines) | internet users (A1 > 0) | mandatory |
| E2 a–g | Purposes: marketing/sales, production/service processes, administration/management, logistics, ICT security, accounting/finance, R&D/innovation | AI users (any E1 = Yes) | mandatory |
| E3 a–e | How AI was acquired: developed by own employees; commercial software modified by own employees; open-source modified by own employees; ready-to-use commercial software purchased; external providers contracted | AI users | optional |
| E4 | Ever considered using AI? (filter question) | non-users (all E1 = No) | optional |
| E5 a–h | Reasons for not using AI: costs, lack of expertise, incompatibility, data availability/quality, data protection and privacy, unclear legal consequences, ethics, not useful | E4 = Yes | optional |

**Against the user's description:**

- Asks whether the enterprise uses AI and which technologies: **yes** (E1).
- Asks for purposes: **yes** (E2). It is mandatory in 2025.
- If not, why: **yes, with a restriction.** E5 goes only to non-users who answered E4 "ever considered using AI" with
  Yes. Non-users who never considered AI get no reasons question. E4 and E5 are both optional.
- The reasons the user named are all there:
  - E5f "Lack of clarity about the legal consequences (e.g. liability in case of damage caused by the use of
    Artificial Intelligence)";
  - E5e data protection and privacy;
  - E5b lack of expertise;
  - E5a costs;
  - E5c incompatibility;
  - E5g ethics.
- Two further reasons that the user did not name: E5d (data availability/quality) and E5h (not useful).
- **What the module does not contain:**
  - There is no question on why AI use is restricted or limited among enterprises that use AI.
  - There is no main or "biggest" hindrance; each reason is a separate Yes/No item.
  - There is no "other" reason.
  - There is no question on AI skills, training or literacy. Skills appear only as reason E5b, and indirectly in E3 a–c
    (development or modification by own employees).
  - There is nothing on the AI Act, AI policies, personal data or bias.

**Mapping to the user's topics** (✔ = direct, (✔) = indirect):

| Q | AI use | reasons for not using AI | legal uncertainty / data protection | skills |
|---|---|---|---|---|
| E1 | ✔ (technologies) | – | – | – |
| E2 | ✔ (purposes) | – | – | – |
| E3 | ✔ (sourcing) | – | – | (✔) a–c: development/modification by own employees |
| E4 | ✔ (considered use, non-users) | filter for E5 | – | – |
| E5 | – | ✔ a–h | ✔ e (data protection), f (legal consequences), g (ethics) | ✔ b (lack of expertise) |

## Differences from the 2026 draft (ICT-ENT 2026 V1.1, comparison only)

- **Size:** 2026 has 8 questions (E1–E8) and 31 items, 13 mandatory and 18 optional. 2025 has 5 questions and 29 items,
  15 mandatory and 14 optional.
- **Reasons for not using AI: the same.** 2025 E4/E5 correspond to 2026 E7/E8. They have the same question texts, the
  same eight reasons a–h with identical wording, and the same filter (all E1 = No → considered? → Yes → reasons). Both
  are optional in both years.
  - Small differences: the 2025 routing sentence is "E1 a) to h) = “No” then go to E4 (if optional included) else go to
    F1"; the instruction reads "Questions E4 is presented …"; and 2025 writes "– Optional" with a capital O.
- **E2 (purposes):** mandatory in 2025, optional in 2026. 2026 adds two finance-sector examples (robo-advisor, credit
  scoring), which come with a footnote. In 2025, E2c reads "risk assessment based on" and E2f ends "used for bookkeeping
  tasks"; 2026 has "risk assessment, based on" and "for bookkeeping documents".
- **Acquisition:**
  - 2025 E3 is optional, with 5 options that mix sourcing and modification.
  - 2026 E3 is mandatory, with 4 options (own employees / external providers / open-source / closed-source with
    examples OpenLLaMA, ChatGPT), plus a separate mandatory E4 "Has the AI software been modified …".
- **New in 2026, not in 2025:** E5 (processing data on individuals with AI) and E6 (measures to check AI results for
  bias).
- **E1:** 2025 c) reads "generating written or spoken language"; 2026 adds "programming codes". 2025 b)–h) write "AI
  Technologies" (capital T); 2026 uses lower case.
- **Intro definition:**
  - 2025 Annex 1: "to gather and/or use data to predict, recommend or decide … the best action to achieve specific
    goals", "purely software based", examples include "machine translation software".
  - 2026: "gather, use and/or produce data …", "generative AI", without machine translation.

## Transcription

- **Rule:** every label, hint and choice label is literal text from Annex 1 of KS-01-25-065:
  - Module E pp. 275–277;
  - the E1 item hints from Annex 1's glossary: Text mining, Speech recognition, Natural language generation, Machine
    learning, Robotic process automation (Artificial Intelligence based).
- PDF layout hyphenation is undone. For example, "tech-/nologies" becomes "technologies" and "deci-/sions" becomes
  "decisions".
- **Kept as printed:**
  - "How did you enterprise acquire …" (E3, sic);
  - "Questions E4 is presented …" (sic);
  - the capitalisation "AI Technologies";
  - the routing sentence "E1 a) to h) = “No” then go to E4 (if optional included) else go to F1." (E1 hint).
- **Structure:**
  - Group `module_e`: the module title, with the scope line as hint. A1 is in Module A and is not transcribed, so it is
    not a `relevant`.
  - `e_intro` (note): the printed AI definition with its examples.
  - E1, E2, E3 and E5 are `field-list` grids of `select_one yes_no` rows, keeping the item letters "a)"…"h)".
  - E2's "some of the examples may be:" lists are item hints.
  - "– Optional" and "(Filter question)" are hints.
- **Codes:** Annex 1 prints only Yes/No tick boxes. The codes 1 = Yes and 0 = No are the manual's own transmission codes
  for the AI_* variables ("1 Yes", "0 No"). Names are `e1_a`…`e5_h` and `e4`, following the 2025 numbering.
- **Routing → `relevant`:**
  - E2 and E3: any E1 = Yes. This is also the scope in Section 2.3.2.6.
  - E4: all E1 = No ("presented only to respondents who answered ‘No’ to E1a)-h)").
  - E5: E4 = Yes ("No -> go to F1").
- **Count:** 5 questions, 29 answer items, 1 note and 5 groups, giving 40 form rows. Choice list: `yes_no` (2).
  - Build: no errors and no warnings. The round trip keeps 30 survey rows in and 30 out.
  - Validation (`validate.sh`): **valid**.

## Character-by-character check

- **How it works.** `verify-labels.py` runs `pdftotext -raw` on the manual and keeps only Annex 1: from "MODEL
  QUESTIONNAIRE VERSION 1.5" up to the start of Annex 2, including Annex 1's glossary. Section 2.3 and the transmission
  tables are excluded. It removes the page footers and the tick-box glyph and collapses whitespace.
- **Hyphens at line ends.** Each line-end hyphen is marked. When matching, the mark may be skipped (layout hyphenation)
  or read as a real hyphen. Nothing else is normalised: quotes, dashes, bullets and case must match exactly.
- **Run** (from the repo root):
  `python3 -I tools/survey-import/out/q20-eu-ict-ent-ai/verify-labels.py tools/survey-import/out/q20-eu-ict-ent-ai/form.json tools/survey-import/downloads/q20-eu-ict-ent-ai/KS-01-25-065-EN-N.pdf`
- **Result (2026-10-27): 83 text lines from 54 texts checked, 0 not found.**
- **Negative test.** Six errors were planted in a copy of the form, and 5 of them were flagged:
  - the 2026 E1c wording "or programming codes";
  - "software-based" from Section 2.3;
  - the silent typo fix "How did your enterprise";
  - the lower-case "artificial intelligence" in E4, as in Section 2.3;
  - "The costs seemed too high", as in Section 2.3.

  The sixth, "deci-sions" (a layout hyphen kept), was **not** flagged. A hyphen at a line-end position cannot be told
  apart from a real one in the PDF text, so this is a known limit. All line-end breaks in the labels were checked by
  hand: none is a real compound.
- **Other limits.** A line is found anywhere in Annex 1, not necessarily under the same question. Short texts ("Yes",
  "– Optional") pass trivially. The pairing was checked by reading Annex 1 against `gen-form.mjs`.

## Dropped / not included

- All other modules of the 2025 questionnaire: A, B, C, D, F (ICT and the environment) and X. Also the optional Web
  purchases module.
- Section 2.3.2.6 explanatory notes (definitions per answer option, Wikipedia references) and the data transmission
  tables, which are not questionnaire text.
- **Dependency outside the section:** Module E applies only if A1 > 0 (internet access). The scope line is kept as the
  group hint.

## Doubts

- The manual is internally inconsistent: the Section 2.3 notes and Annex 1 differ in the AI definition, capitalisation
  and E5a ("seemed" vs "seem"). Only Annex 1 was used.
- The study title says 2025 and cites the manual (2025 survey). The questionnaire it reproduces is V1.5, whose own
  version date is not printed in Annex 1.
- E3–E5 are optional. Countries may omit them, so not every national 2025 survey asks for the reasons for not using AI.

## German version (added 2026-10-28)

- Source: Statistisches Bundesamt (Destatis), Qualitätsbericht "IKT-Erhebung in Unternehmen 2025", questionnaire section E (pp. 6–8, PDF pp. 22–24): https://www.destatis.de/DE/Methoden/Qualitaet/Qualitaetsberichte/Unternehmen/ikt-unternehmen-2025.pdf?__blob=publicationFile&v=2 (retrieved 2026-10-28; in downloads/q20-eu-ict-ent-ai/destatis/).
- License (PDF imprint): "© Statistisches Bundesamt (Destatis), 2025. Vervielfältigung und Verbreitung, auch auszugsweise, mit Quellenangabe gestattet."
- Mapping by content: Destatis E1.1–1.8 = Eurostat E1 a–h; E2.1–2.7 = E2 a–g; E3 = Eurostat E4; E4.1–4.8 = Eurostat E5 a–h. Destatis does not ask Eurostat E3 (acquisition), so E3 has no German text. Destatis E1.3 says "oder Programmiercodes" where Annex 1's E1 c) has no "or programming codes" (the manual's variable table has it); both kept as printed.
- Kept as printed: "Inkompatibillität", "Systeme,die", "Gesichtserkennungsysteme", "Tiefenlernen /Deep Learning", "Pick-and Pack-Lösungen", the full stop after E4.8.
- Removed: soft hyphens, line-end hyphenation, footnote numbers 16–18, item numbers, and the routing lines ("Falls die Fragen E1.1 bis E1.8 mit „Nein" beantwortet wurden, weiter mit Frage E3.", "Bitte beantworten Sie die Frage E3 nur dann, …"), whose numbers follow Destatis' numbering; the English routing hints and `relevant` stay.
- The E1 glossary hints and the module scope hint have no German counterpart and stay English only.
- Check: verify-labels-de.py, 65 German lines, 0 not found; three planted errors were all caught.


## Titles (#64, 2026-10-28)

`titl` „EU Survey on ICT Usage and e-Commerce in Enterprises 2025“ (Annex 1 heading „EU SURVEY ON ICT USAGE AND E-COMMERCE IN ENTERPRISES 2025“), `subTitl` „Module E: Artificial Intelligence“ (Annex 1 module heading), `altTitl` „ICT-ENT 2025“ (Annex 1, „ICT-ENT 2025 - Model Questionnaire V1.5“).
