package converter

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const minimalForm = `{"survey":[{"type":"integer","name":"age","label":"Age?"}]}`

// fakeEmitter points DDIEmitterURL at a server answering every request with
// status and body, and restores it when the test ends.
func fakeEmitter(t *testing.T, status int, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	prev := DDIEmitterURL
	DDIEmitterURL = srv.URL
	t.Cleanup(func() { DDIEmitterURL = prev })
}

func codeBook(dataDscr string) string {
	return `<?xml version="1.0" ?>
<codeBook xmlns="ddi:codebook:2_5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" version="2.5">
  <stdyDscr><citation><titlStmt><titl>T</titl></titlStmt></citation></stdyDscr>
  <fileDscr ID="F1" URI="data.csv"/>
  <dataDscr>` + dataDscr + `</dataDscr>
</codeBook>`
}

// Elements and attributes qwacback's Go structs don't model must reach the
// client unchanged (#14); only the namespace and the dangling `files` go.
func TestXLSFormToDDI_PassesElementsThrough(t *testing.T) {
	fakeEmitter(t, http.StatusOK, codeBook(`
    <var ID="V_sat" name="sat" intrvl="contin" files="F1">
      <qstn responseDomainType="numeric">
        <preQTxt>Intro</preQTxt>
        <qstnLit>Satisfaction &amp; more</qstnLit>
        <ivuInstr>Ask twice</ivuInstr>
      </qstn>
      <valrng><range min="1" max="10"/></valrng>
      <universe xml:lang="de">Alle</universe>
      <concept>Satisfaction</concept>
      <varFormat type="numeric" schema="other"/>
      <notes>kept</notes>
    </var>`))

	out, err := XLSFormToDDI([]byte(minimalForm))
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, want := range []string{
		`<var ID="V_sat" name="sat" intrvl="contin">`,
		`<preQTxt>Intro</preQTxt>`,
		`<qstnLit>Satisfaction &amp; more</qstnLit>`,
		`<ivuInstr>Ask twice</ivuInstr>`,
		`<range min="1" max="10"></range>`,
		`<universe xml:lang="de">Alle</universe>`,
		`<notes>kept</notes>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{`files=`, `xmlns`, `<dataDscr>`, `<codeBook`} {
		if strings.Contains(got, unwanted) {
			t.Errorf("unexpected %q in:\n%s", unwanted, got)
		}
	}
}

func TestXLSFormToDDI_WrapsSeveralInOrder(t *testing.T) {
	fakeEmitter(t, http.StatusOK, codeBook(`
    <varGrp ID="VG_g" type="multipleResp" var="V_a V_b"><concept>G</concept></varGrp>
    <var ID="V_a" name="a"><concept>A</concept></var>
    <var ID="V_b" name="b"><concept>B</concept></var>`))

	out, err := XLSFormToDDI([]byte(minimalForm))
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.HasPrefix(strings.TrimPrefix(got, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"), "<dataDscr>") {
		t.Fatalf("expected <dataDscr> wrapper, got:\n%s", got)
	}
	g, a, b := strings.Index(got, `ID="VG_g"`), strings.Index(got, `ID="V_a"`), strings.Index(got, `ID="V_b"`)
	if !(g < a && a < b) {
		t.Errorf("order not kept:\n%s", got)
	}
}

// A form of only notes yields an empty <dataDscr>; that's an input problem,
// not a 200 with an empty element (#10).
func TestXLSFormToDDI_EmptyDataDscrIsInputError(t *testing.T) {
	fakeEmitter(t, http.StatusOK, codeBook(``))

	_, err := XLSFormToDDI([]byte(`{"survey":[{"type":"note","name":"n","label":"Hi"}]}`))
	if err == nil {
		t.Fatal("expected an error for a form without answerable questions")
	}
	if errors.Is(err, ErrConverterUnavailable) {
		t.Errorf("empty result is an input error, got %v", err)
	}
}

// The library's message names the question; it must reach the client as is (#11).
func TestXLSFormToDDI_RejectionKeepsLibraryMessage(t *testing.T) {
	msg := `type "rank" (question "a") is not in the registry`
	fakeEmitter(t, http.StatusBadRequest, `{"error":"`+strings.ReplaceAll(msg, `"`, `\"`)+`"}`)

	_, err := XLSFormToDDI([]byte(minimalForm))
	if err == nil || err.Error() != msg {
		t.Fatalf("expected %q, got %v", msg, err)
	}
	if errors.Is(err, ErrConverterUnavailable) {
		t.Error("a rejection is an input error, not unavailability")
	}
}

func TestXLSFormToDDI_Unavailable(t *testing.T) {
	t.Run("server error", func(t *testing.T) {
		fakeEmitter(t, http.StatusInternalServerError, "boom")
		if _, err := XLSFormToDDI([]byte(minimalForm)); !errors.Is(err, ErrConverterUnavailable) {
			t.Errorf("expected ErrConverterUnavailable, got %v", err)
		}
	})
	t.Run("malformed response", func(t *testing.T) {
		fakeEmitter(t, http.StatusOK, "<codeBook><dataDscr>")
		if _, err := XLSFormToDDI([]byte(minimalForm)); !errors.Is(err, ErrConverterUnavailable) {
			t.Errorf("expected ErrConverterUnavailable, got %v", err)
		}
	})
	t.Run("unreachable", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		url := srv.URL
		srv.Close()
		prev := DDIEmitterURL
		DDIEmitterURL = url
		t.Cleanup(func() { DDIEmitterURL = prev })
		if _, err := XLSFormToDDI([]byte(minimalForm)); !errors.Is(err, ErrConverterUnavailable) {
			t.Errorf("expected ErrConverterUnavailable, got %v", err)
		}
	})
}

// #22: a guidance_hint column in the request reaches the sidecar. Before,
// SurveyRow had no such field and json.Unmarshal dropped it.
func TestXLSFormToDDI_ForwardsGuidanceHintColumn(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
		_, _ = w.Write([]byte(codeBook(`<var ID="V_a" name="a"><concept>A</concept></var>`)))
	}))
	t.Cleanup(srv.Close)
	prev := DDIEmitterURL
	DDIEmitterURL = srv.URL
	t.Cleanup(func() { DDIEmitterURL = prev })

	if _, err := XLSFormToDDI([]byte(`{"survey":[{"type":"integer","name":"a","label":"A","guidance_hint":"Bei Unsicherheit nachfragen"}]}`)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"guidance_hint":"Bei Unsicherheit nachfragen"`) {
		t.Errorf("guidance_hint not forwarded: %s", got)
	}
}
