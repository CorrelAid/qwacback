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
// Type is the formtransform registry slug (select_one, grid, …), the name the
// converters, /api/question-types and the DDI output use (#49).
type Example struct {
	Type    string  `json:"answer_type"`
	Label   string  `json:"label"`
	XLSForm XLSForm `json:"xlsform"`
	DDI     string  `json:"ddi"`
}

// XLSForm is an example form's survey and choices sheets, with the columns
// the examples use.
type XLSForm struct {
	Survey  []SurveyRow `json:"survey"`
	Choices []ChoiceRow `json:"choices"`
}

type SurveyRow struct {
	Type       string `json:"type"`
	Name       string `json:"name,omitempty"`
	Label      string `json:"label,omitempty"`
	Relevance  string `json:"relevant,omitempty"`
	Appearance string `json:"appearance,omitempty"`
}

type ChoiceRow struct {
	ListName string `json:"list_name"`
	Name     string `json:"name"`
	Label    string `json:"label"`
}

// exampleDef holds the static definition before DDI generation.
type exampleDef struct {
	Type    string
	Label   string
	XLSForm XLSForm
}

// legacyIDs maps the ids examples had before they were keyed by registry slug
// (#49) to the slug. They stay served as aliases until the survey site
// (CorrelAid/cdl-wp-eins#30) has switched; then this map goes.
var legacyIDs = map[string]string{
	"single_choice":             "select_one",
	"multiple_choice":           "select_multiple",
	"single_choice_other":       "select_one_other",
	"multiple_choice_other":     "select_multiple_other",
	"single_choice_long_list":   "select_one_long_list",
	"multiple_choice_long_list": "select_multiple_long_list",
}

var defs = []exampleDef{
	{
		Type:  "select_one",
		Label: "Single Choice — Bildungsgrad",
		XLSForm: XLSForm{
			Survey: []SurveyRow{
				{Type: "select_one bildungsgrad", Name: "bildungsgrad", Label: "Was ist Ihr höchster Bildungsabschluss?"},
			},
			Choices: []ChoiceRow{
				{ListName: "bildungsgrad", Name: "1", Label: "Kein Abschluss"},
				{ListName: "bildungsgrad", Name: "2", Label: "Haupt- oder Realschulabschluss"},
				{ListName: "bildungsgrad", Name: "3", Label: "Fachhochschulreife / Abitur"},
				{ListName: "bildungsgrad", Name: "4", Label: "Abgeschlossene Berufsausbildung"},
				{ListName: "bildungsgrad", Name: "5", Label: "Hochschulabschluss"},
			},
		},
	},
	{
		Type:  "select_multiple",
		Label: "Multiple Choice — Wochenendtage",
		XLSForm: XLSForm{
			Survey: []SurveyRow{
				{Type: "select_multiple wochenendtage", Name: "wochenende", Label: "An welchen Tagen des Wochenendes sind Sie erreichbar?"},
			},
			Choices: []ChoiceRow{
				{ListName: "wochenendtage", Name: "sa", Label: "Samstag"},
				{ListName: "wochenendtage", Name: "so", Label: "Sonntag"},
			},
		},
	},
	{
		Type:  "select_one_other",
		Label: "Single Choice mit Sonstiges — Aufmerksamkeitsquelle",
		XLSForm: XLSForm{
			Survey: []SurveyRow{
				{Type: "select_one quelle", Name: "aufmerksam", Label: "Wie sind Sie auf unser Angebot aufmerksam geworden?"},
				{Type: "text", Name: "aufmerksam_other", Label: "Sonstiges (bitte angeben)", Relevance: "${aufmerksam} = 'other'"},
			},
			Choices: []ChoiceRow{
				{ListName: "quelle", Name: "suchmaschine", Label: "Suchmaschine"},
				{ListName: "quelle", Name: "empfehlung", Label: "Persönliche Empfehlung"},
				{ListName: "quelle", Name: "soziale_medien", Label: "Soziale Medien"},
				{ListName: "quelle", Name: "other", Label: "Sonstiges"},
			},
		},
	},
	{
		Type:  "select_multiple_other",
		Label: "Multiple Choice mit Sonstiges — Gerätebesitz",
		XLSForm: XLSForm{
			Survey: []SurveyRow{
				{Type: "select_multiple geraete", Name: "geraetebesitz", Label: "Welche dieser Geräte besitzen Sie?"},
				{Type: "text", Name: "geraetebesitz_other", Label: "Sonstiges (bitte angeben)", Relevance: "selected(${geraetebesitz}, 'other')"},
			},
			Choices: []ChoiceRow{
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
		XLSForm: XLSForm{
			Survey: []SurveyRow{
				{Type: "begin_group", Name: "institutionsvertrauen", Label: "Vertrauen in Institutionen", Appearance: "table-list"},
				{Type: "select_one skala5", Name: "vertrauen_parlament", Label: "Das Parlament"},
				{Type: "select_one skala5", Name: "vertrauen_polizei", Label: "Die Polizei"},
				{Type: "end_group"},
			},
			Choices: []ChoiceRow{
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
		XLSForm: XLSForm{
			Survey: []SurveyRow{
				{Type: "integer", Name: "alter", Label: "Wie alt sind Sie?"},
			},
			Choices: []ChoiceRow{},
		},
	},
	{
		Type:  "text",
		Label: "Offener Text — Anmerkungen",
		XLSForm: XLSForm{
			Survey: []SurveyRow{
				{Type: "text", Name: "anmerkungen", Label: "Haben Sie weitere Anmerkungen?"},
			},
			Choices: []ChoiceRow{},
		},
	},
	{
		Type:  "select_one_long_list",
		Label: "Single Choice (Lange Liste) — Geburtsland",
		XLSForm: XLSForm{
			Survey: []SurveyRow{
				{Type: "select_one_from_file iso_3166_1.csv", Name: "geburtsland", Label: "In welchem Land wurden Sie geboren?"},
			},
		},
	},
	{
		Type:  "select_multiple_long_list",
		Label: "Multiple Choice (Lange Liste) — Besuchte Länder",
		XLSForm: XLSForm{
			Survey: []SurveyRow{
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

// GetByType returns a single example by registry slug, or nil if not found.
// A legacy id (single_choice, …) resolves to its slug. The error is set only
// when the examples couldn't be loaded.
func GetByType(t string) (*Example, error) {
	all, err := load()
	if err != nil {
		return nil, err
	}
	if slug, ok := legacyIDs[t]; ok {
		t = slug
	}
	for _, e := range all {
		if e.Type == t {
			return &e, nil
		}
	}
	return nil, nil
}
