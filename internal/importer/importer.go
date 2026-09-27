package importer

import (
	"bytes"
	"crypto/sha256"
	"encoding/base32"
	"encoding/xml"
	"log"
	"regexp"
	"strings"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
)

// deterministicID generates a stable 15-character PocketBase record ID
// from a composite key (e.g. "studyIDNo/varName"). The same input always
// produces the same ID.
func deterministicID(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "/")))
	// Base32 lowercase, no padding, take first 15 chars
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(h[:])
	return strings.ToLower(encoded[:15])
}

var abstractRe = regexp.MustCompile(`(?s)<abstract[^>]*>(.*?)</abstract>`)

// xmlLang returns an element's xml:lang attribute, or "".
func xmlLang(attrs []xml.Attr) string {
	for _, a := range attrs {
		if a.Name.Local == "lang" && a.Name.Space == "http://www.w3.org/XML/1998/namespace" {
			return a.Value
		}
	}
	return ""
}

// extractVarQstnLits walks the raw XML token stream and returns a map of
// variable DDI ID → full qstnLit text content. It uses encoding/xml directly
// instead of mxj because mxj cannot represent XML mixed content: text nodes
// interleaved with child elements (e.g. "text <em>bold</em> more text") lose
// the "tail" text that follows each closing tag.
//
// A qstnLit may repeat once per language (formtransform#135). The base
// language wins: the first qstnLit without xml:lang or with the
// codeBook's xml:lang, else the first one (#30).
//
// The second map holds the other languages: var ID → xml:lang → text (#35).
func extractVarQstnLits(rawXML []byte) (map[string]string, map[string]map[string]string) {
	result := make(map[string]string)
	others := make(map[string]map[string]string)
	isBase := make(map[string]bool) // var ID → result holds a base-language text
	var baseLang, curLang string
	decoder := xml.NewDecoder(bytes.NewReader(rawXML))
	decoder.Strict = true

	var currentVarID string
	inVar, inQstn, inQstnLit := false, false, false
	qstnLitDepth := 0
	var sb strings.Builder

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "codeBook":
				baseLang = xmlLang(t.Attr)
			case "var":
				inVar = true
				for _, attr := range t.Attr {
					if attr.Name.Local == "ID" {
						currentVarID = attr.Value
						break
					}
				}
			case "qstn":
				if inVar {
					inQstn = true
				}
			case "qstnLit":
				if inQstn {
					inQstnLit = true
					qstnLitDepth = 1
					curLang = xmlLang(t.Attr)
					sb.Reset()
				}
			default:
				if inQstnLit {
					qstnLitDepth++
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "var":
				inVar, inQstn = false, false
				currentVarID = ""
			case "qstn":
				inQstn = false
			case "qstnLit":
				if inQstnLit {
					qstnLitDepth--
					if qstnLitDepth == 0 {
						base := curLang == "" || curLang == baseLang
						_, have := result[currentVarID]
						text := strings.TrimSpace(sb.String())
						if currentVarID != "" && !isBase[currentVarID] && (base || !have) {
							result[currentVarID] = text
							isBase[currentVarID] = base
						} else if currentVarID != "" && !base && text != "" {
							if others[currentVarID] == nil {
								others[currentVarID] = map[string]string{}
							}
							others[currentVarID][curLang] = text
						}
						inQstnLit = false
					}
				}
			default:
				if inQstnLit {
					qstnLitDepth--
				}
			}
		case xml.CharData:
			if inQstnLit {
				sb.Write(t)
			}
		}
	}
	return result, others
}

// extractText recursively extracts plain text from an mxj value, handling
// nested XHTML elements (e.g. qstnLit with xmlns:xhtml markup). It collects
// #text nodes and recurses into child elements, skipping namespace
// declarations (keys prefixed with "-").
func extractText(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		var sb strings.Builder
		if text, ok := val["#text"].(string); ok {
			sb.WriteString(text)
		}
		for k, child := range val {
			if k == "#text" || strings.HasPrefix(k, "-") {
				continue
			}
			sb.WriteString(extractText(child))
		}
		return sb.String()
	case []interface{}:
		var sb strings.Builder
		for _, item := range val {
			sb.WriteString(extractText(item))
		}
		return sb.String()
	}
	return ""
}

