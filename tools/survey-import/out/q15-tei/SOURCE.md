# q15-tei: TEI – Training Evaluations Inventar

**Verdict: `ready`**

45 items, one question group, one 5-point agreement scale; Deutsch (de) base, English (en). `codebook.xml` validates (XSD + CDL Schematron): `valid`.

## Sources (retrieved 2026-10-27)

- Info page (Open Test Archive): https://www.testarchiv.eu/de/test/9007914
- Inventar TEI (Deutsch), 6 pp.: https://pada.psycharchives.org/bitstream/8b4aab6c-7f2a-4a89-88e5-ef925367269d → `downloads/q15-tei/tei-de.pdf` (used for the German text)
- Inventar TEI (Englisch), 8 pp.: https://pada.psycharchives.org/bitstream/30dc5343-92a3-4b88-b8a8-8002146c6de2 → `downloads/q15-tei/tei-en.pdf` (used for the English text; also contains the English scoring table)
- Verfahrensdokumentation TEI, 33 pp.: https://pada.psycharchives.org/bitstream/6283764a-2f69-455b-9fc6-45625525681f → `downloads/q15-tei/tei-doku.pdf` (item codes and dimensions, Tabelle 1)
- The DOI https://doi.org/10.23668/psycharchives.3454 returned HTTP 502 (psycharchives.org gateway) at retrieval; the PDFs came from the download links on the testarchiv.eu page.

## Citation

Ritzmann, S., Hagemann, V. & Kluge, A. (2020). TEI. Training Evaluations Inventar [Verfahrensdokumentation, Fragebogen Deutsch und Englisch]. In Leibniz-Institut für Psychologie (ZPID) (Hrsg.), Open Test Archive. Trier: ZPID. https://doi.org/10.23668/psycharchives.3454

Note: the testarchiv.eu page and the two questionnaire PDFs cite the current version, https://doi.org/10.23668/psycharchives.4493 ("[Verfahrensdokumentation, Inventar deutsch und englisch]"); the page lists `psycharchives.3454` as "Version 1". The Verfahrensdokumentation PDF still cites `.3454`. The study citation uses `.3454` as given by the user.

## License

testarchiv.eu info page, "Copyright/Lizenz": "Copyright Autor(-Innen); CC-BY-SA 4.0".

