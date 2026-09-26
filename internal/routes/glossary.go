package routes

// The question bank mixes English studies with German ones, and the German
// demographic standards carry English concepts. A German keyword would only
// find German text and an English keyword only English text, so the search
// also tries each query word's translations from this glossary (#19).
//
// Each entry lists the German and the English words for one idea, one word
// per string: a phrase would have to be split, and its parts ("group",
// "status") would match far too much. Words are
// matched by stem, so one inflected form per word is enough (`Vertrauen`
// covers "vertraue", `trust` covers "trusted"). A query word that matches a
// word on one side is also searched as every word on the other side; the hit
// counts as a hit for that query term.
//
// Keep entries to survey vocabulary that actually occurs in questions or
// concepts, and keep each side specific: a broad word ("Frage", "question")
// would expand to half the bank.
var glossary = []struct {
	de []string
	en []string
}{
	// Trust, social capital, community
	{[]string{"Vertrauen"}, []string{"trust"}},
	{[]string{"Nachbarschaft", "Wohnviertel", "Stadtteil"}, []string{"neighbourhood", "neighborhood"}},
	{[]string{"Nachbarn"}, []string{"neighbours", "neighbors"}},
	{[]string{"Gemeinschaft"}, []string{"community"}},
	{[]string{"Freunde", "Freundschaft"}, []string{"friends", "friendship"}},
	{[]string{"Hilfe", "Unterstützung"}, []string{"help", "support"}},
	{[]string{"Gegenseitigkeit"}, []string{"reciprocity"}},
	{[]string{"Netzwerk"}, []string{"network"}},
	{[]string{"Kontakt", "Ansprechperson"}, []string{"contact"}},
	{[]string{"Zusammenhalt"}, []string{"cohesion"}},
	{[]string{"Einsamkeit"}, []string{"loneliness"}},
	{[]string{"Zugehörigkeit"}, []string{"belonging"}},

	// Participation, empowerment, efficacy
	{[]string{"Beteiligung", "Teilhabe", "Partizipation", "Teilnahme"}, []string{"participation", "participate"}},
	{[]string{"Mitbestimmung", "Entscheidung"}, []string{"decision", "empowerment"}},
	{[]string{"Selbstwirksamkeit"}, []string{"self-efficacy", "efficacy"}},
	{[]string{"Einstellungen", "Haltung"}, []string{"attitudes"}},
	{[]string{"Engagement", "engagieren", "Ehrenamt", "ehrenamtlich"}, []string{"volunteering", "volunteer", "engagement"}},
	{[]string{"Stadtverwaltung", "Gemeinderat", "Kommune", "Behörden"}, []string{"council", "authorities"}},

	// Place, safety
	{[]string{"Sicherheit", "sicher"}, []string{"safety", "safe"}},
	{[]string{"Kriminalität"}, []string{"crime"}},
	{[]string{"Attraktivität", "attraktiv"}, []string{"attractiveness", "attractive"}},
	{[]string{"Einrichtung", "Treffpunkt"}, []string{"facility", "space"}},
	{[]string{"Nutzung", "nutzen"}, []string{"usage", "use"}},
	{[]string{"Häufigkeit", "häufig"}, []string{"frequency", "often"}},
	{[]string{"Wohndauer"}, []string{"residence"}},
	{[]string{"Wohneigentum", "Eigentum"}, []string{"ownership", "homeowner"}},

	// Evaluation constructs
	{[]string{"Zufriedenheit", "zufrieden"}, []string{"satisfaction", "satisfied"}},
	{[]string{"Weiterempfehlung", "weiterempfehlen"}, []string{"recommendation", "recommend"}},
	{[]string{"Wirkung", "Auswirkung"}, []string{"impact", "effect"}},
	{[]string{"Qualität"}, []string{"quality"}},
	{[]string{"Wissen", "Kenntnisse"}, []string{"knowledge"}},
	{[]string{"Kompetenzen", "Fähigkeiten"}, []string{"skills", "competencies"}},
	{[]string{"Motivation"}, []string{"motivation"}},
	{[]string{"Gesundheit"}, []string{"health"}},
	{[]string{"Wohlbefinden"}, []string{"wellbeing", "well-being"}},
	{[]string{"Anmerkungen", "Kommentare"}, []string{"comments"}},
	{[]string{"Sprache"}, []string{"language"}},

	// Organisations (SVR study)
	{[]string{"Organisation"}, []string{"organisation", "organization"}},
	{[]string{"Verein"}, []string{"association", "club"}},
	{[]string{"Stiftung"}, []string{"foundation"}},
	{[]string{"Mitglieder", "Mitgliedschaft"}, []string{"members", "membership"}},
	{[]string{"Mitgliedsbeiträge"}, []string{"fees"}},
	{[]string{"Vorstand"}, []string{"board"}},
	{[]string{"Geschäftsführung"}, []string{"management", "director"}},
	{[]string{"Gründungsjahr", "Gründung"}, []string{"founding", "founded"}},
	{[]string{"Förderung", "Fördermittel"}, []string{"funding", "grants"}},
	{[]string{"Spenden"}, []string{"donations"}},
	{[]string{"Einnahmen"}, []string{"revenue"}},
	{[]string{"Geflüchtete", "Flüchtlinge"}, []string{"refugees"}},
	{[]string{"Migrationshintergrund"}, []string{"migrant", "migrants"}},
	{[]string{"Religion", "Religionszugehörigkeit"}, []string{"religion", "religious"}},
	{[]string{"Nutzergruppen", "Zielgruppe"}, []string{"users"}},

	// Demographics
	{[]string{"Geschlecht"}, []string{"gender", "sex"}},
	{[]string{"Alter", "Altersgruppe"}, []string{"age"}},
	{[]string{"Geburtsjahr", "geboren"}, []string{"birth", "born"}},
	{[]string{"Staatsangehörigkeit", "Staatsbürgerschaft"}, []string{"citizenship", "nationality"}},
	{[]string{"Familienstand"}, []string{"marital"}},
	{[]string{"Haushalt"}, []string{"household"}},
	{[]string{"Einkommen", "Nettoeinkommen"}, []string{"income"}},
	{[]string{"Schulabschluss", "Bildungsabschluss", "Bildung"}, []string{"education", "school"}},
	{[]string{"Ausbildung", "Berufsausbildung"}, []string{"vocational"}},
	{[]string{"Erwerbstätigkeit", "Erwerbssituation", "Beschäftigung"}, []string{"employment"}},
	{[]string{"Beruf", "Tätigkeit"}, []string{"occupation", "occupational"}},
	{[]string{"Arbeitszeit", "Arbeitsstunden"}, []string{"hours"}},
	{[]string{"Arbeitserlaubnis"}, []string{"permit"}},
	{[]string{"Internetnutzung", "Internet"}, []string{"internet"}},
	{[]string{"Handy", "Smartphone"}, []string{"mobile", "smartphone"}},
}

// glossaryIndex maps each stem of a glossary word to the words on the other
// side of its entry. Built once from glossary.
var glossaryIndex = buildGlossaryIndex()

func buildGlossaryIndex() map[string][]string {
	idx := map[string][]string{}
	add := func(from, to []string) {
		for _, w := range from {
			for _, v := range stemVariants(normalize(w)) {
				if len(v) >= minStemLen {
					idx[v] = append(idx[v], to...)
				}
			}
		}
	}
	for _, e := range glossary {
		add(e.de, e.en)
		add(e.en, e.de)
	}
	return idx
}

// translations returns the glossary words in the other language for a
// normalised query word, matched by stem.
func translations(word string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range stemVariants(word) {
		for _, t := range glossaryIndex[v] {
			if !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	return out
}
