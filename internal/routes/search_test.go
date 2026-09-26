package routes

import (
	"reflect"
	"testing"
)

// allSearchFixtures is the shared input corpus for the tests in this file.
// It mirrors the answer-type variety the seed data covers, so each test
// case below exercises a representative slice.
func allSearchFixtures() []Question {
	return []Question{
		// German nominalisation: query stem finds field stem via shared root.
		{ID: "a", Name: "satisfaction", Concept: "Zufriedenheit", QuestionText: "Wie zufrieden sind Sie mit dem Angebot?", AnswerType: "single_choice"},
		{ID: "b", Name: "age", Concept: "Alter", QuestionText: "Wie alt sind Sie?", AnswerType: "integer"},
		{ID: "c", Name: "comments", Concept: "Anmerkungen", QuestionText: "Haben Sie weitere Anmerkungen?", AnswerType: "text"},
		// NPS via concept containing the literal term.
		{ID: "jq77owc2giqkryi", Name: "nps", Concept: "Net Promoter Score (Weiterempfehlung)", QuestionText: "Wie wahrscheinlich ist es, dass Sie einer Freundin oder Kollegin weiterempfehlen werden?", AnswerType: "single_choice"},
		// Gender with _other suffix (German), useful for word-form regression.
		{ID: "d", Name: "geschlecht", Concept: "Geschlecht (Selbstdefinition)", QuestionText: "Was ist Ihr Geschlecht?", AnswerType: "multiple_choice_other"},
		// English question with German keyword in concept for bilingual regression.
		{ID: "4epbcqti75mvbfz", Name: "neighbour_trust", Concept: "Interpersonal trust", QuestionText: "Do you think that your neighbours act in your best interests?", AnswerType: "single_choice"},
		{ID: "m54bffznrpq5qyb", Name: "council_trust", Concept: "Institutional trust", QuestionText: "Do you trust your local council to act in your best interest?", AnswerType: "single_choice"},
		// German trust study; kept in for symmetry — exercises German stemmer.
		{ID: "vt1", Name: "institutionsvertrauen", Concept: "Institutionenvertrauen", QuestionText: "Vertrauen in Institutionen", AnswerType: "single_choice"},
	}
}

func ids(qs []Question) []string {
	out := make([]string, len(qs))
	for i, q := range qs {
		out[i] = q.ID
	}
	return out
}

func TestFilterAndRankQuestions_NonsenseReturnsEmpty(t *testing.T) {
	got := FilterAndRankQuestions(allSearchFixtures(), "zzzznonexistentzzzz")
	if len(got) != 0 {
		t.Errorf("expected 0 matches, got %d: %v", len(got), ids(got))
	}
}

func TestFilterAndRankQuestions_GermanStemFindsInflection(t *testing.T) {
	// Issue test case: `Zufriedenheit` finds items with "zufrieden".
	// The German stemmer collapses both to `zufried`, so the question whose
	// field text contains "zufrieden" should match.
	got := FilterAndRankQuestions(allSearchFixtures(), "Zufriedenheit")
	if !reflect.DeepEqual(ids(got), []string{"a"}) {
		t.Errorf("expected [a], got %v", ids(got))
	}
}

func TestFilterAndRankQuestions_UmlautFolding(t *testing.T) {
	// Query `Qualitaet` should match a field containing "Qualität" because
	// umlaut folding turns both into "qualitaet".
	qs := []Question{
		{ID: "q1", Name: "qual", Concept: "Qualität der Arbeit", QuestionText: "Wie beurteilen Sie die Qualität unserer Arbeit?", AnswerType: "single_choice"},
		{ID: "q2", Name: "alter", Concept: "Alter", QuestionText: "Wie alt sind Sie?", AnswerType: "integer"},
	}
	got := FilterAndRankQuestions(qs, "Qualitaet")
	if !reflect.DeepEqual(ids(got), []string{"q1"}) {
		t.Errorf("expected [q1], got %v", ids(got))
	}
}

