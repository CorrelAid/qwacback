package examples

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"qwacback/internal/converter"
)

// Example represents a complete answer type example with both XLSForm and DDI representations.
type Example struct {
	Type    string          `json:"answer_type"`
	Label   string          `json:"label"`
	XLSForm converter.XLSForm `json:"xlsform"`
	DDI     string          `json:"ddi"`
}

// exampleDef holds the static definition before DDI generation.
type exampleDef struct {
	Type    string
	Label   string
	XLSForm converter.XLSForm
}


var defs = []exampleDef{
	{
		Type:  "single_choice",
		Label: "Single Choice — Bildungsgrad",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "select_one bildungsgrad", Name: "bildungsgrad", Label: "Was ist Ihr höchster Bildungsabschluss?"},
			},
			Choices: []converter.ChoiceRow{
				{ListName: "bildungsgrad", Name: "1", Label: "Kein Abschluss"},
				{ListName: "bildungsgrad", Name: "2", Label: "Haupt- oder Realschulabschluss"},
				{ListName: "bildungsgrad", Name: "3", Label: "Fachhochschulreife / Abitur"},
				{ListName: "bildungsgrad", Name: "4", Label: "Abgeschlossene Berufsausbildung"},
				{ListName: "bildungsgrad", Name: "5", Label: "Hochschulabschluss"},
			},
		},
	},
	{
		Type:  "multiple_choice",
		Label: "Multiple Choice — Wochenendtage",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "select_multiple wochenendtage", Name: "wochenende", Label: "An welchen Tagen des Wochenendes sind Sie erreichbar?"},
			},
			Choices: []converter.ChoiceRow{
				{ListName: "wochenendtage", Name: "sa", Label: "Samstag"},
				{ListName: "wochenendtage", Name: "so", Label: "Sonntag"},
			},
		},
	},
	{
		Type:  "single_choice_other",
		Label: "Single Choice mit Sonstiges — Aufmerksamkeitsquelle",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "select_one quelle", Name: "aufmerksam", Label: "Wie sind Sie auf unser Angebot aufmerksam geworden?"},
				{Type: "text", Name: "aufmerksam_other", Label: "Sonstiges (bitte angeben)", Relevance: "${aufmerksam} = 'other'"},
			},
			Choices: []converter.ChoiceRow{
				{ListName: "quelle", Name: "suchmaschine", Label: "Suchmaschine"},
				{ListName: "quelle", Name: "empfehlung", Label: "Persönliche Empfehlung"},
				{ListName: "quelle", Name: "soziale_medien", Label: "Soziale Medien"},
				{ListName: "quelle", Name: "other", Label: "Sonstiges"},
			},
		},
	},
	{
		Type:  "multiple_choice_other",
		Label: "Multiple Choice mit Sonstiges — Gerätebesitz",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "select_multiple geraete", Name: "geraetebesitz", Label: "Welche dieser Geräte besitzen Sie?"},
				{Type: "text", Name: "geraetebesitz_other", Label: "Sonstiges (bitte angeben)", Relevance: "${geraetebesitz} = 'other'"},
			},
			Choices: []converter.ChoiceRow{
				{ListName: "geraete", Name: "smartphone", Label: "Smartphone"},
				{ListName: "geraete", Name: "laptop", Label: "Laptop"},
				{ListName: "geraete", Name: "tablet", Label: "Tablet"},
				{ListName: "geraete", Name: "other", Label: "Sonstiges"},
			},
		},
	},
	{
		Type:  "grid",
		Label: "Matrix / Likert-Skala — Institutionsvertrauen",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "begin_group", Name: "institutionsvertrauen", Label: "Vertrauen in Institutionen", Appearance: "table-list"},
				{Type: "select_one skala5", Name: "vertrauen_parlament", Label: "Das Parlament"},
				{Type: "select_one skala5", Name: "vertrauen_polizei", Label: "Die Polizei"},
				{Type: "end_group"},
			},
			Choices: []converter.ChoiceRow{
				{ListName: "skala5", Name: "1", Label: "Gar nicht"},
				{ListName: "skala5", Name: "2", Label: "2"},
				{ListName: "skala5", Name: "3", Label: "3"},
				{ListName: "skala5", Name: "4", Label: "4"},
				{ListName: "skala5", Name: "5", Label: "Vollständig"},
			},
		},
	},
	{
		Type:  "integer",
		Label: "Offene Zahl — Alter",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "integer", Name: "alter", Label: "Wie alt sind Sie?"},
			},
			Choices: []converter.ChoiceRow{},
		},
	},
	{
		Type:  "text",
		Label: "Offener Text — Anmerkungen",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "text", Name: "anmerkungen", Label: "Haben Sie weitere Anmerkungen?"},
			},
			Choices: []converter.ChoiceRow{},
		},
	},
	{
		Type:  "single_choice_long_list",
		Label: "Single Choice (Lange Liste) — Geburtsland",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "select_one_from_file iso_3166_1.csv", Name: "geburtsland", Label: "In welchem Land wurden Sie geboren?"},
			},
		},
	},
	{
		Type:  "multiple_choice_long_list",
		Label: "Multiple Choice (Lange Liste) — Besuchte Länder",
		XLSForm: converter.XLSForm{
			Survey: []converter.SurveyRow{
				{Type: "select_multiple_from_file iso_3166_1.csv", Name: "besuchte_laender", Label: "Welche dieser Länder haben Sie bereits besucht? Mehrere Antworten möglich."},
			},
		},
	},
}

