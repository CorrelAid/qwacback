# q08-mails: MAILS – Meta AI Literacy Scale

**Verdict: `ready`** (license caveat below; user decides at merge)

## Sources (retrieved 2026-10-27)

- Project page: https://hci.uni-wuerzburg.de/research/MAILS/
- Long version, German: https://downloads.hci.informatik.uni-wuerzburg.de/MAILS-questionnaire-german.pdf (used, base language)
- Long version, English: https://downloads.hci.informatik.uni-wuerzburg.de/MAILS-qustionnaire-english.pdf (used, second language)
- Short version, German/English: https://downloads.hci.informatik.uni-wuerzburg.de/MAILS-short-questionnaire-german.pdf, https://downloads.hci.informatik.uni-wuerzburg.de/MAILS-short-questionnaire-english.pdf (for the item mapping below)
- Published article (author copy): https://downloads.hci.informatik.uni-wuerzburg.de/Carolus-and-Kocch-(2023)-MAILS.pdf
- Preprint: https://arxiv.org/abs/2302.09319
- Short-version article: https://pmc.ncbi.nlm.nih.gov/articles/PMC11544068/

## Citation

Carolus, A., Koch, M. J., Straka, S., Latoschik, M. E., & Wienrich, C. (2023). MAILS – Meta AI literacy scale: Development and testing of an AI literacy questionnaire based on well-founded competency models and psychological change- and meta-competencies. Computers in Human Behavior: Artificial Humans, 1(2), 100014. https://doi.org/10.1016/j.chbah.2023.100014

Koch, M. J., Carolus, A., Wienrich, C., & Latoschik, M. E. (2024). Meta AI literacy scale: Further validation and development of a short version. Heliyon, 10(21), e39686. https://doi.org/10.1016/j.heliyon.2024.e39686

## License / terms

- Questionnaire PDFs (the transcribed source): published for download by the authors on the project page; **no license or terms-of-use statement** on the page or in the PDFs.
- Published article, which prints all 34 English items in Appendix A (in an earlier wording, see below): "This is an open access article under the CC BY-NC-ND license (http://creativecommons.org/licenses/by-nc-nd/4.0/)." (author copy linked above, p. 1)
- arXiv preprint 2302.09319, which also lists the English items (earlier wording): CC BY 4.0 (http://creativecommons.org/licenses/by/4.0/), per https://arxiv.org/abs/2302.09319
- Heliyon short-version article: "This is an open access article under the CC BY license (http://creativecommons.org/licenses/by/4.0/)." It does not print the item texts (only codes AIL01–AIL34).

Assessment: the authors distribute the scale openly for use, and the items are published open access, but the exact questionnaire wording carries no explicit license. The article license is NC-ND. Flag for the user.

## Transcription

- 34 items (`mails_1` … `mails_34`, source numbering) in one `field-list` group, `select_one mails_0_10` (values 0–10, labels are the printed numbers).
- Group label = instruction paragraph; group hint = anchor explanation (0 / 10) + "Bitte kreisen Sie den Wert ein…" (verbatim, de + en).
- Base language Deutsch (de), English (en) from the official English PDF.
- Skipped: the scoring section (subscale mapping), page headers, citation footer.

Subscales (scoring, from the PDF): Use & Apply 1–6, Know & Understand 7–12, Detect AI 13–15, AI Ethics 16–18, Create AI 19–22, AI Problem-Solving 23–25, Learning 26–28, Persuasion Literacy 29–31, Emotion Regulation 32–34.

### Short version (Koch et al., 2024)

Not imported as a separate group; the 10 short-version items are identical in wording to these long-version items (short no. → long item):

| Short | Long | Facet |
|---|---|---|
| 1 | mails_13 | Detect AI |
| 2 | mails_20 | Create AI |
| 3 | mails_28 | Learning |
| 4 | mails_33 | Emotion Regulation |
| 5 | mails_16 | AI Ethics |
| 6 | mails_19 | Create AI |
| 7 | mails_3 | Apply AI |
| 8 | mails_25 | AI Problem-Solving |
| 9 | mails_30 | Persuasion Literacy |
| 10 | mails_10 | Understand AI |

(English short no. 4 lacks the final period in the short PDF; otherwise identical.)

## Doubts

- German item 2 reads "um meine Leben zu vereinfachen" (sic) in the source; kept verbatim.
- pdftotext split words at "f" (e.g. "vereinf achen", "Auf gaben"); checked against the rendered page and rejoined. Item 7 de has „KI" with a straight closing quote as in the PDF.
- The questionnaire PDFs differ from the article's Appendix A (article has "in everyday life" in items 1–6 and 29–34; items 26/27 swapped). The questionnaire PDFs were used as the authoritative instrument.
- License: see above (no explicit license on the questionnaire PDFs; article CC BY-NC-ND).

## Decision (2026-10-27)
The user decided that a verbatim conversion to DDI is a format change, not an alteration, so the ND terms allow it. Verdict: **ready**.

## Titles (#64, 2026-10-28)

`titl` „Meta AI Literacy Scale“, `altTitl` „MAILS“: printed in the published article (abstract: „the Meta AI Literacy Scale (MAILS)“; title „MAILS - Meta AI literacy scale: …“).
