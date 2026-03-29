package exporter

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type CodeBook struct {
	XMLName  xml.Name `xml:"codeBook"`
	Xmlns    string   `xml:"xmlns,attr"`
	Xsi      string   `xml:"xmlns:xsi,attr"`
	StdyDscr StdyDscr `xml:"stdyDscr"`
	DataDscr DataDscr `xml:"dataDscr"`
}

type StdyDscr struct {
	Citation Citation `xml:"citation"`
	StdyInfo StdyInfo `xml:"stdyInfo"`
}

type Citation struct {
	TitlStmt TitlStmt  `xml:"titlStmt"`
	RspStmt  *RspStmt  `xml:"rspStmt,omitempty"`
	ProdStmt *ProdStmt `xml:"prodStmt,omitempty"`
	Holdings *Holdings `xml:"holdings,omitempty"`
}

type TitlStmt struct {
	Titl string `xml:"titl"`
	IDNo string `xml:"IDNo,omitempty"`
}

type RspStmt struct {
	AuthEnty AuthEnty `xml:"AuthEnty"`
}

type AuthEnty struct {
	Affiliation string `xml:"affiliation,attr,omitempty"`
	Value       string `xml:",chardata"`
}

type ProdStmt struct {
	Producer Producer `xml:"producer"`
}

type Producer struct {
	Affiliation string `xml:"affiliation,attr,omitempty"`
	Value       string `xml:",chardata"`
}

type Holdings struct {
	URI   string `xml:"URI,attr,omitempty"`
	Value string `xml:",chardata"`
}

type StdyInfo struct {
	Subject  *Subject `xml:"subject,omitempty"`
	Abstract Abstract `xml:"abstract"`
	SumDscr  SumDscr  `xml:"sumDscr"`
}

type Abstract struct {
	Content string `xml:",chardata"`
}

type Subject struct {
	Keywords []string `xml:"keyword"`
	TopcClas []string `xml:"topcClas"`
}

type SumDscr struct {
	TimePrd  string `xml:"timePrd,omitempty"`
	Nation   string `xml:"nation,omitempty"`
	AnlyUnit string `xml:"anlyUnit,omitempty"`
	Universe string `xml:"universe,omitempty"`
	DataKind string `xml:"dataKind,omitempty"`
}

type DataDscr struct {
	XMLName xml.Name `xml:"dataDscr"`
	VarGrp  []VarGrp `xml:"varGrp"`
	Vars    []Var    `xml:"var"`
}

type Concept struct {
	Vocab string `xml:"vocab,attr,omitempty"`
	Value string `xml:",chardata"`
}

type VarGrp struct {
	XMLName   xml.Name `xml:"varGrp"`
	ID        string   `xml:"ID,attr"`
	Name      string   `xml:"name,attr,omitempty"`
	Type      string   `xml:"type,attr,omitempty"`
	Var       string   `xml:"var,attr,omitempty"`       // Space-separated variable IDs
	VarGrpRef string   `xml:"varGrp,attr,omitempty"`    // Space-separated child varGrp IDs
	Txt       string   `xml:"txt,omitempty"`
	Concept   Concept  `xml:"concept"`
}

type Var struct {
	XMLName   xml.Name   `xml:"var"`
	ID        string     `xml:"ID,attr"`
	Name      string     `xml:"name,attr"`
	Intrvl    string     `xml:"intrvl,attr,omitempty"`
	Qstn      *Qstn      `xml:"qstn,omitempty"`
	Catgry    []Category `xml:"catgry,omitempty"`
	Concept   Concept    `xml:"concept"`
	VarFormat *VarFormat `xml:"varFormat,omitempty"`
}

type VarFormat struct {
	Type   string `xml:"type,attr,omitempty"`
	Schema string `xml:"schema,attr,omitempty"`
}

type Qstn struct {
	ResponseDomainType string `xml:"responseDomainType,attr,omitempty"`
	PreQTxt            string `xml:"preQTxt,omitempty"`
	QstnLit            string `xml:"qstnLit,omitempty"`
	IvuInstr           string `xml:"ivuInstr,omitempty"`
}

type Category struct {
	Missing string `xml:"missing,attr,omitempty"`
	CatValu string `xml:"catValu"`
	Labl    string `xml:"labl,omitempty"`
}

// answerTypeToResponseDomain maps an answer type back to DDI responseDomainType.
func answerTypeToResponseDomain(answerType string) string {
	switch answerType {
	case "integer":
		return "numeric"
	case "text":
		return "text"
	case "single_choice", "grid":
		return "category"
	case "multiple_choice":
		return "multiple"
	default:
		return ""
	}
}