func TestFilterAndRankQuestions_MultiTermIsUnion(t *testing.T) {
	// Issue test case: `Wirkung Bildungsprogramm Zufriedenheit` returns the
	// union of the single-term hits. `wirkung` matches two terms (Wirkung in
	// concept and name, Bildungsprogramm in question_text), `zfr` one
	// (Zufriedenheit), so `wirkung` ranks first; `andere` matches none.
	qs := []Question{
		{ID: "wirkung", Name: "wirkung", Concept: "Wirkung des Projekts", QuestionText: "Wie wirksam war das Bildungsprogramm?", AnswerType: "single_choice"},
		{ID: "zfr", Name: "zfr", Concept: "Zufriedenheit", QuestionText: "Wie zufrieden sind Sie?", AnswerType: "single_choice"},
		{ID: "andere", Name: "andere", Concept: "Alter", QuestionText: "Wie alt sind Sie?", AnswerType: "integer"},
	}
	got := ids(FilterAndRankQuestions(qs, "Wirkung Bildungsprogramm Zufriedenheit"))
	want := []string{"wirkung", "zfr"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v (sorted by termsHit, then fieldWeight), got %v", want, got)
	}
}

func TestFilterAndRankQuestions_FindsCompounds(t *testing.T) {
	// German compounds: the old whole-text substring search found these, and
	// stem equality alone would not.
	qs := []Question{
		{ID: "arbeit", Name: "jobsat", Concept: "Arbeitszufriedenheit", QuestionText: "Wie bewerten Sie Ihre Arbeitssituation?", AnswerType: "single_choice"},
		{ID: "vt1", Name: "inst", Concept: "Institutionenvertrauen", QuestionText: "Wie sehr trauen Sie Behörden?", AnswerType: "single_choice"},
	}
	if got := ids(FilterAndRankQuestions(qs, "Zufriedenheit")); !reflect.DeepEqual(got, []string{"arbeit"}) {
		t.Errorf("Zufriedenheit: expected [arbeit], got %v", got)
	}
	if got := ids(FilterAndRankQuestions(qs, "Vertrauen")); !reflect.DeepEqual(got, []string{"vt1"}) {
		t.Errorf("Vertrauen: expected [vt1], got %v", got)
	}
}

func TestFilterAndRankQuestions_ShortStemsStayExact(t *testing.T) {
	// A short term must not match inside longer words: `alt` finds
	// "Wie alt sind Sie?" but not "Verwaltung".
	qs := []Question{
		{ID: "age", Name: "age", Concept: "Alter", QuestionText: "Wie alt sind Sie?", AnswerType: "integer"},
		{ID: "admin", Name: "admin", Concept: "Verwaltung", QuestionText: "Wie bewerten Sie die Verwaltung?", AnswerType: "single_choice"},
	}
	if got := ids(FilterAndRankQuestions(qs, "alt")); !reflect.DeepEqual(got, []string{"age"}) {
		t.Errorf("expected [age], got %v", got)
	}
}

func TestFilterAndRankQuestions_NPSConceptLiteral(t *testing.T) {
	// Issue test case: `Weiterempfehlung` finds the NPS question via its
	// concept text — the German noun from the verb "weiterempfehlen".
	got := FilterAndRankQuestions(allSearchFixtures(), "Weiterempfehlung")
	if !reflect.DeepEqual(ids(got), []string{"jq77owc2giqkryi"}) {
		t.Errorf("expected [jq77owc2giqkryi], got %v", ids(got))
	}
}

