package routes

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"qwacback/internal/exporter"
	"qwacback/internal/importer"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// multilingualProveIt turns prove_it.xml into a German/English codebook:
// base language de on the codeBook, English siblings on neighbour_trust's
// texts, its first category label, and contact_knowledge's txt.
func multilingualProveIt(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../../seed_data/prove_it.xml")
	if err != nil {
		t.Fatal(err)
	}
	x := strings.Replace(string(raw), "<codeBook ", `<codeBook xml:lang="de" `, 1)
	replace := func(old, new string) {
		t.Helper()
		if !strings.Contains(x, old) {
			t.Fatalf("fixture text not found: %s", old)
		}
		x = strings.Replace(x, old, new, 1)
	}
	replace("<qstnLit>Do you think that your neighbours act in your best interests?</qstnLit>",
		`<preQTxt>Nachbarschaft</preQTxt><preQTxt xml:lang="en">Neighbourhood</preQTxt>`+
			`<qstnLit>Handeln Ihre Nachbarn in Ihrem Interesse?</qstnLit><qstnLit xml:lang="en">Do you think that your neighbours act in your best interests?</qstnLit>`+
			`<ivuInstr>Nachfragen</ivuInstr><ivuInstr xml:lang="en">Probe</ivuInstr>`)
	i := strings.Index(x, `name="neighbour_trust"`)
	j := i + strings.Index(x[i:], "<labl>")
	k := j + strings.Index(x[j:], "</labl>") + len("</labl>")
	x = x[:j] + `<labl>Ja</labl><labl xml:lang="en">Yes</labl>` + x[k:]
	// Grid: the group's txt and its members' preQTxt must match per language
	// (CDL Schematron), so the English sibling goes on all of them.
	const gridText = "If you did want to change things around here, do you know who to contact to help you in the following groups…?"
	replace("<txt>"+gridText+"</txt>", "<txt>"+gridText+`</txt><txt xml:lang="en">Who would you contact?</txt>`)
	x = strings.ReplaceAll(x, "<preQTxt>"+gridText+"</preQTxt>", "<preQTxt>"+gridText+`</preQTxt><preQTxt xml:lang="en">Who would you contact?</preQTxt>`)
	return x
}

// #35: the other languages of a multilingual codebook are stored, exported
// back as xml:lang siblings, returned by the API and searched.
func TestTranslationsRoundTrip(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	x := multilingualProveIt(t)
	mv, err := mxj.NewMapXml([]byte(x))
	if err != nil {
		t.Fatal(err)
	}
	if err := importer.ImportCodebookData(app, mv, []byte(x)); err != nil {
		t.Fatal(err)
	}

	// Stored
	study, _ := app.FindFirstRecordByFilter("studies", "id != ''")
	if got := study.GetString("language"); got != "de" {
		t.Errorf("study language: got %q", got)
	}
	v, _ := app.FindFirstRecordByFilter("variables", "name = 'neighbour_trust'")
	if got := v.GetString("question"); got != "Handeln Ihre Nachbarn in Ihrem Interesse?" {
		t.Errorf("base question: got %q", got)
	}
	var tr map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(v.GetString("translations")), &tr); err != nil {
		t.Fatalf("translations: %v (%s)", err, v.GetString("translations"))
	}
	en := tr["en"]
	if en["question"] != "Do you think that your neighbours act in your best interests?" || en["prequestion_text"] != "Neighbourhood" || en["ivu_instructions"] != "Probe" {
		t.Errorf("en translations: got %v", en)
	}
	if cats, _ := en["categories"].(map[string]interface{}); cats["1"] != "Yes" {
		t.Errorf("en category labels: got %v", en["categories"])
	}
	g, _ := app.FindFirstRecordByFilter("variable_groups", "name = 'contact_knowledge'")
	if !strings.Contains(g.GetString("translations"), `"description":"Who would you contact?"`) {
		t.Errorf("group translations: got %s", g.GetString("translations"))
	}

	// Exported
	out, err := exporter.StudyXML(study)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`xml:lang="de"`,
		`<qstnLit>Handeln Ihre Nachbarn in Ihrem Interesse?</qstnLit>`,
		`<qstnLit xml:lang="en">Do you think that your neighbours act in your best interests?</qstnLit>`,
		`<preQTxt xml:lang="en">Neighbourhood</preQTxt>`,
		`<ivuInstr xml:lang="en">Probe</ivuInstr>`,
		`<labl xml:lang="en">Yes</labl>`,
		`<txt xml:lang="en">Who would you contact?</txt>`,
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("export is missing %s", want)
		}
	}

	// API: language and question-text translations
	qs, err := AssembleQuestions(app, study.Id)
	if err != nil {
		t.Fatal(err)
	}
	var nt *Question
	for i := range qs {
		if qs[i].Name == "neighbour_trust" {
			nt = &qs[i]
		}
	}
	if nt == nil || nt.Language != "de" || nt.Translations["en"] != "Do you think that your neighbours act in your best interests?" {
		t.Fatalf("question: got %+v", nt)
	}

	// Search matches the translation
	got, err := SearchQuestions(app, "neighbours interests", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || got[0].Name != "neighbour_trust" {
		t.Errorf("search by English text: got %v", ids(got))
	}
}
