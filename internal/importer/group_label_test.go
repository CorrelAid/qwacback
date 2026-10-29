package importer

import (
	"testing"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// Group labels as formtransform v0.9.0 writes them from the XLSForm
// `label_short` column: varGrp/labl before txt, one per language, on a grid,
// a select_multiple and a semi-open question's `other` group.
const groupLabelCodebook = `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5" xml:lang="de">
  <stdyDscr><citation><titlStmt><titl>Group labels</titl></titlStmt></citation></stdyDscr>
  <dataDscr>
    <varGrp ID="VG_sec" name="sec" type="section" var="V_a1 V_a2 V_e1_a V_e2_a V_e2_other V_e3">
      <labl>Modul E</labl>
      <txt>Modul E: Künstliche Intelligenz</txt>
    </varGrp>
    <varGrp ID="VG_attari" name="attari" type="grid" var="V_a1 V_a2">
      <labl>ATTARI-12</labl>
      <labl xml:lang="en">ATTARI-12 (en)</labl>
      <txt>Bitte geben Sie an …</txt>
      <txt xml:lang="en">Please indicate …</txt>
    </varGrp>
    <varGrp ID="VG_e1" name="e1" type="multipleResp" var="V_e1_a">
      <labl>E1</labl>
      <txt>Welche Tools?</txt>
    </varGrp>
    <varGrp ID="VG_e2" name="e2" type="other" var="V_e2_other" varGrp="VG_e2_choices">
      <labl>E2</labl>
      <labl xml:lang="en">E2 (en)</labl>
      <txt>Welche noch?</txt>
    </varGrp>
    <varGrp ID="VG_e2_choices" name="e2_choices" type="multipleResp" var="V_e2_a">
      <txt>Welche noch?</txt>
    </varGrp>
    <varGrp ID="VG_e4" name="e4" type="multipleResp" var="V_e4_a">
      <txt>Ohne Namen</txt>
    </varGrp>
    <var ID="V_a1" name="a1" intrvl="discrete">
      <qstn responseDomainType="category"><preQTxt>Bitte geben Sie an …</preQTxt><qstnLit>KI macht das Leben leichter</qstnLit></qstn>
      <catgry><catValu>1</catValu><labl>ja</labl></catgry>
    </var>
    <var ID="V_a2" name="a2" intrvl="discrete">
      <qstn responseDomainType="category"><preQTxt>Bitte geben Sie an …</preQTxt><qstnLit>KI ist gefährlich</qstnLit></qstn>
      <catgry><catValu>1</catValu><labl>ja</labl></catgry>
    </var>
    <var ID="V_e1_a" name="e1_a" intrvl="discrete">
      <qstn responseDomainType="multiple"><preQTxt>Welche Tools?</preQTxt><qstnLit>A</qstnLit></qstn>
      <catgry><catValu>0</catValu></catgry><catgry><catValu>1</catValu></catgry>
    </var>
    <var ID="V_e2_a" name="e2_a" intrvl="discrete">
      <qstn responseDomainType="multiple"><preQTxt>Welche noch?</preQTxt><qstnLit>A</qstnLit></qstn>
      <catgry><catValu>0</catValu></catgry><catgry><catValu>1</catValu></catgry>
    </var>
    <var ID="V_e2_other" name="e2_other" intrvl="discrete">
      <qstn responseDomainType="text"><qstnLit>Andere</qstnLit></qstn>
      <varFormat type="character"/>
    </var>
    <var ID="V_e3" name="e3" intrvl="discrete">
      <qstn responseDomainType="category"><qstnLit>Einzeln</qstnLit></qstn>
      <catgry><catValu>1</catValu><labl>ja</labl></catgry>
    </var>
    <var ID="V_e4_a" name="e4_a" intrvl="discrete">
      <qstn responseDomainType="multiple"><preQTxt>Ohne Namen</preQTxt><qstnLit>A</qstnLit></qstn>
      <catgry><catValu>0</catValu></catgry><catgry><catValu>1</catValu></catgry>
    </var>
  </dataDscr>
</codeBook>`

func TestImportGroupLabel(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(groupLabelCodebook))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ImportCodebook(app, mv, []byte(groupLabelCodebook)); err != nil {
		t.Fatal(err)
	}

	for name, want := range map[string]struct{ label, en, description string }{
		"attari": {"ATTARI-12", "ATTARI-12 (en)", "Bitte geben Sie an …"},
		"e1":     {"E1", "", "Welche Tools?"},
		"e2":     {"E2", "E2 (en)", "Welche noch?"},
		"e4":     {"", "", "Ohne Namen"},
	} {
		g, err := app.FindFirstRecordByData("variable_groups", "name", name)
		if err != nil {
			t.Fatalf("group %s: %v", name, err)
		}
		if got := g.GetString("label"); got != want.label {
			t.Errorf("%s.label = %q, want %q", name, got, want.label)
		}
		// The label must not replace the question text.
		if got := g.GetString("description"); got != want.description {
			t.Errorf("%s.description = %q, want %q", name, got, want.description)
		}
		var tr map[string]map[string]interface{}
		_ = g.UnmarshalJSONField("translations", &tr)
		if got, _ := tr["en"]["label"].(string); got != want.en {
			t.Errorf("%s.translations.en.label = %q, want %q", name, got, want.en)
		}
	}
	// Sections don't become groups, so their labl isn't stored.
	if _, err := app.FindFirstRecordByData("variable_groups", "name", "sec"); err == nil {
		t.Error("section imported as a group")
	}
}
