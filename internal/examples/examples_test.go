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
		"single_choice",
		"multiple_choice",
		"single_choice_other",
		"multiple_choice_other",
		"grid",
		"integer",
		"text",
		"single_choice_long_list",
		"multiple_choice_long_list",
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

	ex, err := GetByType("single_choice")
	if err != nil || ex == nil {
		t.Fatal("Expected single_choice example")
	}
	if ex.Type != "single_choice" {
		t.Errorf("Expected type single_choice, got %s", ex.Type)
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
		if e.Type == "multiple_choice" {
			t.Error("rejected example should be left out")
		}
	}
}