// buildVarFromRecord converts a variable database record into a Var struct.
func buildVarFromRecord(v *core.Record) Var {
	varObj := Var{
		ID:      v.GetString("ddi_id"),
		Name:    v.GetString("name"),
		Intrvl:  v.GetString("interval"),
		Concept: Concept{Value: v.GetString("concept")},
	}
	if std := v.GetString("long_list_standard"); std != "" {
		varObj.Concept.Vocab = std
	}
	if fmtType := v.GetString("var_format_type"); fmtType != "" {
		varObj.VarFormat = &VarFormat{Type: fmtType, Schema: "other"}
	}
	if v.GetString("question") != "" || v.GetString("prequestion_text") != "" || v.GetString("ivu_instructions") != "" {
		varObj.Qstn = &Qstn{
			ResponseDomainType: answerTypeToResponseDomain(v.GetString("answer_type")),
			PreQTxt:            v.GetString("prequestion_text"),
			QstnLit:            v.GetString("question"),
			IvuInstr:           v.GetString("ivu_instructions"),
		}
	}

	var cats []struct {
		Value     string `json:"value"`
		Label     string `json:"label"`
		IsMissing bool   `json:"is_missing"`
	}
	if raw := v.GetString("categories"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &cats); err != nil {
			log.Printf("WARNING: failed to unmarshal categories for variable %s: %v", v.Id, err)
		}
	}
	isMultiple := v.GetString("answer_type") == "multiple_choice"
	for _, cat := range cats {
		catObj := Category{
			CatValu: cat.Value,
		}
		if !isMultiple {
			catObj.Labl = cat.Label
		}
		if cat.IsMissing {
			catObj.Missing = "Y"
		}
		varObj.Catgry = append(varObj.Catgry, catObj)
	}

	return varObj
}

// buildStdyDscrFromRecord converts a study database record into a StdyDscr struct.
func buildStdyDscrFromRecord(study *core.Record) StdyDscr {
	sd := StdyDscr{
		Citation: Citation{
			TitlStmt: TitlStmt{
				Titl: study.GetString("title"),
				IDNo: study.GetString("id_no"),
			},
		},
		StdyInfo: StdyInfo{
			Abstract: Abstract{Content: study.GetString("abstract")},
			SumDscr: SumDscr{
				AnlyUnit: study.GetString("analysis_unit"),
				Universe: study.GetString("universe"),
				TimePrd:  study.GetString("time_period"),
				Nation:   study.GetString("nation"),
				DataKind: study.GetString("data_kind"),
			},
		},
	}

	if author := study.GetString("author"); author != "" {
		sd.Citation.RspStmt = &RspStmt{
			AuthEnty: AuthEnty{
				Value:       author,
				Affiliation: study.GetString("author_affiliation"),
			},
		}
	}
	if prod := study.GetString("producer"); prod != "" {
		sd.Citation.ProdStmt = &ProdStmt{
			Producer: Producer{
				Value:       prod,
				Affiliation: study.GetString("producer_affiliation"),
			},
		}
	}
	if uri := study.GetString("holdings_uri"); uri != "" {
		sd.Citation.Holdings = &Holdings{
			URI:   uri,
			Value: study.GetString("holdings_description"),
		}
	}

	var topics []string
	if raw := study.GetString("topic_classifications"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &topics); err != nil {
			log.Printf("WARNING: failed to unmarshal topic_classifications for study %s: %v", study.Id, err)
		}
	}
	var keywords []string
	if raw := study.GetString("keywords"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &keywords); err != nil {
			log.Printf("WARNING: failed to unmarshal keywords for study %s: %v", study.Id, err)
		}
	}
	if len(keywords) > 0 || len(topics) > 0 {
		sd.StdyInfo.Subject = &Subject{Keywords: keywords, TopcClas: topics}
	}

	return sd
}

// ExportVariableToXML generates the DDI <var> XML fragment for a single variable record.
func ExportVariableToXML(v *core.Record) ([]byte, error) {
	varObj := buildVarFromRecord(v)
	return xml.MarshalIndent(varObj, "", "  ")
}

