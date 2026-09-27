package importer

import (
	"os"
	"strings"
	"testing"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// multilingual rewrites prove_it.xml so neighbour_trust's qstn texts, first
// category label and contact_knowledge's txt come in two languages. el(text,
// lang) renders one element; order decides which language comes first.
func multilingual(t *testing.T, codeBookLang string, order func(tag string) string) string {
	t.Helper()
	raw, err := os.ReadFile("../../seed_data/prove_it.xml")
	if err != nil {
		t.Fatal(err)
	}
	x := string(raw)
	if codeBookLang != "" {
		x = strings.Replace(x, "<codeBook ", `<codeBook xml:lang="`+codeBookLang+`" `, 1)
	}
	replace := func(old, new string) {
		t.Helper()
		if !strings.Contains(x, old) {
			t.Fatalf("fixture text not found: %s", old)
		}
		x = strings.Replace(x, old, new, 1)
	}
	replace("<qstnLit>Do you think that your neighbours act in your best interests?</qstnLit>",
		order("preQTxt")+order("qstnLit")+order("ivuInstr"))
	i := strings.Index(x, `name="neighbour_trust"`)
	j := i + strings.Index(x[i:], "<labl>")
	k := j + strings.Index(x[j:], "</labl>") + len("</labl>")
	x = x[:j] + order("labl") + x[k:]
	g := strings.Index(x, "<txt>If you did want")
	ge := g + strings.Index(x[g:], "</txt>") + len("</txt>")
	x = x[:g] + order("txt") + x[ge:]
	return x
}

func importAndRead(t *testing.T, x string) (question, pre, ivu, labl, txt string) {
	t.Helper()
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(x))
	if err != nil {
		t.Fatal(err)
	}
	if err := ImportCodebookData(app, mv, []byte(x)); err != nil {
		t.Fatal(err)
	}
	v, err := app.FindFirstRecordByFilter("variables", "name = 'neighbour_trust'")
	if err != nil {
		t.Fatal(err)
	}
	g, err := app.FindFirstRecordByFilter("variable_groups", "name = 'contact_knowledge'")
	if err != nil {
		t.Fatal(err)
	}
	cats := v.GetString("categories")
	labl = cats[strings.Index(cats, `"label":"`)+len(`"label":"`):]
	labl = labl[:strings.Index(labl, `"`)]
	return v.GetString("question"), v.GetString("prequestion_text"), v.GetString("ivu_instructions"), labl, g.GetString("description")
}

// #30: with one element per language (formtransform#135), every field keeps
// the base language, whatever the order.
func TestImportMultilingualKeepsBaseLanguage(t *testing.T) {
	el := func(tag, lang, text string) string {
		if lang == "" {
			return "<" + tag + ">" + text + "</" + tag + ">"
		}
		return "<" + tag + ` xml:lang="` + lang + `">` + text + "</" + tag + ">"
	}
	cases := []struct {
		name, codeBookLang string
		order              func(tag string) string
	}{
		{"base untagged first", "de", func(tag string) string { return el(tag, "", "BASE") + el(tag, "en", "EN") }},
		{"base untagged last", "de", func(tag string) string { return el(tag, "en", "EN") + el(tag, "", "BASE") }},
		{"all tagged, base named by codeBook", "de", func(tag string) string { return el(tag, "en", "EN") + el(tag, "de", "BASE") }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			q, pre, ivu, labl, txt := importAndRead(t, multilingual(t, c.codeBookLang, c.order))
			for field, got := range map[string]string{"question": q, "preQTxt": pre, "ivuInstr": ivu, "labl": labl, "txt": txt} {
				if got != "BASE" {
					t.Errorf("%s: got %q, want the base language", field, got)
				}
			}
		})
	}
}
