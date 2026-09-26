package routes

import (
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/blevesearch/snowballstem"
	"github.com/blevesearch/snowballstem/english"
	"github.com/blevesearch/snowballstem/german"
	"github.com/pocketbase/pocketbase/core"
)

// stemEnv is reused across Stem calls; the blevesearch/snowballstem API
// mutates an Env in place rather than returning a value, so reusing one
// per goroutine is cheaper than allocating per token.
var stemEnvPool = sync.Pool{
	New: func() any { return snowballstem.NewEnv("") },
}

// fieldWeight is the relevance weight for each field, indexed by
// [question_text, concept, tags, name, answer_type]. Higher weight = ranked
// up when two questions tie on the number of matched terms. Tags weigh like
// the concept: they are further concepts (#19).
var fieldWeight = [5]int{4, 3, 3, 2, 1}

// tokenize splits the input on whitespace and commas, then drops empty
// fragments. Stays case-insensitive — normalisation runs separately.
func tokenize(s string) []string {
	f := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
	out := make([]string, 0, len(f))
	for _, t := range f {
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// normalize lowercases the input, folds common German umlauts plus a
// handful of other diacritics so that `Qualitaet` and `Qualität` match, and
// `naive` matches `naïve`. The mapping is deliberately small — extending it
// to every Latin-1 diacritic would be a maintenance liability for marginal
// gain. The folding happens on runes so multibyte UTF-8 stays valid.
//
// Trailing punctuation (`?`, `!`, `,`, `.`, `:`, `;`) and surrounding
// brackets/quotes are stripped so that `weiterempfehlen?` (the end of a
// question sentence) tokenises and stems the same way as `weiterempfehlen`.
func normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToLower(s) {
		switch r {
		case 'ä':
			b.WriteString("ae")
		case 'ö':
			b.WriteString("oe")
		case 'ü':
			b.WriteString("ue")
		case 'ß':
			b.WriteString("ss")
		case 'á', 'à', 'â', 'ã', 'å':
			b.WriteRune('a')
		case 'é', 'è', 'ê', 'ë':
			b.WriteRune('e')
		case 'í', 'ì', 'î', 'ï':
			b.WriteRune('i')
		case 'ó', 'ò', 'ô', 'õ', 'ø':
			b.WriteRune('o')
		case 'ú', 'ù', 'û':
			b.WriteRune('u')
		case 'ñ':
			b.WriteRune('n')
		case 'ç':
			b.WriteRune('c')
		case '?', '!', ',', '.', ':', ';', '(', ')', '[', ']', '{', '}', '"', '\'':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// stemVariants returns the set of forms (normalised + German stem +
// English stem) that this token should match on. Empty tokens are dropped.
// The German stemmer handles German morphology (`Zufriedenheit` →
// `zufried`); the English stemmer is a fallback for English tokens and for
// anything the German stemmer does not recognise. The bare normalised
// token is always included as the last resort so unstemmed words still
// hit.
func stemVariants(token string) []string {
	if token == "" {
		return nil
	}
	out := []string{token}
	env := stemEnvPool.Get().(*snowballstem.Env)
	defer stemEnvPool.Put(env)

	env.SetCurrent(token)
	german.Stem(env)
	if ge := env.Current(); ge != "" && ge != token {
		out = append(out, ge)
	}
	env.SetCurrent(token)
	english.Stem(env)
	if en := env.Current(); en != "" && en != token {
		out = append(out, en)
	}
	return out
}

// minStemLen is the shortest token length that participates in matching.
// Below this length the matching rules produce too many false positives —
// a 1-character field stem (`i` from a tokenised "I") would otherwise be a
// substring of nearly any longer query, and 2-letter stems collide with
// common English abbreviations.
const minStemLen = 3

// minInfixLen is the shortest query stem that may also match inside a longer
// field word. German builds compounds: `Zufriedenheit` must find
// "Arbeitszufriedenheit" and `Vertrauen` "Institutionenvertrauen", as the
// old whole-text substring search did. Short stems stay exact-only, since
// `sie` or `alt` would otherwise match half the corpus.
const minInfixLen = 5

// stemsMatch reports whether a query term matches a field word: some variant
// pair is equal, or a query variant of at least minInfixLen characters
// occurs inside a field variant.
//
// Stem equality covers inflection (Zufriedenheit ↔ zufrieden,
// Weiterempfehlung ↔ weiterempfehlen, Qualität ↔ Qualitaet after umlaut
// folding); the infix rule covers compounds.
func stemsMatch(queryVariants, fieldVariants []string) bool {
	for _, q := range queryVariants {
		if len(q) < minStemLen {
			continue
		}
		for _, f := range fieldVariants {
			if len(f) < minStemLen {
				continue
			}
			if q == f || (len(q) >= minInfixLen && strings.Contains(f, q)) {
				return true
			}
		}
	}
	return false
}

// stopwords are function words skipped in field text. Stems of content
// words can collide with them: the German stemmer reduces "Sicherheit" to
// `sich`, which then matched every question containing "sich". Only words
// of minStemLen or more letters need listing; shorter ones never match.
var stopwords = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(`
		der die das den dem des ein eine einen einem einer eines und oder aber
		sich sie ihr ihre ihren ihrem ihrer ihres ich wir uns euch wie was wer
		wann ist sind war waren hat haben wird werden wurde wurden mit fuer von
		auf aus bei nach seit ueber unter als auch nicht noch nur dass wenn sehr
		mehr man kann koennen einem diese dieser dieses
		the and for with from are was were been has have had does did you your
		our they their she its this that these those what which who when where
		how not yes than very more any`) {
		m[w] = true
	}
	return m
}()

// fieldVariants tokenises, normalises and stems a field once, so the result
// can be matched against every query term. Stopwords are skipped.
func fieldVariants(fieldText string) [][]string {
	var out [][]string
	for _, ft := range tokenize(fieldText) {
		if n := normalize(ft); n != "" && !stopwords[n] {
			out = append(out, stemVariants(n))
		}
	}
	return out
}

// fieldHit reports whether a query term matches any word of a field.
func fieldHit(field [][]string, queryVariants []string) bool {
	for _, fv := range field {
		if stemsMatch(queryVariants, fv) {
			return true
		}
	}
	return false
}

// fieldTexts returns the searchable text fields in fieldWeight order.
func (q Question) fieldTexts() [5]string {
	tags := make([]string, len(q.Tags))
	for i, t := range q.Tags {
		tags[i] = t.Text
	}
	return [5]string{q.QuestionText, q.Concept, strings.Join(tags, ", "), q.Name, q.AnswerType}
}

// FilterAndRankQuestions filters questions by the query and ranks them
// by relevance.
//
// Relevance has two tiers:
//  1. How many of the query's terms matched any field (terms matched,
//     descending). This is what makes `Wirkung Bildungsprogramm
//     Zufriedenheit` return the union of the three single-term hits
//     instead of an empty list.
//  2. Sum of field weights (question_text > concept = tags > name > answer_type)
//     across all (term, field) hits, breaking ties.
//
// Query terms are split on whitespace and commas. Both sides are
// lowercased, German umlauts folded (`ä/ö/ü/ß` → `ae/oe/ue/ss`), other
// diacritics stripped, and each token is stemmed with both the German and
// English Snowball stemmers (github.com/blevesearch/snowballstem). A term
// matches a word if the stems are equal or, for stems of minInfixLen or more,
// if it occurs inside the word (compounds); see stemsMatch.
func FilterAndRankQuestions(questions []Question, q string) []Question {
	tokens := tokenize(q)
	if len(tokens) == 0 {
		return nil
	}

	// Pre-compute stem variants for each query token (one per token).
	queryVariants := make([][]string, len(tokens))
	for i, tok := range tokens {
		queryVariants[i] = stemVariants(normalize(tok))
	}

	type scored struct {
		q           Question
		termsHit    int
		fieldWeight int
	}

	scoredAll := make([]scored, 0, len(questions))
	for _, question := range questions {
		termsHit := 0
		weighted := 0
		texts := question.fieldTexts()
		var fields [5][][]string
		for fi, ft := range texts {
			fields[fi] = fieldVariants(ft)
		}
		for _, qv := range queryVariants {
			termHitAny := false
			for fi, fv := range fields {
				if fieldHit(fv, qv) {
					termHitAny = true
					weighted += fieldWeight[fi]
				}
			}
			if termHitAny {
				termsHit++
			}
		}
		if termsHit == 0 {
			continue
		}
		scoredAll = append(scoredAll, scored{question, termsHit, weighted})
	}

	sort.SliceStable(scoredAll, func(i, j int) bool {
		if scoredAll[i].termsHit != scoredAll[j].termsHit {
			return scoredAll[i].termsHit > scoredAll[j].termsHit
		}
		return scoredAll[i].fieldWeight > scoredAll[j].fieldWeight
	})

	out := make([]Question, len(scoredAll))
	for i, s := range scoredAll {
		out[i] = s.q
	}
	return out
}

// SearchQuestions assembles the questions of the selected studies and ranks
// them against q (see FilterAndRankQuestions). include limits the search to
// these study IDs (all studies when empty); exclude leaves studies out, e.g.
// the demographic standards a caller offers separately. Shared by the REST
// endpoint and the MCP tool.
func SearchQuestions(app core.App, q string, include, exclude []string) ([]Question, error) {
	studies, err := app.FindRecordsByFilter("studies", "", "", 0, 0)
	if err != nil {
		return nil, err
	}
	keep := make(map[string]bool, len(include))
	for _, id := range include {
		keep[id] = true
	}
	drop := make(map[string]bool, len(exclude))
	for _, id := range exclude {
		drop[id] = true
	}

	var all []Question
	for _, s := range studies {
		if (len(keep) > 0 && !keep[s.Id]) || drop[s.Id] {
			continue
		}
		qs, err := AssembleQuestions(app, s.Id)
		if err != nil {
			continue
		}
		all = append(all, qs...)
	}
	return FilterAndRankQuestions(all, q), nil
}