// ExportVariableWithGroupToXML generates a DDI fragment for a single variable.
// If the variable belongs to a grid group, it returns a <dataDscr> containing the
// <varGrp> and only this single <var> (no sibling variables).
// If the variable has no group, it returns a plain <var> element.
func ExportVariableWithGroupToXML(app core.App, v *core.Record) ([]byte, error) {
	groupID := v.GetString("group")
	if groupID == "" {
		return ExportVariableToXML(v)
	}

	groupRecord, err := app.FindRecordById("variable_groups", groupID)
	if err != nil {
		// Group not found — fall back to plain var export
		return ExportVariableToXML(v)
	}

	if groupRecord.GetString("type") != "grid" {
		return ExportVariableToXML(v)
	}

	varObj := buildVarFromRecord(v)

	grp := VarGrp{
		ID:      groupRecord.GetString("ddi_id"),
		Name:    groupRecord.GetString("name"),
		Type:    groupRecord.GetString("type"),
		Var:     v.GetString("ddi_id"),
		Concept: Concept{Value: groupRecord.GetString("concept")},
		Txt:     groupRecord.GetString("description"),
	}

	dd := DataDscr{
		VarGrp: []VarGrp{grp},
		Vars:   []Var{varObj},
	}
	return xml.MarshalIndent(dd, "", "  ")
}

// ExportVarGrpToXML generates the DDI <varGrp> XML fragment for a single variable group record.
func ExportVarGrpToXML(app core.App, g *core.Record) ([]byte, error) {
	varRecords, err := app.FindRecordsByFilter(
		"variables",
		"group = {:id}",
		"order", 0, 0,
		dbx.Params{"id": g.Id},
	)
	if err != nil {
		return nil, err
	}

	var groupVars []string
	for _, v := range varRecords {
		groupVars = append(groupVars, v.GetString("ddi_id"))
	}

	grp := VarGrp{
		ID:      g.GetString("ddi_id"),
		Name:    g.GetString("name"),
		Type:    g.GetString("type"),
		Var:     strings.Join(groupVars, " "),
		Concept: Concept{Value: g.GetString("concept")},
		Txt:     g.GetString("description"),
	}
	return xml.MarshalIndent(grp, "", "  ")
}

// ExportVarGrpCodebookToXML generates a DDI <dataDscr> fragment containing the
// <varGrp> and all its member <var> elements.
func ExportVarGrpCodebookToXML(app core.App, g *core.Record) ([]byte, error) {
	varRecords, err := app.FindRecordsByFilter(
		"variables",
		"group = {:id}",
		"order", 0, 0,
		dbx.Params{"id": g.Id},
	)
	if err != nil {
		return nil, err
	}

	var groupVars []string
	var vars []Var
	for _, v := range varRecords {
		vars = append(vars, buildVarFromRecord(v))
		groupVars = append(groupVars, v.GetString("ddi_id"))
	}

	grp := VarGrp{
		ID:      g.GetString("ddi_id"),
		Name:    g.GetString("name"),
		Type:    g.GetString("type"),
		Var:     strings.Join(groupVars, " "),
		Concept: Concept{Value: g.GetString("concept")},
		Txt:     g.GetString("description"),
	}

	dd := DataDscr{
		VarGrp: []VarGrp{grp},
		Vars:   vars,
	}
	return xml.MarshalIndent(dd, "", "  ")
}

// ExportStudyToXML converts a study and its variables into a DDI-XML byte slice.
func ExportStudyToXML(app core.App, study *core.Record) ([]byte, error) {
	// Fetch groups
	groupRecords, err := app.FindRecordsByFilter(
		"variable_groups",
		"study = {:id}",
		"order", 0, 0,
		dbx.Params{"id": study.Id},
	)
	if err != nil {
		return nil, err
	}

	// Fetch variables
	varRecords, err := app.FindRecordsByFilter(
		"variables",
		"study = {:id}",
		"order", 0, 0,
		dbx.Params{"id": study.Id},
	)
	if err != nil {
		return nil, err
	}

	cb := CodeBook{
		Xmlns:    "ddi:codebook:2_5",
		Xsi:      "http://www.w3.org/2001/XMLSchema-instance",
		StdyDscr: buildStdyDscrFromRecord(study),
	}

	cb.DataDscr.Vars = make([]Var, 0, len(varRecords))
	for _, v := range varRecords {
		cb.DataDscr.Vars = append(cb.DataDscr.Vars, buildVarFromRecord(v))
	}

	cb.DataDscr.VarGrp = make([]VarGrp, 0, len(groupRecords))
	for _, g := range groupRecords {
		var groupVars []string
		for _, v := range varRecords {
			if v.GetString("group") == g.Id {
				groupVars = append(groupVars, v.GetString("ddi_id"))
			}
		}

		cb.DataDscr.VarGrp = append(cb.DataDscr.VarGrp, VarGrp{
			ID:      g.GetString("ddi_id"),
			Name:    g.GetString("name"),
			Type:    g.GetString("type"),
			Var:     strings.Join(groupVars, " "),
			Concept: Concept{Value: g.GetString("concept")},
			Txt:     g.GetString("description"),
		})
	}

	return xml.MarshalIndent(cb, "", "  ")
}