// retryAfter is how long a failed populate is remembered before the next
// request tries again. Without it every request, while the sidecar is down,
// would queue for the lock and repeat all conversions.
const retryAfter = 30 * time.Second

// populate generates DDI for every example via the ddi-emitter sidecar. It
// runs lazily on first GetAll/GetByType so that `go test ./...` and `go run`
// don't panic at import time when the sidecar isn't up yet (the unit tests
// for other packages don't need examples to be populated).
//
// An example the converter rejects is logged and left out; retrying won't
// change the answer. If the sidecar is unavailable, populate stops at the first
// failure and returns an error wrapping converter.ErrConverterUnavailable.
func populate() ([]Example, error) {
	out := make([]Example, 0, len(defs))
	for _, d := range defs {
		xlsJSON, err := json.Marshal(d.XLSForm)
		if err != nil {
			return nil, fmt.Errorf("examples: failed to marshal XLSForm for %s: %w", d.Type, err)
		}
		ddiXML, err := converter.XLSFormToDDI(xlsJSON)
		if errors.Is(err, converter.ErrConverterUnavailable) {
			return nil, fmt.Errorf("examples: %w", err)
		}
		if err != nil {
			log.Printf("ERROR: example %s left out, DDI conversion rejected it: %v", d.Type, err)
			continue
		}
		out = append(out, Example{
			Type:    d.Type,
			Label:   d.Label,
			XLSForm: d.XLSForm,
			DDI:     string(ddiXML),
		})
	}
	return out, nil
}

var (
	cacheMu sync.Mutex
	cached  []Example
	loaded  bool
	lastErr error
	lastTry time.Time
)

// load returns the examples, populating them on first use. The mutex also
// makes concurrent first requests share one populate. A failure is kept for
// retryAfter, then the next call tries again, so /api/examples recovers once
// the sidecar is reachable.
func load() ([]Example, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if loaded {
		return cached, nil
	}
	if lastErr != nil && time.Since(lastTry) < retryAfter {
		return nil, lastErr
	}
	out, err := populate()
	lastTry = time.Now()
	if err != nil {
		log.Printf("ERROR: %v (retrying in %s)", err, retryAfter)
		lastErr = err
		return nil, err
	}
	cached, loaded, lastErr = out, true, nil
	return cached, nil
}

// reset forgets the cache; for tests.
func reset() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cached, loaded, lastErr, lastTry = nil, false, nil, time.Time{}
}

// GetAll returns all examples. The error wraps
// converter.ErrConverterUnavailable while the sidecar can't be reached.
func GetAll() ([]Example, error) {
	return load()
}

// GetByType returns a single example by type identifier, or nil if not found.
// The error is set only when the examples couldn't be loaded.
func GetByType(t string) (*Example, error) {
	all, err := load()
	if err != nil {
		return nil, err
	}
	for _, e := range all {
		if e.Type == t {
			return &e, nil
		}
	}
	return nil, nil
}
