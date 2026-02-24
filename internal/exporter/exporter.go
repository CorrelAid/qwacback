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
	TopcClas []string `xml:"topcClas"`
}

type SumDscr struct {
	AnlyUnit string `xml:"anlyUnit,omitempty"`
	Universe string `xml:"universe,omitempty"`
	TimePrd  string `xml:"timePrd,omitempty"`
	Nation   string `xml:"nation,omitempty"`
	DataKind string `xml:"dataKind,omitempty"`
}

type DataDscr struct {
	VarGrp []VarGrp `xml:"varGrp"`
	Vars   []Var    `xml:"var"`
}

type VarGrp struct {
	ID   string `xml:"ID,attr"`
	Type string `xml:"type,attr,omitempty"`
	Var  string `xml:"var,attr"` // Space separated variable IDs
	Labl string `xml:"labl,omitempty"`
	Txt  string `xml:"txt,omitempty"`
}

type Var struct {
	ID        string     `xml:"ID,attr"`
	Name      string     `xml:"name,attr"`
	Intrvl    string     `xml:"intrvl,attr,omitempty"`
	Labl      string     `xml:"labl,omitempty"`
	Qstn      *Qstn     `xml:"qstn,omitempty"`
	Catgry    []Category `xml:"catgry,omitempty"`
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

// questionTypeToResponseDomain maps an XLSForm question type back to DDI responseDomainType.
func questionTypeToResponseDomain(questionType string) string {
	switch questionType {
	case "integer":
		return "numeric"
	case "text":
		return "text"
	case "select_one", "matrix":
		return "category"
	case "select_multiple":
		return "multiple"
	default:
		return ""
	}
}

// buildVarFromRecord converts a variable database record into a Var struct.
func buildVarFromRecord(v *core.Record) Var {
	varObj := Var{
		ID:     v.GetString("ddi_id"),
		Name:   v.GetString("name"),
		Intrvl: v.GetString("interval"),
		Labl:   v.GetString("label"),
	}
	if fmtType := v.GetString("var_format_type"); fmtType != "" {
		varObj.VarFormat = &VarFormat{Type: fmtType, Schema: "other"}
	}
	if v.GetString("question") != "" || v.GetString("prequestion_text") != "" || v.GetString("ivu_instructions") != "" {
		varObj.Qstn = &Qstn{
			ResponseDomainType: questionTypeToResponseDomain(v.GetString("question_type")),
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
	for _, cat := range cats {
		catObj := Category{
			CatValu: cat.Value,
			Labl:    cat.Label,
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
	if len(topics) > 0 {
		sd.StdyInfo.Subject = &Subject{TopcClas: topics}
	}

	return sd
}

// ExportVariableToXML generates the DDI <var> XML fragment for a single variable record.
func ExportVariableToXML(v *core.Record) ([]byte, error) {
	varObj := buildVarFromRecord(v)
	return xml.MarshalIndent(varObj, "", "  ")
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
		ID:   g.GetString("ddi_id"),
		Type: g.GetString("type"),
		Var:  strings.Join(groupVars, " "),
		Labl: g.GetString("label"),
		Txt:  g.GetString("description"),
	}
	return xml.MarshalIndent(grp, "", "  ")
}

// ExportStdyDscrToXML generates the DDI <stdyDscr> XML fragment for a study record.
func ExportStdyDscrToXML(study *core.Record) ([]byte, error) {
	sd := buildStdyDscrFromRecord(study)
	return xml.MarshalIndent(sd, "", "  ")
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
			ID:   g.GetString("ddi_id"),
			Type: g.GetString("type"),
			Var:  strings.Join(groupVars, " "),
			Labl: g.GetString("label"),
			Txt:  g.GetString("description"),
		})
	}

	return xml.MarshalIndent(cb, "", "  ")
}
