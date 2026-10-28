# q22-lpp: Linked Personnel Panel (LPP), Arbeitgeberbefragung (Betriebsbefragung)

**Verdict: `blocked-license`**. Nothing is transcribed and there is no form.json. A CC BY-SA route via the data report exists (see "Alternative"); the user decides.

## Source (newest employer questionnaire)

- Questionnaire collection (German), FDZ page "Linked Personnel Panel (LPP) – Version 1224 v1":
  https://fdz.iab.de/int_bd_pd/linked-personnel-panel-lpp-version-1224-v1/ →
  https://doku.iab.de/fdz/LPP/lpp_frageboegen_de.zip (retrieved 2026-10-28; zip entries dated 2025-11-19).
  - Used: `LPP_1224_Frageboegen_de/LPP_Fragebogen_Betriebsbefragung_Welle_6_DE.pdf` (19 pages, text layer,
    QuarkXPress, created 2023-08-01; sha256 prefix cee4b2c92bed2995).
  - Title on the form: „Zusatzbefragung zum IAB-Betriebspanel zu Personalarbeit und Unternehmenskultur –
    Arbeitgeberbefragung 2023 – im Auftrag der Bundesagentur für Arbeit“; field institute Kantar Public,
    „Projekt 334410601 · Juni 2023 – f2f“. **Wave 6 (LPP 1224), fieldwork 2023**: the newest wave.
  - An official English version exists (`lpp_questionnaires_en.zip`, `…Betriebsbefragung_Welle_6_EN.pdf`) under the same (absent) terms.
- Data documentation: Ruf, K., Berger, V., Wolter, S., & Grunau, P. (2025). *LPP – Linked Personnel Panel 1224:
  Arbeitsqualität und wirtschaftlicher Erfolg: Längsschnittstudie in deutschen Betrieben (Datendokumentation der
  sechsten Welle)* (FDZ-Datenreport 07/2025). FDZ der BA im IAB. https://doi.org/10.5164/IAB.FDZD.2507.de.v1,
  PDF https://doku.iab.de/fdz/reporte/2025/DR_07-25.pdf (published 2025-08-19, 215 pp.; retrieved 2026-10-28; sha256 prefix 8c01ba3276f8d09f).
- Panel citation: Ruf, K., Mackeben, J., Grunau, P., & Wolter, S. (2020). A unique employer-employee study: The Linked
  Personnel Panel (LPP) – Design, extensions and research potential. *Jahrbücher für Nationalökonomie und Statistik,
  240*(1), 133–145. https://doi.org/10.1515/jbnst-2018-0045
- Dataset: IAB LPP1224 v1, DOI 10.5164/IAB.LPP1224.de.en.v1

## License / terms

- The questionnaire PDF (DE and EN) has **no** copyright or license note. The zip holds only the PDFs. The FDZ dataset
  page gives no terms for the questionnaires either.
- FDZ imprint, „Urheber- und Kennzeichenrecht“ (https://fdz.iab.de/impressum/, retrieved 2026-10-28):
  > Alle vom FDZ als Herausgeber selbst erstellte Publikationen und Grafiken sind, sofern dort jeweils vermerkt, unter
  > folgender Creative-Commons-Lizenz veröffentlicht: https://creativecommons.org/licenses/by-sa/4.0/deed.de. Sie dürfen
  > dieser Lizenz entsprechend verwendet werden. In allen anderen Fällen ist eine Vervielfältigung und Verwendung in
  > anderen elektronischen oder gedruckten Publikationen ohne ausdrückliche Zustimmung des IAB als Herausgeber nicht gestattet.
  The IAB imprint (https://iab.de/impressum/) has the same wording.
- FDZ-Datenreport 07/2025, imprint page:
  > Nutzungsrechte
  > Diese Publikation ist unter folgender Creative-Commons-Lizenz veröffentlicht:
  > Namensnennung - Weitergabe unter gleichen Bedingungen 4.0 International (CC BY-SA 4.0)
  > https://creativecommons.org/licenses/by-sa/4.0/deed.de
- Assessment: the questionnaire carries no CC note, so the „In allen anderen Fällen … nicht gestattet“ clause applies.
  Republishing it needs the IAB's express consent. The form was also produced by a commercial field institute (Kantar
  Public, now Verian). § 5 UrhG does not clearly apply: this is a research survey, not an official statistic. → `blocked-license`.

## Alternative (user decision): CC BY-SA 4.0 data report

FDZ-Datenreport 07/2025, chapter 6 („Variablenbeschreibung: LPP-Betriebsbefragung 1224“, pp. 44 ff.), prints for every
employer variable the question text („Frage“) and the answer codes („Ausprägungen“), under CC BY-SA 4.0. An import
could take the wording **only** from this report. It would not be a faithful copy of the questionnaire:
- Item letters are re-lettered. The questionnaire's Kleinbetriebe Q17 has „a) Mit Führungskräften“ and „c) Mit
  Beschäftigten ohne Führungsverantwortung“; the report has „A:“ and „B:“.
- Number fields are rendered as „… und zwar mit __“.
- Respondent and interviewer instructions („Mehrere Angaben möglich!“, „Falls genaue Angabe nicht möglich, bitte
  schätzen!“) and the routing („weiter mit Frage …“, filter column F04=1 …) are largely missing. Filters would have to come
  from the unlicensed questionnaire, which mixes sources.
- Text may differ from the form. In 7 places (e.g. f80) the report text extracts as „Ihre Antwortabstufen“ where the form has „Ihre Antwort abstufen“. pdftotext gives this in all modes; it may still be a spacing artifact, not checked visually.
- Scale endpoints are summarised („1 = Gar nicht aktiv bis 5 = Sehr aktiv“).
This route was not taken: the rule requires explicit permission for the questionnaire itself and no mixing of sources.
If the user accepts the report as the source (CC BY-SA, share-alike), the employer part can be imported from it.

## Content note: no training section

The wave 6 employer questionnaire has **no Weiterbildung section**. Organisation-level training is collected in the
IAB-Betriebspanel, and the LPP employer survey runs as a supplement to it. The only training-related items in wave 6:
- Block „Personalentwicklung“ (Q16–Q29): Mitarbeitergespräche, Zielvereinbarungen, Entwicklungspläne,
  Leistungsbeurteilung, Umgang mit Leistungsschwäche (incl. „Personalentwicklungsmaßnahmen werden gezielt angeboten …“).
- Q60: Maßnahmen (z. B. Schulungen oder Sensibilisierungen) gegen sexuelle Belästigung.
- Q80 G: Engagement des Betriebsrats bei „Weiterbildungen und Umschulungen“ (1–5).
Older waves had training items in the technology block: wave 4 (2018/19) Q74–75 and wave 5 (2021) Q78–79, on training
for new technologies and the share of the training budget it took. Under the latest-source rule these do not count as
the current instrument.

## Transcription

None (blocked-license). Questions/rows: 0.

## Doubts

- Whether the IAB would grant consent on request. Contact for the questionnaire: FDZ / Philipp Grunau (IAB), not contacted.
- The FDZ-Datenreport abstract mentions the „fifth wave“ although the report covers wave 6. This looks like an editorial slip and does not affect the verdict.
