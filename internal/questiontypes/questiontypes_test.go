package questiontypes

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qwacback/internal/converter"
)

// A trimmed QUESTION_TYPES as formtransform v0.7.1 serializes it.
const catalogue = `{
  "select_one": {"id":"type:select_one","label":"Select One","labels":{"en":"Select One","de":"Einfachauswahl"},"kind":"question","useWhen":"u","isVariant":false,"isComposite":false,"typeString":"select_one"},
  "select_one_other": {"id":"variant:select_one_other","label":"Select One with Other","labels":{"en":"Select One with Other","de":"Einfachauswahl mit Sonstiges"},"kind":"question","useWhen":"u","isVariant":true,"isComposite":false,"typeString":"select_one","base":"select_one","presentation":{"withOther":true,"withLongList":false}},
  "select_one_long_list": {"id":"variant:select_one_long_list","label":"Select One (Long List)","kind":"question","useWhen":"u","isVariant":true,"isComposite":false,"typeString":"select_one","base":"select_one","presentation":{"appearanceString":"minimal","withOther":false,"withLongList":true}},
  "integer": {"id":"type:integer","label":"Integer","kind":"question","useWhen":"u","isVariant":false,"isComposite":false,"typeString":"integer","aliases":["int"],"constraints":{"maxNameLength":20}},
  "grid": {"id":"composite:grid","label":"Grid / Matrix Group","kind":"question","useWhen":"u","isVariant":false,"isComposite":true,"bases":["begin_group","select_one"]}
}`

func useCatalogue(t *testing.T, status int, body string) *int {
	t.Helper()
	reset()
	t.Cleanup(reset)
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet || r.URL.Path != "/question-types" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	prev := converter.DDIEmitterURL
	converter.DDIEmitterURL = srv.URL
	t.Cleanup(func() { converter.DDIEmitterURL = prev })
	return &calls
}

func TestByAnswerType(t *testing.T) {
	calls := useCatalogue(t, http.StatusOK, catalogue)
	got, err := ByAnswerType()
	if err != nil {
		t.Fatal(err)
	}
	other := got["single_choice_other"]
	if other.RegistryType != "select_one_other" || other.Label["de"] != "Einfachauswahl mit Sonstiges" || other.Base != "select_one" ||
		other.Presentation != (Presentation{Choice: "one", WithOther: true}) {
		t.Errorf("single_choice_other: %+v", other)
	}
	if ll := got["single_choice_long_list"]; ll.Presentation != (Presentation{Choice: "one", LongList: true, Appearance: "minimal"}) {
		t.Errorf("single_choice_long_list: %+v", ll.Presentation)
	}
	// An entry without `labels` (before v0.7.1) still gets its English label.
	if in := got["integer"]; in.Label["en"] != "Integer" {
		t.Errorf("integer label: %+v", in.Label)
	}
	if in := got["integer"]; len(in.Aliases) != 1 || in.Aliases[0] != "int" {
		t.Errorf("integer aliases: %+v", in)
	}
	if !got["grid"].Presentation.Grid {
		t.Errorf("grid: %+v", got["grid"])
	}
	// Types the trimmed catalogue lacks are left out, not invented.
	if _, ok := got["multiple_choice"]; ok {
		t.Error("multiple_choice without a registry entry")
	}

	reg, _ := Registry()
	if reg["integer"].Label["en"] != "Integer" || !strings.Contains(string(reg["integer"].Constraints), "maxNameLength") {
		t.Errorf("registry integer: %+v", reg["integer"])
	}
	ByAnswerType()
	if *calls != 1 {
		t.Errorf("fetched %d times, want once", *calls)
	}
}

func TestUnavailableIsRetriedLater(t *testing.T) {
	calls := useCatalogue(t, http.StatusInternalServerError, "boom")
	if _, err := ByAnswerType(); !errors.Is(err, converter.ErrConverterUnavailable) {
		t.Fatalf("got %v", err)
	}
	ByAnswerType()
	if *calls != 1 {
		t.Errorf("retried within retryAfter: %d calls", *calls)
	}
	mu.Lock()
	lastTry = time.Now().Add(-retryAfter)
	mu.Unlock()
	ByAnswerType()
	if *calls != 2 {
		t.Errorf("not retried after retryAfter: %d calls", *calls)
	}
}

// Against the pinned formtransform (the sidecar, when it runs): every answer
// type qwacback hands out has its registry type.
func TestIntegration_EveryAnswerTypeInRegistry(t *testing.T) {
	raw, err := converter.QuestionTypes()
	if err != nil {
		t.Skipf("ddi-emitter not reachable: %v", err)
	}
	reg, ans, err := build(raw)
	if err != nil {
		t.Fatal(err)
	}
	for name, at := range answerTypes {
		got, ok := ans[name]
		if !ok {
			t.Errorf("answer type %s: registry has no %s", name, at.slug)
			continue
		}
		if got.Label["de"] == "" {
			t.Errorf("answer type %s: no German label", name)
		}
		// qwacback's fallback agrees with the registry.
		if rp := reg[at.slug].Presentation; rp != nil && (rp.WithOther != at.p.WithOther || rp.WithLongList != at.p.LongList) {
			t.Errorf("answer type %s: registry presentation %+v, qwacback %+v", name, *rp, at.p)
		}
	}
	if len(reg) < len(answerTypes) {
		t.Errorf("registry has %d types", len(reg))
	}
}
