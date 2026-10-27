# q14-van-westendorp: Van Westendorp Price Sensitivity Meter

**Verdict: `ready`**: 4 questions (decimal), German only.

## Sources (retrieved 2026-10-27)

- Method: Van Westendorp, P. H. (1976). NSS-Price Sensitivity Meter (PSM): A new approach to study
  consumer perception of prices. In *Proceedings of the 29th ESOMAR Congress* (pp. 139–167). ESOMAR.
  The paper is only available through ESOMAR. It was not downloaded.
- German wording: Lautner, K. (2019). *Der Einfluss des Alters auf die Kaufentscheidung und
  Qualitätseinschätzung von Biowein* [Masterarbeit, FH Burgenland, FH-StG Internationales
  Weinmarketing; Betreuer: W. Ziniel], p. 57 (section 6.1.5 "Zahlungsbereitschaft – Kaufbereitschaft").
  - Record: https://fhburgenland.contentdm.oclc.org/digital/collection/p15425dc/id/77431
  - Page PDF: https://fhburgenland.contentdm.oclc.org/digital/api/collection/p15425dc/id/77390/download
    (page 67 of the scan = printed p. 57). Local copy: `downloads/q14-van-westendorp/fhburgenland_77390.pdf`
  - The thesis cites the method description to "Hermann & Fassnach, 2016, pp. 174-176". This is probably
    a garbled citation of Simon, H. & Fassnacht, M. (2016), *Preismanagement* (4. Aufl.), Springer Gabler.
    That book is not open access and was not checked.
- Checked, not used: Müller, Voigt & Erichson (2009), FEMM Working Paper 27, OVGU Magdeburg
  (https://www.fww.ovgu.de/fww_media/femm/femm_2009/2009_27.pdf). It cites Van Westendorp but prints
  no question wording. A quantilope brochure (commercial) appears to use the same four sentences (it was
  not downloaded).

## License / terms

- Method: the PSM is a published, widely used method with no protection on the procedure. The 1976 paper
  carries no reuse statement that we could see (we had no access).
- Wording: the repository record for Lautner (2019) has the field `Rights : openAccess`. There is no
  explicit CC license. The four sentences are short, generic questions that follow the standard PSM
  pattern. They are probably below the threshold for copyright protection (Schöpfungshöhe), but **the user
  should decide** at merge. Attribution is in `study.citation` and `study.license`.

## Transcription

- Four `decimal` questions, verbatim and in the source's order: günstig, teuer aber gerade noch akzeptabel,
  zu günstig, zu teuer. The source has no item codes. The names `psm_*` are derived from the
  wording.
- **Placeholder:** the source says "für das Produkt". I kept that verbatim and did not replace it
  with a `<PRODUKT>` template placeholder, because that would change the source wording. The source has no
  currency or unit, so there is no hint and no constraint.
- No English version: the German source has no official English version. The common English
  wordings (e.g. "…so inexpensive that you would question its quality") are not equivalent to "zu
  günstig". Pairing them would amount to a translation.
- Build: no errors or round-trip warnings (4 rows in, 4 out). Validation: `valid`.

## Doubts

- The German wording comes from a master's thesis, not a textbook or an official page. It is the best openly
  accessible, citable source found. The thesis wording of "zu günstig" lacks the quality-doubt clause of the
  English original.
- The question order follows the thesis (günstig, teuer, zu günstig, zu teuer), not the commonly listed
  order (too cheap, cheap, expensive, too expensive).
