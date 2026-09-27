package exporter

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// importXML imports a codebook into a fresh test app and returns its study.
func importXML(t *testing.T, xmlData []byte) (*tests.TestApp, *core.Record) {
	t.Helper()
	app, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	mv, err := mxj.NewMapXml(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	id, err := importer.ImportCodebook(app, mv, xmlData)
	if err != nil {
		t.Fatal(err)
	}
	study, err := app.FindRecordById("studies", id)
	if err != nil {
		t.Fatal(err)
	}
	return app, study
}

func importSeed(t *testing.T, name string) (*tests.TestApp, *core.Record) {
	t.Helper()
	xmlData, err := os.ReadFile(filepath.Join("..", "..", "seed_data", name))
	if err != nil {
		t.Fatal(err)
	}
	return importXML(t, xmlData)
}

func wellFormed(t *testing.T, b []byte) {
	t.Helper()
	var doc interface{}
	if err := xml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("not well-formed: %v\n%s", err, b)
	}
}

// The study export is the codebook as imported, byte for byte.
func TestStudyXMLIsTheImportedCodebook(t *testing.T) {
	seeds, _ := filepath.Glob(filepath.Join("..", "..", "seed_data", "*.xml"))
	if len(seeds) == 0 {
		t.Fatal("no seed files")
	}
	for _, path := range seeds {
		t.Run(filepath.Base(path), func(t *testing.T) {
			want, _ := os.ReadFile(path)
			_, study := importXML(t, want)
			got, err := StudyXML(study)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("export differs from the imported file")
			}
		})
	}
}

func TestStudyXMLWithoutCodebook(t *testing.T) {
	app, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	c, _ := app.FindCollectionByNameOrId("studies")
	study := core.NewRecord(c)
	if _, err := StudyXML(study); err != ErrNoCodebook {
		t.Errorf("got %v, want ErrNoCodebook", err)
	}
}

// A group question has its varGrp and every var it refers to.
func TestGroupXMLGrid(t *testing.T) {
	app, _ := importSeed(t, "prove_it.xml")
	g, err := app.FindFirstRecordByFilter("variable_groups", "name = 'contact_knowledge'")
	if err != nil {
		t.Fatal(err)
	}
	out, err := GroupXML(app, g)
	if err != nil {
		t.Fatal(err)
	}
	wellFormed(t, out)
	s := string(out)
	if !strings.Contains(s, "<dataDscr>") || !strings.Contains(s, `<varGrp ID="VG1"`) {
		t.Errorf("want the varGrp in a dataDscr:\n%s", s)
	}
	if n := strings.Count(s, "<var "); n != 4 {
		t.Errorf("want 4 vars, got %d", n)
	}
	if strings.Contains(s, "ddi:codebook:2_5") {
		t.Errorf("fragment keeps the codebook namespace:\n%s", s)
	}
}

// A type="other" group brings its nested groups (@varGrp) and their vars.
func TestGroupXMLNested(t *testing.T) {
	app, _ := importSeed(t, "cdl_instrumente.xml")
	g, err := app.FindFirstRecordByFilter("variable_groups", "name = 'geschlecht'")
	if err != nil {
		t.Fatal(err)
	}
	out, err := GroupXML(app, g)
	if err != nil {
		t.Fatal(err)
	}
	wellFormed(t, out)
	s := string(out)
	for _, want := range []string{`ID="VG_geschlecht"`, `ID="VG_geschlecht_choices"`, `ID="V_geschlecht_weiblich"`, `ID="V_geschlecht_other"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(s, `ID="V_nps"`) {
		t.Errorf("holds a var outside the group")
	}
}

// A grid member comes with the grid, narrowed to it; a standalone var is bare.
func TestVariableXML(t *testing.T) {
	app, _ := importSeed(t, "prove_it.xml")

	member, err := app.FindFirstRecordByFilter("variables", "name = 'contact_council'")
	if err != nil {
		t.Fatal(err)
	}
	out, err := VariableXML(app, member)
	if err != nil {
		t.Fatal(err)
	}
	wellFormed(t, out)
	s := string(out)
	if !strings.Contains(s, `var="`+member.GetString("ddi_id")+`"`) || strings.Count(s, "<var ") != 1 {
		t.Errorf("want the grid narrowed to %s:\n%s", member.GetString("ddi_id"), s)
	}

	alone, err := app.FindFirstRecordByFilter("variables", "name = 'crime_change'")
	if err != nil {
		t.Fatal(err)
	}
	out, err = VariableXML(app, alone)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimPrefix(string(out), xml.Header), "<var ") {
		t.Errorf("want a bare var:\n%s", out)
	}
}

// What the records don't model (skip logic, cdl: notes, question order,
// postQTxt) reaches the question export, with the base language.
func TestVariableXMLKeepsWhatRecordsDontModel(t *testing.T) {
	x := `<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5" xml:lang="de">
  <stdyDscr><citation><titlStmt><titl>Notes</titl></titlStmt></citation></stdyDscr>
  <dataDscr>
    <varGrp ID="S1" name="teil1" type="section" var="V1 V2"><txt>Teil 1</txt></varGrp>
    <var ID="V1" name="alter" intrvl="contin" dcml="0">
      <qstn seqNo="1" responseDomainType="numeric"><qstnLit>Wie alt sind Sie?</qstnLit><postQTxt>In Jahren</postQTxt><postQTxt xml:lang="en">In years</postQTxt></qstn>
      <valrng><range min="0" max="120"/></valrng>
      <concept>Alter</concept>
      <varFormat type="numeric"/>
      <notes type="cdl:constraint" subject="xlsform-xpath">. &lt;= 120</notes>
    </var>
    <var ID="V2" name="rente" intrvl="discrete">
      <qstn seqNo="2" responseDomainType="text"><qstnLit>Seit wann?</qstnLit><backward qstn="V1"/></qstn>
      <universe clusion="I">Wenn alter größer als 65</universe>
      <concept>Rente</concept>
      <varFormat type="character"/>
      <notes type="cdl:relevant" subject="xlsform-xpath">${alter} &gt; 65</notes>
    </var>
  </dataDscr>
</codeBook>`
	app, _ := importXML(t, []byte(x))

	v, err := app.FindFirstRecordByFilter("variables", "name = 'rente'")
	if err != nil {
		t.Fatal(err)
	}
	if v.GetString("universe") != "Wenn alter größer als 65" {
		t.Errorf("universe: got %q", v.GetString("universe"))
	}
	out, err := VariableXML(app, v)
	if err != nil {
		t.Fatal(err)
	}
	wellFormed(t, out)
	s := string(out)
	for _, want := range []string{
		`<var xml:lang="de" ID="V2"`,
		`<notes type="cdl:relevant" subject="xlsform-xpath">${alter} &gt; 65</notes>`,
		`<universe clusion="I">`,
		`seqNo="2"`,
		`<backward qstn="V1"></backward>`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in\n%s", want, s)
		}
	}

	alter, _ := app.FindFirstRecordByFilter("variables", "name = 'alter'")
	if alter.GetString("hint") != "In Jahren" || !strings.Contains(alter.GetString("translations"), `"hint":"In years"`) {
		t.Errorf("hint: got %q, translations %s", alter.GetString("hint"), alter.GetString("translations"))
	}
	out, _ = VariableXML(app, alter)
	if !strings.Contains(string(out), `<postQTxt xml:lang="en">In years</postQTxt>`) {
		t.Errorf("postQTxt lost:\n%s", out)
	}
}
