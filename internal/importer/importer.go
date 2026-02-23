package importer

import (
	"log"
	"strings"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/core"
)

// ImportCodebookData parses the XML and inserts studies, groups, variables and categories into PocketBase.
func ImportCodebookData(app core.App, mv mxj.Map) error {
	// Extract Study info
	title, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.titlStmt.titl")
	idNo, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.titlStmt.IDNo")
	abstract, _ := mv.ValueForPathString("codeBook.stdyDscr.stdyInfo.abstract")
	timePeriod, _ := mv.ValueForPathString("codeBook.stdyDscr.stdyInfo.sumDscr.timePrd")
	nation, _ := mv.ValueForPathString("codeBook.stdyDscr.stdyInfo.sumDscr.nation")
	universe, _ := mv.ValueForPathString("codeBook.stdyDscr.stdyInfo.sumDscr.universe")
	author, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.rspStmt.AuthEnty")
	authorAffil, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.rspStmt.AuthEnty.-affiliation")
	producer, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.prodStmt.producer")
	producerAffil, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.prodStmt.producer.-affiliation")
	holdingsURI, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.holdings.-URI")
	holdingsDesc, _ := mv.ValueForPathString("codeBook.stdyDscr.citation.holdings")
	analysisUnit, _ := mv.ValueForPathString("codeBook.stdyDscr.stdyInfo.sumDscr.anlyUnit")
	dataKind, _ := mv.ValueForPathString("codeBook.stdyDscr.stdyInfo.sumDscr.dataKind")

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

	// Map to keep track of variable records by their DDI ID for group assignment
	varRecordsMap := make(map[string]*core.Record)

	// Extract Variables
	vars, err := mv.ValuesForPath("codeBook.dataDscr.var")
	if err != nil {
		log.Println("No variables found in codeBook.dataDscr.var")
	} else {
		varCollection, _ := app.FindCollectionByNameOrId("variables")

		for _, v := range vars {
			vMap, ok := v.(map[string]interface{})
			if !ok {
				continue
			}

			vM := mxj.Map(vMap)
			ddiId, _ := vM.ValueForPathString("-ID")
			vName, _ := vM.ValueForPathString("-name")
			vLabel, _ := vM.ValueForPathString("labl")
			vQuest, _ := vM.ValueForPathString("qstn.qstnLit")
			vPreQ, _ := vM.ValueForPathString("qstn.preQTxt")
			vIvInstr, _ := vM.ValueForPathString("qstn.ivuInstr")
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
				val, _ := cM.ValueForPathString("catValu")
				lab, _ := cM.ValueForPathString("labl")
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
			varRecord.Set("categories", categories)

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
		for _, g := range grps {
			gMap, ok := g.(map[string]interface{})
			if !ok {
				continue
			}
			gM := mxj.Map(gMap)
			gId, _ := gM.ValueForPathString("-ID")
			gLab, _ := gM.ValueForPathString("labl")
			gTxt, _ := gM.ValueForPathString("txt")
			gType, _ := gM.ValueForPathString("-type")
			varIdsAttr, _ := gM.ValueForPathString("-var") // Space separated IDs

			groupRecord := core.NewRecord(groupCollection)
			groupRecord.Set("study", studyRecord.Id)
			groupRecord.Set("ddi_id", gId)
			groupRecord.Set("label", gLab)
			groupRecord.Set("description", gTxt)
			groupRecord.Set("type", gType)

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
