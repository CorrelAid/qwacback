package importer

import (
	"log"
	"regexp"
	"strings"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
)

var abstractRe = regexp.MustCompile(`(?s)<abstract[^>]*>(.*?)</abstract>`)

// textAt extracts the text content of an mxj path, handling elements that
// have XML attributes. mxj represents <foo bar="x">text</foo> as a map
// {"-bar":"x","#text":"text"}, so we try the #text sub-key first and fall
// back to the plain path for elements without attributes.
func textAt(mv mxj.Map, path string) string {
	if v, err := mv.ValueForPathString(path + ".#text"); err == nil && v != "" {
		return v
	}
	v, _ := mv.ValueForPathString(path)
	return v
}

// inferQuestionType maps DDI responseDomainType + group type to an XLSForm question type.
func inferQuestionType(responseDomainType, groupType string) string {
	switch responseDomainType {
	case "numeric":
		return "integer"
	case "text":
		return "text"
	case "multiple":
		return "select_multiple"
	case "category":
		if groupType == "grid" {
			return "matrix"
		}
		return "select_one"
	default:
		return ""
	}
}

// ImportCodebookData parses the XML and inserts studies, groups, variables and categories into PocketBase.
func ImportCodebookData(app core.App, mv mxj.Map, rawXML []byte) error {
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

	studyCollection, err := app.FindCollectionByNameOrId("studies")
	if err != nil {
		return err
	}

	studyRecord := core.NewRecord(studyCollection)
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

	if err := app.Save(studyRecord); err != nil {
		return err
	}

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
		if varIdsAttr != "" && gType != "" {
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
			vLabel := textAt(vM, "labl")         // labl may have xml:lang attr
			vQuest := textAt(vM, "qstn.qstnLit") // qstnLit may have xml:lang attr
			vPreQ := textAt(vM, "qstn.preQTxt")
			vIvInstr := textAt(vM, "qstn.ivuInstr")
			vQstnType, _ := vM.ValueForPathString("qstn.-responseDomainType")
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
				lab := textAt(cM, "labl")
				missing, _ := cM.ValueForPathString("-missing")

				categories = append(categories, map[string]interface{}{
					"value":      strings.TrimSpace(val),
					"label":      strings.TrimSpace(lab),
					"is_missing": missing == "Y",
				})
			}

			varRecord := core.NewRecord(varCollection)
			varRecord.Set("study", studyRecord.Id)
			varRecord.Set("ddi_id", ddiId)
			varRecord.Set("name", vName)
			varRecord.Set("label", vLabel)
			varRecord.Set("question", vQuest)
			varRecord.Set("prequestion_text", vPreQ)
			varRecord.Set("ivu_instructions", vIvInstr)
			varRecord.Set("interval", vIntrvl)
			varRecord.Set("var_format_type", vFmtType)
			varRecord.Set("question_type", inferQuestionType(vQstnType, varGroupTypeMap[ddiId]))
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

	// Extract Variable Groups
	grps, err := mv.ValuesForPath("codeBook.dataDscr.varGrp")
	if err == nil {
		groupCollection, _ := app.FindCollectionByNameOrId("variable_groups")
		grpOrder := 0
		for _, g := range grps {
			gMap, ok := g.(map[string]interface{})
			if !ok {
				continue
			}
			gM := mxj.Map(gMap)
			gId, _ := gM.ValueForPathString("-ID")
			gLab := textAt(gM, "labl")
			gTxt := textAt(gM, "txt")
			gType, _ := gM.ValueForPathString("-type")
			varIdsAttr, _ := gM.ValueForPathString("-var") // Space separated IDs

			// Skip section groups – they are structural containers, not semantic variable groups.
			if gType == "section" {
				continue
			}

			groupRecord := core.NewRecord(groupCollection)
			groupRecord.Set("study", studyRecord.Id)
			groupRecord.Set("ddi_id", gId)
			groupRecord.Set("label", gLab)
			groupRecord.Set("description", gTxt)
			groupRecord.Set("type", gType)
			groupRecord.Set("order", grpOrder)
			grpOrder++

			if err := app.Save(groupRecord); err != nil {
				log.Printf("Failed to save group %s: %v", gId, err)
				continue
			}

			// Assign group to variables
			if varIdsAttr != "" {
				ids := strings.Fields(varIdsAttr)
				for _, id := range ids {
					if vr, exists := varRecordsMap[id]; exists {
						vr.Set("group", groupRecord.Id)
						app.Save(vr)
					}
				}
			}
		}
	}

	return nil
}