// textAt extracts the text content of an mxj path, handling elements that
// have XML attributes or XHTML child content. mxj represents
// <foo bar="x">text</foo> as {"-bar":"x","#text":"text"}, so we try the
// #text sub-key first, then fall back to ValuesForPath which returns the
// typed value (rather than a stringified map representation).
func textAt(mv mxj.Map, path string) string {
	if v, err := mv.ValueForPathString(path + ".#text"); err == nil && v != "" {
		return v
	}
	vals, err := mv.ValuesForPath(path)
	if err != nil || len(vals) == 0 {
		return ""
	}
	return strings.TrimSpace(extractText(vals[0]))
}

// textAtLang is textAt for elements that may repeat once per language
// (preQTxt, ivuInstr, labl, txt; formtransform#135). It returns the first
// element without xml:lang or with the base language (the codeBook's
// xml:lang), else the first element. textAt would return the last (#30).
func textAtLang(mv mxj.Map, path, baseLang string) string {
	vals, err := mv.ValuesForPath(path)
	if err != nil || len(vals) == 0 {
		return ""
	}
	pick := vals[0]
	for _, v := range vals {
		lang := ""
		if m, ok := v.(map[string]interface{}); ok {
			lang, _ = m["-lang"].(string)
		}
		if lang == "" || lang == baseLang {
			pick = v
			break
		}
	}
	return strings.TrimSpace(extractText(pick))
}

// translationsAt returns the other-language versions of a repeatable
// element (formtransform#135): xml:lang → text for every element tagged
// with a language other than the base (#35).
func translationsAt(mv mxj.Map, path, baseLang string) map[string]string {
	vals, err := mv.ValuesForPath(path)
	if err != nil {
		return nil
	}
	out := map[string]string{}
	for _, v := range vals {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		lang, _ := m["-lang"].(string)
		if lang == "" || lang == baseLang {
			continue
		}
		if text := strings.TrimSpace(extractText(v)); text != "" {
			out[lang] = text
		}
	}
	return out
}

// translations collects the other-language texts of one record, as stored
// in its `translations` field: lang → field → text (or, for categories,
// value → label).
type translations map[string]map[string]interface{}

func (t translations) add(lang, field string, text string) {
	if t[lang] == nil {
		t[lang] = map[string]interface{}{}
	}
	t[lang][field] = text
}

func (t translations) addCategory(lang, value, label string) {
	if t[lang] == nil {
		t[lang] = map[string]interface{}{}
	}
	cats, _ := t[lang]["categories"].(map[string]string)
	if cats == nil {
		cats = map[string]string{}
		t[lang]["categories"] = cats
	}
	cats[value] = label
}

// value is what the record field is set to: nil when there are none.
func (t translations) value() interface{} {
	if len(t) == 0 {
		return nil
	}
	return t
}

// ConceptTag is a <concept> after the first one on a var or varGrp: an extra
// search term, optionally in another language (#19). Stored as the `tags`
// JSON field.
type ConceptTag struct {
	Lang string `json:"lang,omitempty"`
	Text string `json:"text"`
}

// conceptsAt splits the <concept> children of m. The first is the concept
// (with its vocab attribute, which marks an external code list); any further
// ones are tags. textAt can't be used here: with several concepts it returns
// the last one.
func conceptsAt(m mxj.Map) (concept, vocab string, tags []ConceptTag) {
	var items []interface{}
	switch v := m["concept"].(type) {
	case nil:
		return "", "", nil
	case []interface{}:
		items = v
	default:
		items = []interface{}{v}
	}
	for i, item := range items {
		text := strings.TrimSpace(extractText(item))
		attrs, _ := item.(map[string]interface{})
		if i == 0 {
			concept = text
			vocab, _ = attrs["-vocab"].(string)
			continue
		}
		if text == "" {
			continue
		}
		lang, _ := attrs["-lang"].(string)
		tags = append(tags, ConceptTag{Lang: lang, Text: text})
	}
	return concept, vocab, tags
}

