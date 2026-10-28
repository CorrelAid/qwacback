package importer

import (
	"testing"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

const licensedCodebook = `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5" xml:lang="en">
  <stdyDscr>
    <citation>
      <titlStmt><titl>Licensed scale</titl><altTitl>LS</altTitl></titlStmt>
      <prodStmt><copyright>© The authors 2024</copyright></prodStmt>
      <biblCit>Doe, J. (2024). A scale. Journal, 1, 1–2.</biblCit>
      <holdings URI="https://doi.org/10.0000/example">Article</holdings>
    </citation>
    <dataAccs><useStmt><conditions>CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/)</conditions></useStmt></dataAccs>
  </stdyDscr>
  <dataDscr>
    <var ID="V_q1" name="q1" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric"><qstnLit>How old are you?</qstnLit></qstn>
      <concept>Age</concept>
      <varFormat type="numeric" schema="other"/>
    </var>
  </dataDscr>
</codeBook>`

func TestImportStoresLicense(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(licensedCodebook))
	if err != nil {
		t.Fatal(err)
	}
	id, err := ImportCodebook(app, mv, []byte(licensedCodebook))
	if err != nil {
		t.Fatal(err)
	}
	s, err := app.FindRecordById("studies", id)
	if err != nil {
		t.Fatal(err)
	}
	for field, want := range map[string]string{
		"license":      "CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/)",
		"license_id":   "CC-BY-4.0",
		"copyright":    "© The authors 2024",
		"citation":     "Doe, J. (2024). A scale. Journal, 1, 1–2.",
		"holdings_uri": "https://doi.org/10.0000/example",
		"short_title":  "LS",
	} {
		if got := s.GetString(field); got != want {
			t.Errorf("%s = %q, want %q", field, got, want)
		}
	}
}

func TestLicenseID(t *testing.T) {
	for conditions, want := range map[string]string{
		"CC BY 4.0": "CC-BY-4.0",
		"CC BY-SA 4.0 (https://creativecommons.org/licenses/by-sa/4.0/). …": "CC-BY-SA-4.0",
		"CC BY-NC 4.0 (https://creativecommons.org/licenses/by-nc/4.0/).":   "CC-BY-NC-4.0",
		"  CC BY-NC-ND 4.0 (https://…): non-commercial re-use":              "CC-BY-NC-ND-4.0",
		"CC BY-NC-SA 4.0": "CC-BY-NC-SA-4.0",
		// Only a license named first counts, not one quoted further on.
		"© European Union. Eurostat copyright notice: … licensed under the Creative Commons Attribution 4.0 International licence.": "",
		"Questionnaire PDFs without a license statement; the article is under CC BY-NC-ND 4.0.":                                     "",
		"CC BY 3.0":  "",
		"CC BY 4.01": "",
		"":           "",
	} {
		if got := LicenseID(conditions); got != want {
			t.Errorf("LicenseID(%q) = %q, want %q", conditions, got, want)
		}
	}
}