Questionnaire PDFs, p. 1 (https://www.testarchiv.eu/de/test/9007914):

> Verpflichtungserklärung
> Bei dem Testverfahren handelt es sich um ein Forschungsinstrument, das der Forschung, Lehre und Praxis dient. Es wird vom Testarchiv online und kostenlos zur Verfügung gestellt und ist urheberrechtlich geschützt, d. h. das Urheberrecht liegt weiterhin bei den AutorInnen.
> Mit der Nutzung des Verfahrens verpflichte ich mich, die Bedingungen der Creative Commons Lizenz CC BY-SA 4.0 zu beachten. Ich werde nach Abschluss meiner mit dem Verfahren zusammenhängenden Arbeiten mittels des Rückmeldeformulars die TestautorInnen über den Einsatz des Verfahrens und den damit erzielten Ergebnissen informieren.

> Terms of use
> The test instrument is a research instrument that serves research, teaching and practice. It is made available online and free of charge by the test archive and is protected by copyright, i.e. the copyright remains with the author(s).
> By using this test, I agree to abide by the terms of the Creative Commons License CC BY-SA 4.0. After completion of my work with the measure, I will inform the test authors about the use of the measure and the results I have obtained by means of the feedback form.

Consequence: our codebook is a derivative (adapted/re-formatted) of the TEI and is shared under **CC BY-SA 4.0** as well, with attribution to the authors (set in `study.license`). Users of the instrument are asked to send feedback to the authors via the Rückmeldeformular after use. License confirmed by the user from the PDF.

## Transcription

- Order = the questionnaire form (not the scoring table / Tabelle 1). Names `tei_01` … `tei_45` by form position (the form has no item numbers).
- One `field-list` group `tei` (label = form title "Training Evaluations Inventar (TEI)" / "Training-Evaluation-Inventory (TEI)"), all items `select_one tei_agree5`.
- The adaptation instruction ("Vor der Durchführung müssen die durch eckige Klammern markierten Bereiche …" / "Before conducting the training, the fields in the questionnaire marked by square brackets must be adapted …") is a `note` row `tei_hinweis` (exported as `preQTxt` of `tei_01`). A `guidance_hint` on the group was tried first but is silently dropped by the converter (no round-trip warning), hence the note.
- Bracketed placeholders kept exactly as printed, e.g. "[BEZEICHNUNG AUSBILDUNGSGEFÄSS – z.B. Training]", "[THEMA/INHALT – z.B. Teamarbeit]", "[ARBEITSFELD – z.B. Fliegen]", "[DESCRIPTION OF COURSE TYPE – e.g. training]", "[FIELD OF WORK – e.g. aviation]" (en dash, "z.B." without space as in the form).
- Scale as printed in the form header, codes 1–5: Stimme ich nicht zu / Stimme ich eher nicht zu / Weder noch / Stimme ich eher zu / Stimme ich sehr zu — I do not agree / I rather not agree / Neither nor / I rather agree / I agree very much.
- Line-break hyphenation removed only where the word is hyphenated anyway ("day-to-\nday" → "day-to-day", item 39 en; confirmed by the English scoring table).
- Every label was checked programmatically as a whitespace-normalised substring of the PDF text, in form order.
- Nothing skipped except the cover page, the Verpflichtungserklärung, and (en PDF) the evaluation guide and reference tables.

## Dimensions (from Verfahrensdokumentation Tabelle 1 and the en PDF "Items for each dimension of TEI"; not in the form)

Training outcome (Trainingsoutcome-Dimensionen):

| Dimension | Items (form position, doc code) |
|---|---|
| Subjektiver Spaß / Subjective fun | 1 Spaß1, 2 Spaß2, 3 Spaß3 |
| Wahrgenommene Nützlichkeit / Perceived usefulness | 5 Nutzen1, 6 Nutzen2, 7 Nutzen3, 8 Nutzen4 |
| Wahrgenommene Schwierigkeit / Perceived difficulty | 9 Schwer1, 10 Schwer2, 11 Schwer3, 12 Schwer4 |
| Subjektiver Wissenszuwachs / Subjective knowledge growth | 13 Wissen1, 14 Wissen2, 18 Wissen3 |
| Einstellung gegenüber dem Training / Attitude towards training | 16 Einst1, 17 Einst2, 15 Einst3 |

Training design (Trainingsdesign-Dimensionen):

| Dimension | Items (form position, doc code) |
|---|---|
| Problembasiertes Lernen / Problem-based learning | 19 Prob1, 22 Prob2, 24 Prob3, 26 Prob4, 29 Prob5, 33 Prob6 |
| Aktivierung (von Vorwissen) / Activation | 20 Aktiv1, 28 Aktiv2, 27 Aktiv3, 23 Aktiv4, 39 Aktiv5 |
| Demonstration | 21 Dem1, 30 Dem2, 31 Dem3, 25 Dem4, 4 Dem5, 32 Dem6, 37 Dem7 |
| Anwendung / Application | 34 Anw1, 35 Anw2, 36 Anw3, 44 Anw4, 40 Anw5 |
| Integration | 41 Int1, 43 Int2, 42 Int3, 45 Int4, 38 Int5 |

Scoring: mean per dimension, interpreted separately; the documents mention no reverse-coded items. For "Wahrgenommene Schwierigkeit" the items are worded as comprehensibility (high = easy), although the manual says difficulty "should be kept low".

## Doubts

- Typos kept verbatim from the German form: item 6 "[BEZEICHNUN AUSBILDUNGSGEFÄSS – z.B. Training]" (missing G; visible in the rendered PDF), item 38 lacks the final full stop. Item 24 "selbständige" (form) vs. "selbstständige" (manual), item 45 "einzelner Themen" (form) vs. "der einzelnen Themen" (manual); the form wording is used.
- Scale endpoint: the manual says 1 = „stimme überhaupt nicht zu“, the form prints "Stimme ich nicht zu"; the form label is used.
- Author affiliations are not stated in the documents, so `affiliation` is left empty.
- DOI version (.3454 v1 vs. .4493 current), see Citation.

## Titles (#64, 2026-10-28)

`titl` „Training Evaluations Inventar“, `altTitl` „TEI“: cover of the German form (tei-de.pdf) and the Verfahrensdokumentation, „TEI / Training Evaluations Inventar“.

## Group labels (formtransform 0.9.0, 2026-10-29)

`label_short` `TEI` (de + en) on the grid: cover of tei-de.pdf and tei-en.pdf, „TEI / Training Evaluations Inventar“ / „TEI / Training-Evaluation-Inventory“.