func TestFilterAndRankQuestions_EnglishSubstringMatchesTrust(t *testing.T) {
	// Issue test case (the easy half): `trust` finds both trust questions
	// via the English stem "trust" / direct substring of the concept.
	// Order is set by tie-break: both match in concept with weight 3; the
	// `council_trust` question also matches `trust` as a name token, so it
	// wins on field-weight sum (3 + 2 = 5 vs 3).
	got := ids(FilterAndRankQuestions(allSearchFixtures(), "trust"))
	want := []string{"m54bffznrpq5qyb", "4epbcqti75mvbfz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestFilterAndRankQuestions_RanksMoreTermsHigher(t *testing.T) {
	// A question matching more query terms ranks higher than one matching
	// just one, regardless of which field it hit.
	qs := []Question{
		{ID: "only_concept", Name: "x", Concept: "Wirkung", QuestionText: "Wie alt sind Sie?", AnswerType: "integer"},
		{ID: "both", Name: "x", Concept: "Wirkung Bildungsprogramm", QuestionText: "Wie zufrieden mit dem Bildungsprogramm?", AnswerType: "single_choice"},
	}
	got := FilterAndRankQuestions(qs, "Wirkung Bildungsprogramm")
	want := []string{"both", "only_concept"}
	if !reflect.DeepEqual(ids(got), want) {
		t.Errorf("expected %v, got %v", want, ids(got))
	}
}

func TestFilterAndRankQuestions_RanksByFieldWeightOnTies(t *testing.T) {
	// Two questions, each matching the same single term but in different
	// fields. The one matching question_text should rank above the one
	// matching only `concept` (which is the next-priority field).
	qs := []Question{
		{ID: "in_concept", Name: "x", Concept: "Wirkung", QuestionText: "Wie alt sind Sie?", AnswerType: "integer"},
		{ID: "in_question", Name: "x", Concept: "Alter", QuestionText: "Welche Wirkung hatte das Projekt?", AnswerType: "single_choice"},
	}
	got := FilterAndRankQuestions(qs, "Wirkung")
	want := []string{"in_question", "in_concept"}
	if !reflect.DeepEqual(ids(got), want) {
		t.Errorf("expected %v, got %v", want, ids(got))
	}
}

func TestFilterAndRankQuestions_StopwordsIgnored(t *testing.T) {
	// Short field stems (`I`, `a`) should not match the query just
	// because the letter appears in it. Pure regression test for the
	// bug where a 1-character field token leaked through and matched any
	// longer query that contained that letter.
	qs := []Question{
		{ID: "noisy", Name: "x", Concept: "Trust", QuestionText: "I have a dog", AnswerType: "single_choice"},
	}
	got := FilterAndRankQuestions(qs, "zzzznonexistentzzzz")
	if len(got) != 0 {
		t.Errorf("expected 0 matches for nonsense query, got %v", ids(got))
	}
}

func TestFilterAndRankQuestions_StemHandlesWeiterempfehlen(t *testing.T) {
	// "weiterempfehlen" (verb) and "Weiterempfehlung" (nominalised noun)
	// share the stem "weiterempfehl" — both should match.
	qs := []Question{
		{ID: "verb", Name: "x", Concept: "Bereitschaft weiterzuempfehlen", QuestionText: "Würden Sie das Angebot weiterempfehlen?", AnswerType: "single_choice"},
	}
	got := FilterAndRankQuestions(qs, "weiterempfehlen")
	if !reflect.DeepEqual(ids(got), []string{"verb"}) {
		t.Errorf("expected [verb], got %v", ids(got))
	}
	got = FilterAndRankQuestions(qs, "Weiterempfehlung")
	if !reflect.DeepEqual(ids(got), []string{"verb"}) {
		t.Errorf("expected [verb], got %v", ids(got))
	}
}

// #19: tags (the <concept> elements after the first) make an item findable
// in the other language; the original wording stays as it is.
func TestFilterAndRankQuestions_MatchesTags(t *testing.T) {
	qs := allSearchFixtures()
	for i := range qs {
		if qs[i].ID == "4epbcqti75mvbfz" {
			qs[i].Tags = []QuestionTag{{Lang: "de", Text: "Vertrauen"}, {Lang: "de", Text: "Nachbarschaft"}}
		}
	}
	got := idsSet(FilterAndRankQuestions(qs, "Vertrauen"))
	if !got["4epbcqti75mvbfz"] || !got["vt1"] {
		t.Errorf("Vertrauen: expected the tagged English item and the German one, got %v", got)
	}
	if got["m54bffznrpq5qyb"] {
		t.Errorf("Vertrauen: the untagged English item must not match, got %v", got)
	}
	if got := ids(FilterAndRankQuestions(qs, "Nachbarschaft")); !reflect.DeepEqual(got, []string{"4epbcqti75mvbfz"}) {
		t.Errorf("Nachbarschaft: expected [4epbcqti75mvbfz], got %v", got)
	}
}

// Stems can collide with function words: "Sicherheit" stems to `sich`, which
// used to match every question containing "sich".
func TestFilterAndRankQuestions_StopwordsDontMatchStems(t *testing.T) {
	qs := []Question{
		{ID: "sich", Name: "gruendung", Concept: "Gründungsjahr", QuestionText: "Wann hat sich Ihre Organisation gegründet?", AnswerType: "integer"},
		{ID: "safe", Name: "sicherheit", Concept: "Sicherheit im Stadtteil", QuestionText: "Wie beurteilen Sie die Sicherheit in Ihrem Stadtteil?", AnswerType: "single_choice"},
	}
	if got := ids(FilterAndRankQuestions(qs, "Sicherheit")); !reflect.DeepEqual(got, []string{"safe"}) {
		t.Errorf("expected [safe], got %v", got)
	}
}

func idsSet(qs []Question) map[string]bool {
	out := map[string]bool{}
	for _, q := range qs {
		out[q.ID] = true
	}
	return out
}
