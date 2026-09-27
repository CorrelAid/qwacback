// Package questiontypes serves the registry's question-type catalogue
// (formtransform's QUESTION_TYPES) keyed by qwacback's answer types, so
// clients follow qwacback's formtransform pin instead of importing the
// package themselves (#40).
package questiontypes

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"qwacback/internal/converter"
)

// Entry is one registry question type, as QUESTION_TYPES has it, with the
// label per language.
type Entry struct {
	ID          string            `json:"id"`
	Label       map[string]string `json:"label"`
	UseWhen     string            `json:"useWhen,omitempty"`
	Kind        string            `json:"kind"`
	TypeString  string            `json:"typeString,omitempty"`
	Base        string            `json:"base,omitempty"`
	Bases       []string          `json:"bases,omitempty"`
	IsVariant   bool              `json:"isVariant"`
	IsComposite bool              `json:"isComposite"`
	Aliases     []string          `json:"aliases,omitempty"`
	Constraints json.RawMessage   `json:"constraints,omitempty"`
}

// registryEntry is an entry as QUESTION_TYPES serializes it: one English
// label (skos:prefLabel has no language yet, formtransform#161).
type registryEntry struct {
	Entry
	Label string `json:"label"`
}

// Presentation is how a client renders the answer type.
type Presentation struct {
	// Choice is "one" or "multiple" for selects, "" otherwise.
	Choice    string `json:"choice,omitempty"`
	WithOther bool   `json:"withOther"`
	LongList  bool   `json:"longList"`
	Grid      bool   `json:"grid"`
}

// AnswerType is one of qwacback's answer types with its registry type.
type AnswerType struct {
	RegistryType string            `json:"registryType"`
	Label        map[string]string `json:"label"`
	Kind         string            `json:"kind"`
	Base         string            `json:"base,omitempty"`
	Aliases      []string          `json:"aliases,omitempty"`
	Presentation Presentation      `json:"presentation"`
}

// answerTypes maps qwacback's answer types (variables.answer_type plus the
// _other / _long_list suffixes, see routes.effectiveAnswerType) to registry
// slugs and to how they are presented. The presentation is derived here
// until QUESTION_TYPES exports it (formtransform#161).
var answerTypes = map[string]struct {
	slug string
	p    Presentation
}{
	"single_choice":             {"select_one", Presentation{Choice: "one"}},
	"multiple_choice":           {"select_multiple", Presentation{Choice: "multiple"}},
	"single_choice_other":       {"select_one_other", Presentation{Choice: "one", WithOther: true}},
	"multiple_choice_other":     {"select_multiple_other", Presentation{Choice: "multiple", WithOther: true}},
	"single_choice_long_list":   {"select_one_long_list", Presentation{Choice: "one", LongList: true}},
	"multiple_choice_long_list": {"select_multiple_long_list", Presentation{Choice: "multiple", LongList: true}},
	"grid":                      {"grid", Presentation{Choice: "one", Grid: true}},
	"integer":                   {"integer", Presentation{}},
	"text":                      {"text", Presentation{}},
}

// retryAfter is how long a failed load is remembered before the next request
// tries again, as in the examples package.
const retryAfter = 30 * time.Second

var (
	mu       sync.Mutex
	registry map[string]Entry
	byAnswer map[string]AnswerType
	lastErr  error
	lastTry  time.Time
)

// load fetches the catalogue from the sidecar once; it can't change while
// qwacback runs, since the pin is fixed per deploy.
func load() error {
	mu.Lock()
	defer mu.Unlock()
	if registry != nil {
		return nil
	}
	if lastErr != nil && time.Since(lastTry) < retryAfter {
		return lastErr
	}
	lastTry = time.Now()
	reg, ans, err := fetch()
	if err != nil {
		log.Printf("ERROR: question types: %v (retrying in %s)", err, retryAfter)
		lastErr = err
		return err
	}
	registry, byAnswer, lastErr = reg, ans, nil
	return nil
}

func fetch() (map[string]Entry, map[string]AnswerType, error) {
	raw, err := converter.QuestionTypes()
	if err != nil {
		return nil, nil, err
	}
	return build(raw)
}

func build(raw []byte) (map[string]Entry, map[string]AnswerType, error) {
	var entries map[string]registryEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, nil, fmt.Errorf("%w: QUESTION_TYPES: %v", converter.ErrConverterUnavailable, err)
	}
	reg := make(map[string]Entry, len(entries))
	for slug, e := range entries {
		entry := e.Entry
		entry.Label = map[string]string{"en": e.Label}
		reg[slug] = entry
	}
	ans := make(map[string]AnswerType, len(answerTypes))
	for name, at := range answerTypes {
		e, ok := reg[at.slug]
		if !ok {
			// The registry dropped or renamed a type qwacback hands out:
			// leave it out rather than invent an entry.
			log.Printf("WARNING: question types: registry has no %q for answer type %q", at.slug, name)
			continue
		}
		ans[name] = AnswerType{
			RegistryType: at.slug,
			Label:        e.Label,
			Kind:         e.Kind,
			Base:         e.Base,
			Aliases:      e.Aliases,
			Presentation: at.p,
		}
	}
	return reg, ans, nil
}

// ByAnswerType returns the catalogue keyed by qwacback's answer types. The
// error wraps converter.ErrConverterUnavailable while the sidecar can't be
// reached.
func ByAnswerType() (map[string]AnswerType, error) {
	if err := load(); err != nil {
		return nil, err
	}
	return byAnswer, nil
}

// Registry returns every registry type, keyed by registry slug.
func Registry() (map[string]Entry, error) {
	if err := load(); err != nil {
		return nil, err
	}
	return registry, nil
}

// reset forgets the cache; for tests.
func reset() {
	mu.Lock()
	defer mu.Unlock()
	registry, byAnswer, lastErr, lastTry = nil, nil, nil, time.Time{}
}
