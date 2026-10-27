# q02-aes2022: Adult Education Survey 2022 (Deutschland)

**Verdict: `no-instrument`** (no public German AES 2022 questionnaire was found)

## Sources checked (2026-10-27)

- Trend report, downloaded:
  https://www.kommweiter.bayern.de/imperia/md/content/stmas/qualifizierung/dokumente_download/aes-trendbericht_2022.pdf
  (86 pages; sha256 prefix 6cd09ede405d98de).
  - Citation: Bundesministerium für Bildung und Forschung. (2023). Weiterbildungsverhalten in Deutschland 2022:
    Ergebnisse des Adult Education Survey – AES-Trendbericht.
  - Contents: results only, no question wording. Footnote 9 points to the user handbook (Nutzerhandbuch) with the
    methods part, „voraussichtlich im ersten Quartal 2024 im Datenarchiv der GESIS … gemeinsam mit den AES-Daten
    für die interessierte Fachöffentlichkeit bereitgestellt“.
  - Imprint: „Diese Publikation wird als Fachinformation des Bundesministeriums für Bildung und Forschung kostenlos
    herausgegeben. Sie ist nicht zum Verkauf bestimmt …“. No reuse license.
- GESIS data: ZA8771, „Adult Education Survey (AES 2022 - Germany)“, DOI 10.4232/1.14234.
  - https://search.gesis.org/research_data/ZA8771 was not reachable: HTTP 403 / Cloudflare block for automated access.
  - Metadata via B2FIND (https://b2find.eudat.eu/dataset/f75ea9b9-cd05-5c91-8069-205e39c619b6) and the CESSDA OAI DDI
    record list the topic modules, not the items.
  - Access terms (CESSDA DDI `restrctn`): „A - Daten und Dokumente sind für die akademische Forschung und Lehre
    freigegeben.“
- Web searches for a public „AES 2022 Fragebogen“ / Methodenbericht (BMBF, Verian, GESIS) found only the Austrian
  instrument (Statistik Austria, AES_2022_Fragebogen_Statistik_Austria.pdf). That is a different questionnaire and was
  not used as a substitute.

## Assessment

The trend report is not a questionnaire. The German instrument probably comes with the GESIS documentation of ZA8771,
but that is (a) not reachable here and (b) released for „akademische Forschung und Lehre“, which is not a permission
to republish on the public qwac. Even if someone downloads it by hand, it would likely be `blocked-license` unless
BMBF/Verian grant reuse. No transcription.

Open: someone with a GESIS login could check whether ZA8771 has a questionnaire PDF with its own reuse note.

## Update: questionnaire found (user download)

`downloads/q02-aes2022/ZA8771_cod.pdf`, the GESIS ZA8771 "Handbuch zur Datennutzung" (Verian, 189 pages):
- pp. 22–155: the full German AES 2022 question program (CAPI), in 17 blocks, with filters and variable names;
- pp. 157–177: Listenheft, the show cards with the answer lists;
- pp. 178–: the method description.

Rights: "© Verian 2024" in the page footers; no reuse license. The data is released by GESIS for scientific use only.
Commissioned research (BMBF), not an official statistics form, so § 5 UrhG is unlikely to apply.
Verdict: blocked-license until BMBF/Verian consent. If granted, the blocks most relevant to this project are 11 (Bildungsbarrieren/-bedarf),
12 (Transparenz und Beratung) and 14 (Digitalisierung).
