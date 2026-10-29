package examples

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"qwacback/internal/converter"
)

func TestGetAll(t *testing.T) {
	reset()
	all, err := GetAll()
	if err != nil {
		t.Skip("ddi-emitter sidecar not reachable; examples not populated")
	}

	expectedTypes := []string{
		"select_one",
		"select_multiple",
		"select_one_other",
		"select_multiple_other",
		"grid",
		"integer",
		"text",
		"select_one_long_list",
		"select_multiple_long_list",
	}

	if len(all) != len(expectedTypes) {
		t.Fatalf("Expected %d examples, got %d", len(expectedTypes), len(all))
	}

	for i, exp := range expectedTypes {
		if all[i].Type != exp {
			t.Errorf("Example %d: expected type %s, got %s", i, exp, all[i].Type)
		}
		if all[i].Label == "" {
			t.Errorf("Example %d (%s): label is empty", i, exp)
		}
		if len(all[i].XLSForm.Survey) == 0 {
			t.Errorf("Example %d (%s): survey is empty", i, exp)
		}
		if all[i].DDI == "" {
			t.Errorf("Example %d (%s): DDI is empty", i, exp)
		}
		if !strings.Contains(all[i].DDI, "<?xml") {
			t.Errorf("Example %d (%s): DDI missing XML declaration", i, exp)
		}
	}
}

func TestGetByType(t *testing.T) {
	reset()
	if _, err := GetAll(); err != nil {
		t.Skip("ddi-emitter sidecar not reachable; examples not populated")
	}

	ex, err := GetByType("select_one")
	if err != nil || ex == nil {
		t.Fatal("Expected select_one example")
	}
	if ex.Type != "select_one" {
		t.Errorf("Expected type select_one, got %s", ex.Type)
	}

	missing, err := GetByType("nonexistent")
	if err != nil || missing != nil {
		t.Error("Expected nil for nonexistent type")
	}
}

// fakeEmitter points the converter at handler for the rest of the test and
// counts the requests it gets.
func fakeEmitter(t *testing.T, handler http.HandlerFunc) *atomic.Int32 {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	prev := converter.DDIEmitterURL
	converter.DDIEmitterURL = srv.URL
	t.Cleanup(func() { converter.DDIEmitterURL = prev; reset() })
	reset()
	return &calls
}

const oneVar = `<codeBook xmlns="ddi:codebook:2_5"><dataDscr><var ID="V_x" name="x"><concept>X</concept></var></dataDscr></codeBook>`

// While the sidecar is down, GetAll reports it instead of an empty list, and
// doesn't repeat the conversions on every call (#13).
func TestGetAll_SidecarDown(t *testing.T) {
	calls := fakeEmitter(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})

	if _, err := GetAll(); !errors.Is(err, converter.ErrConverterUnavailable) {
		t.Fatalf("expected ErrConverterUnavailable, got %v", err)
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("expected populate to stop at the first failure, got %d sidecar calls", n)
	}
	if _, err := GetByType("single_choice"); err == nil {
		t.Error("GetByType must report the outage, not 'not found'")
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("expected no retry within retryAfter, got %d sidecar calls", n)
	}
}

// One rejected example doesn't take the others down with it.
func TestGetAll_KeepsExamplesThatConvert(t *testing.T) {
	fakeEmitter(t, func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		if strings.Contains(buf.String(), "wochenende") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"rejected"}`))
			return
		}
		_, _ = w.Write([]byte(oneVar))
	})

	all, err := GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != len(defs)-1 {
		t.Fatalf("expected %d examples, got %d", len(defs)-1, len(all))
	}
	for _, e := range all {
		if e.Type == "select_multiple" {
			t.Error("rejected example should be left out")
		}
	}
}

// Examples are keyed by registry slug; the ids from before #49 still resolve
// to the same example while the survey site switches over.
func TestGetByType_LegacyAliases(t *testing.T) {
	fakeEmitter(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(oneVar))
	})

	for old, slug := range legacyIDs {
		ex, err := GetByType(old)
		if err != nil {
			t.Fatal(err)
		}
		if ex == nil {
			t.Errorf("%s: alias of %s not found", old, slug)
			continue
		}
		if ex.Type != slug {
			t.Errorf("%s: got example %s, want %s", old, ex.Type, slug)
		}
		bySlug, _ := GetByType(slug)
		if bySlug == nil || bySlug.Label != ex.Label {
			t.Errorf("%s and %s should serve the same example", old, slug)
		}
	}
}

// Every example id is a registry slug, not an old id, and there is no range
// example: the answer type was dropped (#69).
func TestDefs_KeyedBySlug(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range defs {
		if _, old := legacyIDs[d.Type]; old {
			t.Errorf("example %s uses a legacy id", d.Type)
		}
		if d.Type == "range" {
			t.Error("range was dropped (#69); there must be no range example")
		}
		if seen[d.Type] {
			t.Errorf("duplicate example %s", d.Type)
		}
		seen[d.Type] = true
	}
	for old, slug := range legacyIDs {
		if !seen[slug] {
			t.Errorf("alias %s points at %s, which has no example", old, slug)
		}
	}
}
