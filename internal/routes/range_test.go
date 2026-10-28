package routes

import (
	"encoding/json"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// QuestionDetail (GET /api/questions/{id}, MCP get_question) returns a range
// variable's bounds as `range`, standalone and as a number-grid row, with
// only the bounds the codebook states; other variables have no `range`.
func TestQuestionDetailRange(t *testing.T) {
	const x = `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5">
  <stdyDscr><citation><titlStmt><titl>Ranges</titl></titlStmt></citation></stdyDscr>
  <dataDscr>
    <varGrp ID="VG_hours" name="hours" type="grid" var="V_mon V_tue"><txt>Hours</txt></varGrp>
    <var ID="V_half" name="half" intrvl="contin">
      <qstn responseDomainType="numeric" seqNo="1"><qstnLit>Half steps</qstnLit></qstn>
      <valrng><range min="0" max="5"/></valrng>
      <varFormat type="numeric" schema="other"/>
      <notes type="cdl:parameters">start=0 end=5 step=0.5</notes>
    </var>
    <var ID="V_mon" name="mon" intrvl="contin">
      <qstn responseDomainType="numeric" seqNo="2"><preQTxt>Hours</preQTxt><qstnLit>Mon</qstnLit></qstn>
      <valrng><range min="1" max="10"/></valrng>
      <varFormat type="numeric" schema="other"/>
    </var>
    <var ID="V_tue" name="tue" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric" seqNo="3"><preQTxt>Hours</preQTxt><qstnLit>Tue</qstnLit></qstn>
      <varFormat type="numeric" schema="other"/>
    </var>
  </dataDscr>
</codeBook>`
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(x))
	if err != nil {
		t.Fatal(err)
	}
	studyID, err := importer.ImportCodebook(app, mv, []byte(x))
	if err != nil {
		t.Fatal(err)
	}
	questions, err := AssembleQuestions(app, studyID)
	if err != nil {
		t.Fatal(err)
	}

	// detail returns each variable's `range` as JSON, by name ("" if absent).
	detail := func(id string) map[string]string {
		t.Helper()
		d, err := QuestionDetail(app, id)
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		var out struct {
			Variables []struct {
				Name  string          `json:"name"`
				Range json.RawMessage `json:"range"`
			} `json:"variables"`
		}
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}
		ranges := map[string]string{}
		for _, v := range out.Variables {
			ranges[v.Name] = string(v.Range)
		}
		return ranges
	}

	want := map[string]string{
		"half": `{"min":0,"max":5,"step":0.5}`,
		"mon":  `{"min":1,"max":10}`,
		"tue":  "",
	}
	got := map[string]string{}
	for _, q := range questions {
		for name, r := range detail(q.ID) {
			got[name] = r
		}
	}
	for name, w := range want {
		if r, ok := got[name]; !ok || r != w {
			t.Errorf("%s: range %q, want %q", name, r, w)
		}
	}
}
