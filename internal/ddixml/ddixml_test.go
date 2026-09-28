package ddixml

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
)

// A question text with inline XHTML: mixed content, where a line break is
// part of the text (the NPS question of seed_data/cdl_instrumente.xml).
const mixedCodebook = `<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5" xml:lang="de">
  <dataDscr>
    <var ID="V_nps" name="nps">
      <qstn responseDomainType="category">
        <qstnLit xmlns:xhtml="http://www.w3.org/1999/xhtml"><xhtml:p>Wie wahrscheinlich ist es, dass Sie <xhtml:em>&lt;ETWAS&gt;</xhtml:em> einer Freund*in oder Kolleg*in weiterempfehlen werden?</xhtml:p></qstnLit>
      </qstn>
      <catgry><catValu>0</catValu><labl>0 - Unwahrscheinlich</labl></catgry>
    </var>
  </dataDscr>
</codeBook>`

// text returns the character data of the first element named local.
func text(t *testing.T, doc []byte, local string) string {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(string(doc)))
	depth := 0
	var b strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch x := tok.(type) {
		case xml.StartElement:
			if depth > 0 || x.Name.Local == local {
				depth++
			}
		case xml.EndElement:
			if depth > 0 {
				depth--
				if depth == 0 {
					return b.String()
				}
			}
		case xml.CharData:
			if depth > 0 {
				b.Write(x)
			}
		}
	}
	t.Fatalf("no <%s> in output", local)
	return ""
}

func TestFragmentKeepsMixedContentText(t *testing.T) {
	lang, children, err := DataDscr([]byte(mixedCodebook))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Fragment(children, lang)
	if err != nil {
		t.Fatal(err)
	}

	// Regression: Encoder.Indent put "\n  " before <em>, so the XLSForm label
	// became "…dass Sie\n<ETWAS> einer…" and split across rows when copied.
	want := "Wie wahrscheinlich ist es, dass Sie <ETWAS> einer Freund*in oder Kolleg*in weiterempfehlen werden?"
	if got := text(t, out, "p"); got != want {
		t.Errorf("question text changed:\n got %q\nwant %q", got, want)
	}
	if got := text(t, out, "labl"); got != "0 - Unwahrscheinlich" {
		t.Errorf("labl = %q", got)
	}

	// Elements that only hold elements are still indented.
	s := string(out)
	for _, line := range []string{"\n  <qstn ", "\n    <qstnLit>", "\n  <catgry>", "\n</var>"} {
		if !strings.Contains(s, line) {
			t.Errorf("output lacks indented %q:\n%s", line, s)
		}
	}
}

func TestSeedQuestionTextsSurviveAFragment(t *testing.T) {
	raw, err := os.ReadFile("../../seed_data/cdl_instrumente.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, children, err := DataDscr(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range children {
		if c.Name != "var" {
			continue
		}
		// The text of every qstnLit must come out as it went in.
		out, err := Fragment([]*Element{c}, "")
		if err != nil {
			t.Fatal(err)
		}
		// Compare against the tokens as parsed, without any layout.
		var src strings.Builder
		enc := xml.NewEncoder(&src)
		for _, tok := range c.Tokens {
			enc.EncodeToken(tok)
		}
		enc.Flush()
		if !strings.Contains(src.String(), "qstnLit") {
			continue
		}
		if a, b := text(t, []byte(src.String()), "qstnLit"), text(t, out, "qstnLit"); strings.TrimSpace(a) != strings.TrimSpace(b) {
			t.Errorf("%s: qstnLit text changed:\n got %q\nwant %q", c.ID, b, a)
		}
	}
}
