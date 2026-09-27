package importer

import (
	"testing"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// #44: the answer types formtransform's DDI marks, as xlsformToDdi v0.7
// writes them.
func TestImportAnswerTypes(t *testing.T) {
	x := `<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5">
  <stdyDscr><citation><titlStmt><titl>Answer types</titl></titlStmt></citation></stdyDscr>
  <dataDscr>
    <var ID="V_a" name="a" intrvl="contin" dcml="0"><qstn responseDomainType="numeric" seqNo="1"><qstnLit>A</qstnLit></qstn><concept>A</concept><varFormat type="numeric" schema="other"/></var>
    <var ID="V_b" name="b" intrvl="contin"><qstn responseDomainType="numeric" seqNo="2"><qstnLit>B</qstnLit></qstn><concept>B</concept><varFormat type="numeric" schema="other"/></var>
    <var ID="V_c" name="c" intrvl="discrete"><qstn responseDomainType="text" seqNo="3"><qstnLit>C</qstnLit></qstn><concept>C</concept><varFormat type="character" schema="other" category="date"/></var>
    <var ID="V_d" name="d" intrvl="discrete"><qstn responseDomainType="text" seqNo="4"><qstnLit>D</qstnLit></qstn><concept>D</concept><varFormat type="character" schema="other" category="time"/></var>
    <var ID="V_e" name="e" intrvl="contin"><qstn responseDomainType="numeric" seqNo="5"><qstnLit>E</qstnLit></qstn><valrng><range min="0" max="10"/></valrng><concept>E</concept><varFormat type="numeric" schema="other"/></var>
    <var ID="V_f" name="f" intrvl="discrete"><qstn responseDomainType="text" seqNo="6"><qstnLit>F</qstnLit></qstn><concept>F</concept><varFormat type="character" schema="other"/></var>
    <var ID="V_g" name="g" intrvl="contin"><qstn responseDomainType="numeric" seqNo="7"><qstnLit>G</qstnLit></qstn><valrng><range maxExclusive="120"/></valrng><concept>G</concept><varFormat type="numeric" schema="other"/><notes type="cdl:constraint" subject="xlsform-xpath">. &lt; 120</notes></var>
    <var ID="V_h" name="h" intrvl="contin" dcml="0"><qstn responseDomainType="numeric" seqNo="8"><qstnLit>H</qstnLit></qstn><valrng><range maxExclusive="120"/></valrng><concept>H</concept><varFormat type="numeric" schema="other"/><notes type="cdl:constraint" subject="xlsform-xpath">. &lt; 120</notes><notes type="cdl:required">yes</notes></var>
  </dataDscr>
</codeBook>`
	app, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(x))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ImportCodebook(app, mv, []byte(x)); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"a": "integer", // dcml="0"
		"b": "decimal", // numeric without dcml
		"c": "date",
		"d": "time",
		"e": "range", // valrng without a constraint
		"f": "text",
		"g": "decimal", // valrng from a constraint
		"h": "integer", // integer with a constraint (two notes)
	}
	for name, at := range want {
		v, err := app.FindFirstRecordByFilter("variables", "name = {:n}", map[string]any{"n": name})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := v.GetString("answer_type"); got != at {
			t.Errorf("%s: answer_type %q, want %q", name, got, at)
		}
	}
}
