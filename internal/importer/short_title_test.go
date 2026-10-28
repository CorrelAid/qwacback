package importer

import (
	"testing"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// altTitl and subTitl become studies.short_title and studies.subtitle, in
// the base language; a codebook without them leaves both empty (#64).
func TestStudyTitles(t *testing.T) {
	cases := []struct {
		name, titlStmt, short, sub string
	}{
		{"none", `<titl>Plain</titl>`, "", ""},
		{"both", `<titl>Münsteraner Fragebogen zur Evaluation von Seminaren</titl><subTitl>2. Revision</subTitl><altTitl>MFE-Sr2</altTitl>`, "MFE-Sr2", "2. Revision"},
		{"base language first", `<titl>Scale</titl><altTitl xml:lang="en">EN</altTitl><altTitl>DE</altTitl>`, "DE", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			x := `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5" xml:lang="de">
  <stdyDscr><citation><titlStmt>` + c.titlStmt + `</titlStmt></citation></stdyDscr>
  <dataDscr>
    <var ID="V_q1" name="q1" intrvl="discrete">
      <qstn responseDomainType="text"><qstnLit>Frage?</qstnLit></qstn>
      <varFormat type="character" schema="other"/>
    </var>
  </dataDscr>
</codeBook>`
			mv, err := mxj.NewMapXml([]byte(x))
			if err != nil {
				t.Fatal(err)
			}
			if short, sub := StudyTitles(mv); short != c.short || sub != c.sub {
				t.Errorf("StudyTitles = %q, %q, want %q, %q", short, sub, c.short, c.sub)
			}

			app, err := tests.NewTestApp()
			if err != nil {
				t.Fatal(err)
			}
			defer app.Cleanup()
			id, err := ImportCodebook(app, mv, []byte(x))
			if err != nil {
				t.Fatal(err)
			}
			s, err := app.FindRecordById("studies", id)
			if err != nil {
				t.Fatal(err)
			}
			if got := s.GetString("short_title"); got != c.short {
				t.Errorf("short_title = %q, want %q", got, c.short)
			}
			if got := s.GetString("subtitle"); got != c.sub {
				t.Errorf("subtitle = %q, want %q", got, c.sub)
			}
		})
	}
}
