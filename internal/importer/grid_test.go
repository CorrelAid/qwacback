package importer

import (
	"testing"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// A table of number fields as formtransform v0.8.0 writes it: a grid with
// numeric members, no concepts, and a sub-heading between two rows.
const numberGridCodebook = `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5" xml:lang="de">
  <stdyDscr>
    <citation><titlStmt><titl>Number grid</titl></titlStmt></citation>
  </stdyDscr>
  <dataDscr>
    <varGrp ID="VG_f1_1" name="f1_1" type="grid" var="V_gesamt V_m V_w">
      <txt>Wie viele Beschäftigte hatte Ihr Unternehmen insgesamt?</txt>
      <txt xml:lang="en">How many employees did your enterprise have in total?</txt>
      <notes type="cdl:grid_note" subject="m">davon: Geschlecht</notes>
      <notes type="cdl:grid_note" subject="m" xml:lang="en">of whom: sex</notes>
    </varGrp>
    <var ID="V_gesamt" name="gesamt" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric"><preQTxt>Wie viele Beschäftigte hatte Ihr Unternehmen insgesamt?</preQTxt><qstnLit>Beschäftigte insgesamt</qstnLit></qstn>
      <varFormat type="numeric" schema="other"/>
    </var>
    <var ID="V_m" name="m" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric"><preQTxt>Wie viele Beschäftigte hatte Ihr Unternehmen insgesamt?</preQTxt><qstnLit>Männlich</qstnLit></qstn>
      <varFormat type="numeric" schema="other"/>
    </var>
    <var ID="V_w" name="w" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric"><preQTxt>Wie viele Beschäftigte hatte Ihr Unternehmen insgesamt?</preQTxt><qstnLit>Weiblich</qstnLit></qstn>
      <varFormat type="numeric" schema="other"/>
    </var>
  </dataDscr>
</codeBook>`

func TestImportNumberGrid(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(numberGridCodebook))
	if err != nil {
		t.Fatal(err)
	}
	studyID, err := ImportCodebook(app, mv, []byte(numberGridCodebook))
	if err != nil {
		t.Fatal(err)
	}

	vars, err := app.FindAllRecords("variables")
	if err != nil {
		t.Fatal(err)
	}
	if len(vars) != 3 {
		t.Fatalf("imported %d variables, want 3 (a variable without <concept> must import)", len(vars))
	}
	byName := map[string]map[string]string{}
	for _, v := range vars {
		if v.GetString("study") != studyID {
			continue
		}
		byName[v.GetString("name")] = map[string]string{
			"answer_type": v.GetString("answer_type"),
			"concept":     v.GetString("concept"),
			"grid_note":   v.GetString("grid_note"),
		}
	}
	for name, want := range map[string]map[string]string{
		"gesamt": {"answer_type": "integer", "concept": "", "grid_note": ""},
		"m":      {"answer_type": "integer", "concept": "", "grid_note": "davon: Geschlecht"},
		"w":      {"answer_type": "integer", "concept": "", "grid_note": ""},
	} {
		for field, w := range want {
			if got := byName[name][field]; got != w {
				t.Errorf("%s.%s = %q, want %q", name, field, got, w)
			}
		}
	}

	m, err := app.FindFirstRecordByData("variables", "name", "m")
	if err != nil {
		t.Fatal(err)
	}
	var tr map[string]map[string]interface{}
	if err := m.UnmarshalJSONField("translations", &tr); err != nil {
		t.Fatal(err)
	}
	if got := tr["en"]["grid_note"]; got != "of whom: sex" {
		t.Errorf("m.translations.en.grid_note = %v, want %q", got, "of whom: sex")
	}

	groups, err := app.FindAllRecords("variable_groups")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups[0].GetString("type") != "grid" {
		t.Fatalf("groups = %d, want one grid", len(groups))
	}
}