// inferAnswerType maps a var to an answer type the way formtransform's
// ddiToXlsform reads it back (#44):
//   - numeric: integer with var/@dcml="0", range with a valrng/range that
//     no cdl:constraint explains, else decimal
//   - text: date or time by varFormat/@category, else text
//   - category: grid inside a grid group, else single_choice
//   - multiple: multiple_choice
func inferAnswerType(vM mxj.Map, groupType string) string {
	domain, _ := vM.ValueForPathString("qstn.-responseDomainType")
	switch domain {
	case "numeric":
		if dcml, _ := vM.ValueForPathString("-dcml"); dcml == "0" {
			return "integer"
		}
		if ranges, _ := vM.ValuesForPath("valrng.range"); len(ranges) > 0 && !hasNote(vM, "cdl:constraint") {
			return "range"
		}
		return "decimal"
	case "text":
		switch category, _ := vM.ValueForPathString("varFormat.-category"); category {
		case "date", "time":
			return category
		}
		return "text"
	case "multiple":
		return "multiple_choice"
	case "category":
		if groupType == "grid" {
			return "grid"
		}
		return "single_choice"
	default:
		return ""
	}
}

// hasNote reports whether m has a <notes> child of the given type.
func hasNote(m mxj.Map, noteType string) bool {
	notes, _ := m.ValuesForPath("notes")
	for _, n := range notes {
		if nm, ok := n.(map[string]interface{}); ok && nm["-type"] == noteType {
			return true
		}
	}
	return false
}

