package routes

import (
	"encoding/json"
	"os"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

const groupLabelCodebook = `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5" xml:lang="de">
  <stdyDscr><citation><titlStmt><titl>Attitudes Towards Artificial Intelligence Scale</titl></titlStmt></citation></stdyDscr>
  <dataDscr>
    <varGrp ID="VG_attari" name="attari" type="grid" var="V_a1 V_a2">
      <labl>ATTARI-12</labl>
      <labl xml:lang="en">ATTARI-12 EN</labl>
      <txt>Bitte geben Sie an, wie sehr Sie zustimmen.</txt>
      <txt xml:lang="en">Please indicate how much you agree.</txt>
    </varGrp>
    <var ID="V_a1" name="a1" intrvl="discrete">
      <qstn responseDomainType="category"><preQTxt>Bitte geben Sie an, wie sehr Sie zustimmen.</preQTxt><qstnLit>KI macht das Leben leichter</qstnLit></qstn>
      <catgry><catValu>1</catValu><labl>ja</labl></catgry>
    </var>
    <var ID="V_a2" name="a2" intrvl="discrete">
      <qstn responseDomainType="category"><preQTxt>Bitte geben Sie an, wie sehr Sie zustimmen.</preQTxt><qstnLit>KI ist gefährlich</qstnLit></qstn>
      <catgry><catValu>1</catValu><labl>ja</labl></catgry>
    </var>
    <var ID="V_age" name="age" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric"><qstnLit>Wie alt sind Sie?</qstnLit></qstn>
      <varFormat type="numeric" schema="other"/>
    </var>
  </dataDscr>
</codeBook>`

// A group's label (varGrp/labl) is the question's label in the list, the
// detail and its group, and it is searched; a standalone variable has none.
func TestQuestionGroupLabel(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(groupLabelCodebook))
	if err != nil {
		t.Fatal(err)
	}
	studyID, err := importer.ImportCodebook(app, mv, []byte(groupLabelCodebook))
	if err != nil {
		t.Fatal(err)
	}
	qs, err := AssembleQuestions(app, studyID)
	if err != nil || len(qs) != 2 {
		t.Fatalf("questions = %+v, %v", qs, err)
	}
	grid, age := qs[0], qs[1]
	if grid.Label != "ATTARI-12" || grid.LabelTranslations["en"] != "ATTARI-12 EN" {
		t.Errorf("grid label = %q %v", grid.Label, grid.LabelTranslations)
	}
	if grid.QuestionText != "Bitte geben Sie an, wie sehr Sie zustimmen." {
		t.Errorf("grid question_text = %q", grid.QuestionText)
	}
	b, _ := json.Marshal(age)
	var raw map[string]any
	_ = json.Unmarshal(b, &raw)
	if _, ok := raw["label"]; ok {
		t.Errorf("standalone variable has a label: %s", b)
	}

	d, err := QuestionDetail(app, grid.ID)
	if err != nil {
		t.Fatal(err)
	}
	if d["label"] != "ATTARI-12" {
		t.Errorf("detail label = %v", d["label"])
	}
	b, _ = json.Marshal(d["group"])
	var g struct {
		Label        string                       `json:"label"`
		Description  string                       `json:"description"`
		Translations map[string]map[string]string `json:"translations"`
	}
	if err := json.Unmarshal(b, &g); err != nil {
		t.Fatal(err)
	}
	if g.Label != "ATTARI-12" || g.Translations["en"]["label"] != "ATTARI-12 EN" || g.Description == "" {
		t.Errorf("detail group = %s", b)
	}

	got, err := SearchQuestions(app, "ATTARI-12", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || got[0].ID != grid.ID {
		t.Errorf("search by label: got %v", ids(got))
	}
}

// The seed codebook of ATTARI-12 names its grid as the source prints it.
func TestSeedGroupLabel(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	raw, err := os.ReadFile("../../seed_data/attari12.xml")
	if err != nil {
		t.Fatal(err)
	}
	mv, err := mxj.NewMapXml(raw)
	if err != nil {
		t.Fatal(err)
	}
	studyID, err := importer.ImportCodebook(app, mv, raw)
	if err != nil {
		t.Fatal(err)
	}
	qs, err := AssembleQuestions(app, studyID)
	if err != nil || len(qs) != 1 {
		t.Fatalf("questions = %d, %v", len(qs), err)
	}
	if qs[0].Label != "ATTARI-12" || qs[0].LabelTranslations["en"] != "ATTARI-12" {
		t.Errorf("label = %q %v", qs[0].Label, qs[0].LabelTranslations)
	}
}
