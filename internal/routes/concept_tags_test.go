package routes

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"qwacback/internal/exporter"
	"qwacback/internal/importer"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// #19: extra <concept> elements are imported as tags (the first stays the
// concept), exported again with their xml:lang, and found by the search.
// Study keywords with attributes are kept too.
func TestConceptTagsRoundTrip(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	raw, err := os.ReadFile("../../seed_data/prove_it.xml")
	if err != nil {
		t.Fatal(err)
	}
	// The seed file carries tags of its own; start from an untagged copy.
	untagged := regexp.MustCompile(`\n\s*<concept xml:lang="[^"]*">[^<]*</concept>`).ReplaceAllString(string(raw), "")
	xmlData := strings.Replace(untagged,
		"<concept>Interpersonal trust</concept>",
		"<concept>Interpersonal trust</concept>\n    <concept xml:lang=\"de\">Vertrauen</concept>\n    <concept xml:lang=\"de\">Nachbarschaft</concept>", 1)
	xmlData = strings.Replace(xmlData,
		"<keyword>social capital</keyword>",
		"<keyword>social capital</keyword>\n        <keyword xml:lang=\"de\">Sozialkapital</keyword>", 1)
	if xmlData == untagged {
		t.Fatal("fixture edit did not apply")
	}

	mv, err := mxj.NewMapXml([]byte(xmlData))
	if err != nil {
		t.Fatal(err)
	}
	if err := importer.ImportCodebookData(app, mv, []byte(xmlData)); err != nil {
		t.Fatal(err)
	}

	v, err := app.FindFirstRecordByFilter("variables", "name = 'neighbour_trust'")
	if err != nil {
		t.Fatal(err)
	}
	if got := v.GetString("concept"); got != "Interpersonal trust" {
		t.Errorf("concept: want the first <concept>, got %q", got)
	}
	if got := v.GetString("tags"); got != `[{"lang":"de","text":"Vertrauen"},{"lang":"de","text":"Nachbarschaft"}]` {
		t.Errorf("tags: got %s", got)
	}
	study, err := app.FindFirstRecordByFilter("studies", "id != ''")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(study.GetString("keywords"), "Sozialkapital") {
		t.Errorf("keyword with xml:lang dropped: %s", study.GetString("keywords"))
	}

	out, err := exporter.ExportStudyToXML(app, study)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"<concept>Interpersonal trust</concept>",
		`<concept xml:lang="de">Vertrauen</concept>`,
		`<concept xml:lang="de">Nachbarschaft</concept>`,
		"Sozialkapital",
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("export is missing %s", want)
		}
	}

	got, err := SearchQuestions(app, "Vertrauen", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "neighbour_trust" {
		t.Errorf("search Vertrauen: expected neighbour_trust only, got %v", ids(got))
	}
}