// ImportCodebook parses the XML and inserts studies, groups, variables and
// categories into PocketBase. It returns the ID of the new study.
func ImportCodebook(app core.App, mv mxj.Map, rawXML []byte) (string, error) {
	// Base language of a multilingual codebook (formtransform#135): the
	// untagged texts are in it; other languages come as xml:lang siblings.
	baseLang, _ := mv.ValueForPathString("codeBook.-lang")

	// Extract Study info — use textAt() for fields that may carry XML attributes
	title := textAt(mv, "codeBook.stdyDscr.citation.titlStmt.titl")
	idNo := textAt(mv, "codeBook.stdyDscr.citation.titlStmt.IDNo")
	timePeriod := textAt(mv, "codeBook.stdyDscr.stdyInfo.sumDscr.timePrd")   // has event attr
	nation := textAt(mv, "codeBook.stdyDscr.stdyInfo.sumDscr.nation")         // has abbr attr
	universe := textAt(mv, "codeBook.stdyDscr.stdyInfo.sumDscr.universe")     // has clusion attr
	analysisUnit := textAt(mv, "codeBook.stdyDscr.stdyInfo.sumDscr.anlyUnit")
	dataKind := textAt(mv, "codeBook.stdyDscr.stdyInfo.sumDscr.dataKind")

	// Elements with attributes need #text to get just the text content
	author, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.rspStmt.AuthEnty.#text")
	authorAffil, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.rspStmt.AuthEnty.-affiliation")
	producer, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.prodStmt.producer.#text")
	producerAffil, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.prodStmt.producer.-affiliation")
	holdingsURI, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.holdings.-URI")
	holdingsDesc, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.holdings.#text")

	// Extract abstract as raw inner XML to preserve XHTML content
	abstract := ""
	if matches := abstractRe.FindSubmatch(rawXML); len(matches) > 1 {
		abstract = strings.TrimSpace(string(matches[1]))
	}

	// Extract topic classifications (can be multiple)
	var topicClassifications []string
	topics, _ := mv.ValuesForPath("codeBook.stdyDscr.stdyInfo.subject.topcClas")
	for _, t := range topics {
		if s, ok := t.(string); ok {
			topicClassifications = append(topicClassifications, s)
		}
	}

	// Extract keywords (can be multiple)
	var keywords []string
	kws, _ := mv.ValuesForPath("codeBook.stdyDscr.stdyInfo.subject.keyword")
	for _, k := range kws {
		// A keyword with attributes (xml:lang, vocab) arrives as a map; it was
		// dropped before.
		if s := strings.TrimSpace(extractText(k)); s != "" {
			keywords = append(keywords, s)
		}
	}

	studyCollection, err := app.FindCollectionByNameOrId("studies")
	if err != nil {
		return "", err
	}

	studyRecord := core.NewRecord(studyCollection)
	studyRecord.Id = deterministicID(title)
	studyRecord.Set("title", title)
	studyRecord.Set("id_no", idNo)
	studyRecord.Set("abstract", abstract)
	studyRecord.Set("time_period", timePeriod)
	studyRecord.Set("nation", nation)
	studyRecord.Set("universe", universe)
	studyRecord.Set("author", author)
	studyRecord.Set("author_affiliation", authorAffil)
	studyRecord.Set("producer", producer)
	studyRecord.Set("producer_affiliation", producerAffil)
	studyRecord.Set("holdings_uri", holdingsURI)
	studyRecord.Set("holdings_description", holdingsDesc)
	studyRecord.Set("analysis_unit", analysisUnit)
	studyRecord.Set("data_kind", dataKind)
	studyRecord.Set("topic_classifications", topicClassifications)
	studyRecord.Set("keywords", keywords)
	studyRecord.Set("language", baseLang)
	// The codebook as imported: exports read it, not the fields above, so
	// whatever it carries beyond them (skip logic, cdl: notes, sections,
	// question order) comes back out (#37).
	studyRecord.Set("codebook", string(rawXML))

	if err := app.Save(studyRecord); err != nil {
		return "", err
	}

	// Pre-extract qstnLit texts via the XML token stream.
	// mxj cannot represent mixed content (text interleaved with child elements),
	// so tail text after closing tags (e.g. " is more attractive than it was <TIME PERIOD> ago"
	// after the first <xhtml:em>) is silently dropped. The token-based extractor
	// collects all CharData across the full element depth.
	qstnLitTexts, qstnLitOthers := extractVarQstnLits(rawXML)

	// Pre-scan variable groups to build a map of variable DDI ID -> group type
	// This is needed to infer XLSForm question types (e.g. matrix vs select_one)
	varGroupTypeMap := make(map[string]string) // variable DDI ID -> group type
	grpsPre, _ := mv.ValuesForPath("codeBook.dataDscr.varGrp")
	for _, g := range grpsPre {
		gMap, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		gM := mxj.Map(gMap)
		gType, _ := gM.ValueForPathString("-type")
		varIdsAttr, _ := gM.ValueForPathString("-var")
		// A section (formtransform's plain group) says nothing about the
		// answer type; it must not hide the grid its members are also in.
		if varIdsAttr != "" && gType != "" && gType != "section" {
			for _, id := range strings.Fields(varIdsAttr) {
				varGroupTypeMap[id] = gType
			}
		}
	}

	// Map to keep track of variable records by their DDI ID for group assignment
	varRecordsMap := make(map[string]*core.Record)

	// Extract Variables
	vars, err := mv.ValuesForPath("codeBook.dataDscr.var")
	if err != nil {
		log.Println("No variables found in codeBook.dataDscr.var")
	} else {
		varCollection, _ := app.FindCollectionByNameOrId("variables")

		for i, v := range vars {
			vMap, ok := v.(map[string]interface{})
			if !ok {
				continue
			}

			vM := mxj.Map(vMap)
			ddiId, _ := vM.ValueForPathString("-ID")
			vName, _ := vM.ValueForPathString("-name")
			vConcept, vocab, vTags := conceptsAt(vM)
			vQuest := qstnLitTexts[ddiId] // token-based extraction preserves mixed-content text
			vPreQ := textAtLang(vM, "qstn.preQTxt", baseLang)
			vIvInstr := textAtLang(vM, "qstn.ivuInstr", baseLang)
			vHint := textAtLang(vM, "qstn.postQTxt", baseLang)
			vUniverse := textAtLang(vM, "universe", baseLang)
			vTr := translations{}
			for lang, text := range qstnLitOthers[ddiId] {
				vTr.add(lang, "question", text)
			}
			for lang, text := range translationsAt(vM, "qstn.preQTxt", baseLang) {
				vTr.add(lang, "prequestion_text", text)
			}
			for lang, text := range translationsAt(vM, "qstn.ivuInstr", baseLang) {
				vTr.add(lang, "ivu_instructions", text)
			}
			for lang, text := range translationsAt(vM, "qstn.postQTxt", baseLang) {
				vTr.add(lang, "hint", text)
			}
			for lang, text := range translationsAt(vM, "universe", baseLang) {
				vTr.add(lang, "universe", text)
			}
			vIntrvl, _ := vM.ValueForPathString("-intrvl")
			vFmtType, _ := vM.ValueForPathString("varFormat.-type")

			// Build categories as JSON array
			var categories []map[string]interface{}
			cats, _ := vM.ValuesForPath("catgry")
			for _, c := range cats {
				cMap, ok := c.(map[string]interface{})
				if !ok {
					continue
				}
				cM := mxj.Map(cMap)
				val := textAt(cM, "catValu")
				lab := textAtLang(cM, "labl", baseLang)
				missing, _ := cM.ValueForPathString("-missing")
				for lang, text := range translationsAt(cM, "labl", baseLang) {
					vTr.addCategory(lang, strings.TrimSpace(val), text)
				}

				categories = append(categories, map[string]interface{}{
					"value":      strings.TrimSpace(val),
					"label":      strings.TrimSpace(lab),
					"is_missing": missing == "Y",
				})
			}

			// Detect long list (external code list via concept/@vocab)
			hasLongList := vocab != ""

			varRecord := core.NewRecord(varCollection)
			varRecord.Id = deterministicID(title, vName)
			varRecord.Set("study", studyRecord.Id)
			varRecord.Set("ddi_id", ddiId)
			varRecord.Set("name", vName)
			varRecord.Set("concept", vConcept)
			varRecord.Set("tags", vTags)
			varRecord.Set("translations", vTr.value())
			varRecord.Set("question", vQuest)
			varRecord.Set("prequestion_text", vPreQ)
			varRecord.Set("ivu_instructions", vIvInstr)
			varRecord.Set("hint", vHint)
			varRecord.Set("universe", vUniverse)
			varRecord.Set("interval", vIntrvl)
			varRecord.Set("var_format_type", vFmtType)
			varRecord.Set("answer_type", inferAnswerType(vM, varGroupTypeMap[ddiId]))
			varRecord.Set("has_long_list", hasLongList)
			varRecord.Set("long_list_standard", vocab)
			varRecord.Set("categories", categories)
			varRecord.Set("order", i)

			if err := app.Save(varRecord); err != nil {
				log.Printf("Failed to save variable %s: %v", vName, err)
				continue
			}

			if ddiId != "" {
				varRecordsMap[ddiId] = varRecord
			}
		}
	}

	// Second pass: detect has_other by checking for _other companion variables.
	// A variable/group named "foo" has_other=true if a variable named "foo_other" exists.
	varNameSet := make(map[string]bool)
	for _, vr := range varRecordsMap {
		varNameSet[vr.GetString("name")] = true
	}
	for _, vr := range varRecordsMap {
		name := vr.GetString("name")
		if strings.HasSuffix(name, "_other") {
			continue
		}
		if varNameSet[name+"_other"] {
			vr.Set("has_other", true)
			app.Save(vr)
		}
	}

	// Extract Variable Groups
	//
	// type="other" parent groups reference child groups via @varGrp and vars via @var.
	// We flatten the hierarchy: skip child groups and assign all their member vars
	// to the parent group directly. This keeps the DB model flat (no parent_group relation).
	grps, err := mv.ValuesForPath("codeBook.dataDscr.varGrp")
	if err == nil {
		groupCollection, _ := app.FindCollectionByNameOrId("variable_groups")

		// Pre-parse all groups to identify children of type="other" parents.
		type grpInfo struct {
			mxjMap      mxj.Map
			ddiID       string
			name        string
			varIdsAttr  string
			varGrpAttr  string
			grpType     string
			concept     string
			tags        []ConceptTag
			txt         string
			translations translations
		}
		var parsed []grpInfo
		childDDIIDs := make(map[string]bool) // DDI IDs of groups that are children of "other" parents

		for _, g := range grps {
			gMap, ok := g.(map[string]interface{})
			if !ok {
				continue
			}
			gM := mxj.Map(gMap)
			gi := grpInfo{mxjMap: gM}
			gi.ddiID, _ = gM.ValueForPathString("-ID")
			gi.name, _ = gM.ValueForPathString("-name")
			gi.grpType, _ = gM.ValueForPathString("-type")
			gi.varIdsAttr, _ = gM.ValueForPathString("-var")
			gi.varGrpAttr, _ = gM.ValueForPathString("-varGrp")
			gi.concept, _, gi.tags = conceptsAt(gM)
			gi.txt = textAtLang(gM, "txt", baseLang)
			gi.translations = translations{}
			for lang, text := range translationsAt(gM, "txt", baseLang) {
				gi.translations.add(lang, "description", text)
			}
			parsed = append(parsed, gi)
		}

		// Mark child groups referenced by type="other" parents
		for _, gi := range parsed {
			if gi.grpType == "other" && gi.varGrpAttr != "" {
				for _, childID := range strings.Fields(gi.varGrpAttr) {
					childDDIIDs[childID] = true
				}
			}
		}

		// Build a map of child group DDI ID → its @var attribute (member var IDs)
		childVarIDs := make(map[string]string)
		for _, gi := range parsed {
			if childDDIIDs[gi.ddiID] {
				childVarIDs[gi.ddiID] = gi.varIdsAttr
			}
		}

		grpOrder := 0
		for _, gi := range parsed {
			// Skip section groups and child groups absorbed by "other" parents
			if gi.grpType == "section" || childDDIIDs[gi.ddiID] {
				continue
			}

			groupRecord := core.NewRecord(groupCollection)
			groupRecord.Id = deterministicID(title, gi.name)
			groupRecord.Set("study", studyRecord.Id)
			groupRecord.Set("ddi_id", gi.ddiID)
			groupRecord.Set("name", gi.name)
			groupRecord.Set("concept", gi.concept)
			groupRecord.Set("tags", gi.tags)
			groupRecord.Set("translations", gi.translations.value())
			groupRecord.Set("description", gi.txt)
			groupRecord.Set("type", gi.grpType)
			groupRecord.Set("order", grpOrder)
			grpOrder++

			if err := app.Save(groupRecord); err != nil {
				log.Printf("Failed to save group %s: %v", gi.ddiID, err)
				continue
			}

			// Collect all member var IDs: direct @var plus vars from absorbed child groups
			var allVarIDs []string
			if gi.varIdsAttr != "" {
				allVarIDs = append(allVarIDs, strings.Fields(gi.varIdsAttr)...)
			}
			if gi.grpType == "other" && gi.varGrpAttr != "" {
				for _, childID := range strings.Fields(gi.varGrpAttr) {
					if cvids, ok := childVarIDs[childID]; ok && cvids != "" {
						allVarIDs = append(allVarIDs, strings.Fields(cvids)...)
					}
				}
			}

			// Assign group to variables
			for _, id := range allVarIDs {
				if vr, exists := varRecordsMap[id]; exists {
					vr.Set("group", groupRecord.Id)
					app.Save(vr)
				}
			}
		}
	}

	return studyRecord.Id, nil
}

// ImportCodebookData is ImportCodebook for callers that don't need the study ID.
func ImportCodebookData(app core.App, mv mxj.Map, rawXML []byte) error {
	_, err := ImportCodebook(app, mv, rawXML)
	return err
}
