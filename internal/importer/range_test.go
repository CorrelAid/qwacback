package importer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// Range variables as formtransform v0.8.0's xlsformToDdi writes them: start
// and end in valrng/range (defaults filled in), the XLSForm parameters
// verbatim in a cdl:parameters note when the form had any. "age" is an
// integer whose valrng comes from a constraint: not a range, no bounds.
const rangeCodebook = `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5">
  <stdyDscr><citation><titlStmt><titl>Ranges</titl></titlStmt></citation></stdyDscr>
  <dataDscr>
    <varGrp ID="VG_hours" name="hours" type="grid" var="V_mon V_tue">
      <txt>Hours</txt>
    </varGrp>
    <var ID="V_half" name="half" intrvl="contin">
      <qstn responseDomainType="numeric" seqNo="1"><qstnLit>Half steps</qstnLit></qstn>
      <valrng><range min="0" max="5"/></valrng>
      <varFormat type="numeric" schema="other"/>
      <notes type="cdl:parameters">start=0 end=5 step=0.5</notes>
    </var>
    <var ID="V_plain" name="plain" intrvl="contin">
      <qstn responseDomainType="numeric" seqNo="2"><qstnLit>No parameters</qstnLit></qstn>
      <valrng><range min="1" max="10"/></valrng>
      <varFormat type="numeric" schema="other"/>
    </var>
    <var ID="V_mon" name="mon" intrvl="contin">
      <qstn responseDomainType="numeric" seqNo="3"><preQTxt>Hours</preQTxt><qstnLit>Mon</qstnLit></qstn>
      <valrng><range min="0" max="24"/></valrng>
      <varFormat type="numeric" schema="other"/>
      <notes type="cdl:parameters">start=0;end=24;STEP=2</notes>
    </var>
    <var ID="V_tue" name="tue" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric" seqNo="4"><preQTxt>Hours</preQTxt><qstnLit>Tue</qstnLit></qstn>
      <varFormat type="numeric" schema="other"/>
    </var>
    <var ID="V_age" name="age" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric" seqNo="5"><qstnLit>Age</qstnLit></qstn>
      <valrng><range min="0" max="120"/></valrng>
      <varFormat type="numeric" schema="other"/>
      <notes type="cdl:constraint" subject="xlsform-xpath">. &gt;= 0 and . &lt;= 120</notes>
    </var>
  </dataDscr>
</codeBook>`

func TestImportRangeBounds(t *testing.T) {
	app, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(rangeCodebook))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ImportCodebook(app, mv, []byte(rangeCodebook)); err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]float64{
		"half":  {"min": 0, "max": 5, "step": 0.5},
		"plain": {"min": 1, "max": 10}, // no step stated, none invented
		"mon":   {"min": 0, "max": 24, "step": 2},
		"tue":   nil,
		"age":   nil, // constraint bounds, not a range
	}
	for name, w := range want {
		v, err := app.FindFirstRecordByFilter("variables", "name = {:n}", map[string]any{"n": name})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var got map[string]float64
		if raw := v.GetString("range_bounds"); raw != "" {
			if err := json.Unmarshal([]byte(raw), &got); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
		}
		if !reflect.DeepEqual(got, w) {
			t.Errorf("%s: range_bounds %v, want %v", name, got, w)
		}
	}
}
