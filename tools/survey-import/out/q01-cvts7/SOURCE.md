# q01-cvts7: CVTS7 Fragebogen

**Verdict: `ready`** (license: user decides at merge, see below)

## Source

- URL: https://www.statistik-nord.de/fileadmin/Dokumente/Erhebungen/Bildung-Rechtspflege/CVTS7_Fragebogen.pdf
  (retrieved 2026-10-27; 24 pages, text layer, no scanned pages; sha256 prefix 888563bdf18682d1)
- Citation: Statistische Ämter des Bundes und der Länder. (2026). Europäische Erhebung über die berufliche
  Weiterbildung in Unternehmen (CVTS7): Fragebogen [Erhebungsunterlage]. Statistikamt Nord.
- Title on the form: „Berufliche Weiterbildung in Unternehmen 2025“ (CVTS), reference year 2025.

## License / terms

- Statistikamt Nord imprint, section „Copyright“ (https://www.statistik-nord.de/impressum, retrieved 2026-10-27):
  > Soweit der Copyright-Vermerk bei dem jeweiligen Produkt/Internetseite keine ergänzenden Angaben enthält, gilt:
  > © Statistisches Amt für Hamburg und Schleswig-Holstein, Hamburg 2026
  > Sie können Informationen speichern, (auch auszugsweise) mit Quellenangabe weitergeben, vervielfältigen und
  > verbreiten sowie Links zur Homepage des Statistikamtes Nord speichern. Die Informationen dürfen im Übrigen nicht
  > verändert oder verfälscht werden.
  > […] Die ausführlichen Bestimmungen zum Copyright können Sie den Allgemeinen Geschäftsbedingungen (AGB) entnehmen.
- The PDF itself carries no separate copyright note.
- Official statistics survey form (Statistische Ämter des Bundes und der Länder; EU legal basis VO (EG) 1552/2005,
  198/2006), so § 5 UrhG may apply as well.
- Assessment: republication with attribution is allowed. The „nicht verändert oder verfälscht“ clause: wording is kept
  verbatim; only the format changes (paper → DDI). The AGB were not checked. The user decides at merge.

## Transcription

- 29 numbered questions (1.1–7.2) → 106 answer items + 10 notes = 116 survey rows; 135 DDI variables
  (select_multiple expanded). Build: no errors, no round-trip warnings (116 rows in, 116 out). Validation: **valid**.
- Names: `f<question>_<form field code>` (e.g. `f1_1_002a`, `f3_3_054a`, `f7_1_146a`); Ja/Nein grids
  `f2_3`, `f2_8`, `f3_3`, `f3_7`, `f5_1`, `f5_2`, `f6`, `f7_2`. Choice codes (1 = Ja, 2 = Nein, …) are ours; the
  paper form has none. Multi-select options use the field codes (`k021` …).
- Tables of number fields (1.1, 4.1, 4.2, 4.4): a `table-list` group whose label is the printed question (4.1: its printed heading „Teilnehmende an internen und externen Lehrveranstaltungen“); formtransform ≥ 0.8.0 writes it as a grid of numeric rows, each carrying the question. Each row is labelled with its verbatim row label; column headings (Anzahl, Personen / Teilnehmenden-Fälle, Prozent) are hints. The „davon: Geschlecht (nach Geburtenregister)“ sub-heading is a note row between the rows (`cdl:grid_note`). No text is joined across printed lines (2026-10-28).
- Questions with one answer field (1.2, 1.3): label = the printed question; the printed row label and column heading are the hint (e.g. „Geleistete Arbeitsstunden im Jahr 2025 insgesamt. Stunden“).
- Ja/Nein rows with an amount box on „Ja“ (3.6, 4.6): the amount item takes the printed „Falls „Ja“, …“ line (3.6) or its category text (4.6) as label, hint „Volle Euro“.
  (+ column label for 4.1 Personen / Teilnehmenden-Fälle). Units (Anzahl, Stunden, Volle Euro, Prozent) are hints.
- 3.6 and 4.6: Ja/Nein item plus a Euro amount item that is shown only for „Ja“.
- 3.3: the paper puts the share category (unter 10 % / 10 % bis unter 50 % / 50 % und mehr) next to each „Ja“ box.
  Here the five Ja/Nein items form one grid and the five share items a second grid (`f3_3_anteil`), each share
  item shown only if its activity is „Ja“.
- 2.4, 4.3, 4.5 „bis zu drei“: `select_multiple` with `count-selected(.) <= 3`.
- Routing, from the instructions on pages 6, 10, 11, 15, 16, 18:
  - 2.8 follow-ups only if 038 = Ja, then 039 = Ja.
  - 3.7 only if 3.6 Einnahmen (068) = Ja.
  - Teil 4 if 3.1 or 3.2 = Ja; 4.5 only if 3.2 = Ja.
  - Teil 5 if any of 3.1, 3.2, 3.3 = Ja; Teil 6 otherwise.
  - 5.1 methods only if 5.1 = „Ja, …“; 7.2 only if 7.1 = Ja.
  - The paper's „Weiter mit …“ notes are kept as hints too.

## Skipped

- Address box, „Identitätsnummer des Unternehmens“, and „Ansprechperson für Rückfragen“ (name, phone, fax, e-mail):
  administrative fields only (Hilfsmerkmale).
- Filling instructions with examples (p. 2), „Aufbau des Fragebogens“, „Vielen Dank“, the legal notice
  (§ 17 BStatG / DS-GVO, pp. 20–21) and the notes pages.
- The superscript references 1–13 to the „Erläuterungen“: these explanations are a separate sheet that is **not** in
  this PDF, so they are not transcribed and the reference numbers are removed from the text.

## Doubts

- The cover page says both „Erläuterungen zu 0 bis 0“ (template leftover) and „zu 1 bis 13“; the sheet was not found online.
- The 4.4 row „= Teilnahmestunden an Lehrveranstaltungen insgesamt 100 %“ is a fixed total on paper and is kept as
  hint text, not as an item.
- 3.3 grid split, see above (the item order differs from the paper layout).
- The Teil 6 condition „alle drei Fragen mit Nein“ is read as: 3.1, 3.2 and all five 3.3 activities = Nein.
- `gen-form.mjs` in this folder generates `form.json` (run it with the output path as its argument).
