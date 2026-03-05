package converter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/clbanning/mxj/v2"
)

/*
Answer Type Mapping Between DDI Codebook and XLSForm
========================================================

DDI uses the "responseDomainType" attribute to indicate answer types,
while XLSForm uses the "type" field. This package handles bidirectional
conversion between these formats.

The XLSForm format mirrors the actual spreadsheet structure with three sheets:
- survey: questions and groups (columns: type, name, label, hint, required, appearance, parameters)
- choices: answer options for select questions (columns: list_name, name, label)
- settings: form metadata (columns: form_title, form_id, version)

┌─────────────────────────────────────────────────────────────────────────┐
│ DDI Codebook → XLSForm Mapping                                          │
├─────────────────────┬──────────────┬──────────────────────────────────┤
│ DDI                 │ XLSForm      │ Notes                            │
│ responseDomainType  │ type         │                                  │
├─────────────────────┼──────────────┼──────────────────────────────────┤
│ numeric             │ integer      │ Numeric input (age, count, etc.) │
│ text                │ text         │ Open-ended text response         │
│ category            │ select_one   │ Single choice from options       │
│ category (in grid)  │ matrix       │ Grid/table question              │
│ multiple            │ select_multi │ Multiple choice (checkboxes)     │
└─────────────────────┴──────────────┴──────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ XLSForm → DDI Codebook Mapping                                          │
├──────────────┬─────────────────────┬──────────┬──────────────────────┤
│ XLSForm      │ DDI                 │ DDI      │ DDI                  │
│ type         │ responseDomainType  │ intrvl   │ varFormat.type       │
├──────────────┼─────────────────────┼──────────┼──────────────────────┤
│ integer      │ numeric             │ discrete │ numeric              │
│ decimal      │ numeric             │ discrete │ numeric              │
│ range        │ numeric             │ discrete │ numeric              │
│ text         │ text                │ contin   │ character            │
│ note         │ text                │ contin   │ character            │
│ select_one   │ category            │ discrete │ numeric              │
│ matrix       │ category            │ discrete │ numeric              │
│ select_multi │ multiple            │ discrete │ numeric              │
│ select_one   │ category + vocab    │ discrete │ numeric              │
│  _from_file  │   (no catgry)       │          │                      │
│ select_multi │ multiple + vocab    │ discrete │ numeric              │
│  _from_file  │   (no catgry)       │          │                      │
└──────────────┴─────────────────────┴──────────┴──────────────────────┘

Additional Field Mappings:
- XLSForm "label" ↔ DDI "qstnLit" (main question text)
- XLSForm "hint" ↔ DDI "preQTxt" (pre-question text/hint)
- XLSForm "parameters" (guidance_hint=...) ↔ DDI "ivuInstr" (interviewer instructions)
- XLSForm "choices" sheet ↔ DDI "catgry" elements (answer options)
- XLSForm "name" ↔ DDI "name" attribute (variable identifier)

Group Types:
- XLSForm "begin_group"/"end_group" rows ↔ DDI "<varGrp>" (question groups)
- DDI varGrp type="grid" represents matrix/table questions
- DDI varGrp type="multipleResp" represents checkbox groups
*/

// XLSForm represents the complete XLSForm with its three sheets,
// mirroring the actual XLSForm spreadsheet structure.
type XLSForm struct {
	Survey   []SurveyRow `json:"survey"`
	Choices  []ChoiceRow `json:"choices"`
	Settings SettingsRow `json:"settings"`
}

// SurveyRow represents one row in the "survey" sheet.
// For select questions, the Type field includes the list_name reference,
// e.g. "select_one gender" or "select_multiple hobbies".
type SurveyRow struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Label      string `json:"label,omitempty"`
	Hint       string `json:"hint,omitempty"`
	Required   string `json:"required,omitempty"`
	Relevance  string `json:"relevance,omitempty"`
	Appearance string `json:"appearance,omitempty"`
	Parameters string `json:"parameters,omitempty"`
}

// ChoiceRow represents one row in the "choices" sheet.
type ChoiceRow struct {
	ListName string `json:"list_name"`
	Name     string `json:"name"`
	Label    string `json:"label"`
}

// SettingsRow represents the "settings" sheet (typically a single row).
type SettingsRow struct {
	FormTitle string `json:"form_title,omitempty"`
	FormID    string `json:"form_id,omitempty"`
	Version   string `json:"version,omitempty"`
}

// DDIConcept represents a DDI <concept> element with optional vocabulary attributes.
// When Vocab is set, the variable references an external code list
// (e.g. ISO 3166-1) and inline catgry elements may be omitted.
type DDIConcept struct {
	Vocab string `xml:"vocab,attr,omitempty"`
	Value string `xml:",chardata"`
}

// DDIVar represents a DDI <var> element
type DDIVar struct {
	XMLName   xml.Name      `xml:"var"`
	ID        string        `xml:"ID,attr"`
	Name      string        `xml:"name,attr"`
	Intrvl    string        `xml:"intrvl,attr,omitempty"`
	Qstn      *DDIQstn      `xml:"qstn,omitempty"`
	Catgry    []DDICategory `xml:"catgry,omitempty"`
	Concept   DDIConcept    `xml:"concept"`
	VarFormat *DDIVarFormat `xml:"varFormat,omitempty"`
}

// DDIVarGrp represents a DDI <varGrp> element
type DDIVarGrp struct {
	XMLName xml.Name   `xml:"varGrp"`
	ID      string     `xml:"ID,attr"`
	Name    string     `xml:"name,attr,omitempty"`
	Type    string     `xml:"type,attr,omitempty"`
	Var     string     `xml:"var,attr"` // Space separated variable IDs
	Txt     string     `xml:"txt,omitempty"`
	Concept DDIConcept `xml:"concept"`
}

// DDIQstn represents a DDI <qstn> element
type DDIQstn struct {
	ResponseDomainType string `xml:"responseDomainType,attr,omitempty"`
	PreQTxt            string `xml:"preQTxt,omitempty"`
	QstnLit            string `xml:"qstnLit,omitempty"`
	IvuInstr           string `xml:"ivuInstr,omitempty"`
}

// DDICategory represents a DDI <catgry> element
type DDICategory struct {
	Missing string `xml:"missing,attr,omitempty"`
	CatValu string `xml:"catValu"`
	Labl    string `xml:"labl,omitempty"`
}

// DDIVarFormat represents a DDI <varFormat> element
type DDIVarFormat struct {
	Type   string `xml:"type,attr,omitempty"`
	Schema string `xml:"schema,attr,omitempty"`
}

// DDICodeBook represents a full DDI <codeBook> element (used for parsing study-level exports).
type DDICodeBook struct {
	XMLName  xml.Name    `xml:"codeBook"`
	DataDscr DDIDataDscr `xml:"dataDscr"`
}

// DDIDataDscr is a wrapper element that can hold multiple <varGrp> and <var> elements.
// Used when the output contains both a varGrp and its member variables (e.g. select_multiple).
type DDIDataDscr struct {
	XMLName xml.Name    `xml:"dataDscr"`
	VarGrps []DDIVarGrp `xml:"varGrp"`
	Vars    []DDIVar    `xml:"var"`
}

// DDIToXLSForm converts DDI XML to XLSForm sheet-based JSON format.
//
// This function accepts DDI XML fragments and converts them to the XLSForm
// spreadsheet structure with survey, choices, and settings sheets:
//   - <var> element → survey row(s) + choice rows (if select question)
//   - <varGrp> element → begin_group/end_group survey rows
//
// Returns:
//   - JSON bytes representing the XLSForm struct with 2-space indentation
//   - Error if input is not valid DDI XML or is neither <var> nor <varGrp>
func DDIToXLSForm(ddiXML []byte) ([]byte, error) {
	form := XLSForm{
		Survey:  []SurveyRow{},
		Choices: []ChoiceRow{},
	}

	// Try to parse as a single variable (<var>)
	var v DDIVar
	if err := xml.Unmarshal(ddiXML, &v); err == nil && v.XMLName.Local == "var" {
		convertDDIVarToXLSForm(v, &form)
		return json.MarshalIndent(form, "", "  ")
	}

	// Try to parse as a variable group (<varGrp>)
	var vg DDIVarGrp
	if err := xml.Unmarshal(ddiXML, &vg); err == nil && vg.XMLName.Local == "varGrp" {
		convertDDIVarGrpToXLSForm(vg, &form)
		return json.MarshalIndent(form, "", "  ")
	}

	// Try to parse as a <dataDscr> wrapper (contains varGrp + var elements)
	var dd DDIDataDscr
	if err := xml.Unmarshal(ddiXML, &dd); err == nil && dd.XMLName.Local == "dataDscr" {
		convertDDIDataDscrToXLSForm(dd, &form)
		return json.MarshalIndent(form, "", "  ")
	}

	// Try to parse as a full <codeBook> (extract its dataDscr)
	var cb DDICodeBook
	if err := xml.Unmarshal(ddiXML, &cb); err == nil && cb.XMLName.Local == "codeBook" {
		convertDDIDataDscrToXLSForm(cb.DataDscr, &form)
		return json.MarshalIndent(form, "", "  ")
	}

	return nil, fmt.Errorf("input XML is neither a <var>, <varGrp>, <dataDscr>, nor <codeBook> element")
}

// convertDDIDataDscrToXLSForm converts a <dataDscr> wrapper to XLSForm.
// It detects multipleResp groups and collapses them into select_multiple rows.
func convertDDIDataDscrToXLSForm(dd DDIDataDscr, form *XLSForm) {
	// Build a map of var ID → DDIVar for quick lookup
	varByID := make(map[string]DDIVar, len(dd.Vars))
	for _, v := range dd.Vars {
		varByID[v.ID] = v
	}

	// Track which vars are consumed by multipleResp groups
	consumed := make(map[string]bool)

	for _, grp := range dd.VarGrps {
		switch grp.Type {
		case "multipleResp":
			convertMultipleRespToXLSForm(grp, varByID, form)
			for _, id := range strings.Fields(grp.Var) {
				consumed[id] = true
			}
		case "grid":
			convertGridToXLSForm(grp, varByID, form)
			for _, id := range strings.Fields(grp.Var) {
				consumed[id] = true
			}
		default:
			convertDDIVarGrpToXLSForm(grp, form)
		}
	}

	// Convert remaining (non-consumed) vars
	for _, v := range dd.Vars {
		if !consumed[v.ID] {
			convertDDIVarToXLSForm(v, form)
		}
	}
}

// convertMultipleRespToXLSForm collapses a multipleResp varGrp + its binary member
// vars into a single select_multiple survey row with choices.
func convertMultipleRespToXLSForm(grp DDIVarGrp, varByID map[string]DDIVar, form *XLSForm) {
	memberIDs := strings.Fields(grp.Var)
	if len(memberIDs) == 0 {
		// No member vars — fall back to group conversion
		convertDDIVarGrpToXLSForm(grp, form)
		return
	}

	listName := grp.Name

	row := SurveyRow{
		Type:  "select_multiple " + listName,
		Name:  grp.Name,
		Label: grp.Txt,
	}
	if row.Label == "" {
		row.Label = grp.Concept.Value
	}

	form.Survey = append(form.Survey, row)

	// Each member var's qstnLit becomes a choice label.
	// The choice name is derived from the var name by stripping the group prefix.
	for _, id := range memberIDs {
		v, ok := varByID[id]
		if !ok {
			continue
		}
		choiceName := v.Name
		// Strip group name prefix (e.g. "geraetebesitz_smartphone" → "smartphone")
		if strings.HasPrefix(choiceName, grp.Name+"_") {
			choiceName = strings.TrimPrefix(choiceName, grp.Name+"_")
		}

		choiceLabel := ""
		if v.Qstn != nil {
			choiceLabel = v.Qstn.QstnLit
		}

		form.Choices = append(form.Choices, ChoiceRow{
			ListName: listName,
			Name:     choiceName,
			Label:    choiceLabel,
		})
	}
}

// convertGridToXLSForm converts a grid varGrp + its member vars into a
// begin_group with table-list appearance containing select_one rows with shared choices.
func convertGridToXLSForm(grp DDIVarGrp, varByID map[string]DDIVar, form *XLSForm) {
	memberIDs := strings.Fields(grp.Var)
	if len(memberIDs) == 0 {
		convertDDIVarGrpToXLSForm(grp, form)
		return
	}

	label := grp.Txt
	if label == "" {
		label = grp.Concept.Value
	}

	form.Survey = append(form.Survey, SurveyRow{
		Type:       "begin_group",
		Name:       grp.Name,
		Label:      label,
		Appearance: "table-list",
	})

	// All grid members share the same choice list; use the group name as list_name
	listName := grp.Name
	choicesAdded := false

	for _, id := range memberIDs {
		v, ok := varByID[id]
		if !ok {
			continue
		}

		row := SurveyRow{
			Type: "select_one " + listName,
			Name: v.Name,
		}
		if v.Qstn != nil {
			row.Label = v.Qstn.QstnLit
		}
		form.Survey = append(form.Survey, row)

		// Add choices from the first member var (all grid members share the same categories)
		if !choicesAdded && len(v.Catgry) > 0 {
			for _, cat := range v.Catgry {
				if cat.Missing == "Y" {
					continue
				}
				form.Choices = append(form.Choices, ChoiceRow{
					ListName: listName,
					Name:     cat.CatValu,
					Label:    cat.Labl,
				})
			}
			choicesAdded = true
		}
	}

	form.Survey = append(form.Survey, SurveyRow{
		Type: "end_group",
	})
}

// convertDDIVarToXLSForm converts a single DDI variable to XLSForm survey/choice rows.
//
// Mapping logic:
//   - DDI @name → survey row "name" column
//   - DDI <qstnLit> → survey row "label" column (falls back to <concept>)
//   - DDI <preQTxt> → survey row "hint" column
//   - DDI <ivuInstr> → survey row "parameters" column as "guidance_hint=..."
//   - DDI responseDomainType → survey row "type" column:
//     "numeric" → "integer", "text" → "text",
//     "category" → "select_one <name>", "multiple" → "select_multiple <name>"
//   - DDI <catgry> elements → choice rows with list_name = variable name
func convertDDIVarToXLSForm(v DDIVar, form *XLSForm) {
	row := SurveyRow{
		Name:  v.Name,
		Label: v.Concept.Value, // Fallback label
	}

	if v.Qstn != nil {
		// Use qstnLit as primary label if available
		if v.Qstn.QstnLit != "" {
			row.Label = v.Qstn.QstnLit
		}

		// Map DDI responseDomainType to XLSForm type
		switch v.Qstn.ResponseDomainType {
		case "numeric":
			row.Type = "integer"
		case "text":
			row.Type = "text"
		case "category":
			if v.Concept.Vocab != "" {
				row.Type = "select_one_from_file " + v.Concept.Vocab + ".csv"
			} else {
				row.Type = "select_one " + v.Name
			}
		case "multiple":
			if v.Concept.Vocab != "" {
				row.Type = "select_multiple_from_file " + v.Concept.Vocab + ".csv"
			} else {
				row.Type = "select_multiple " + v.Name
			}
		default:
			row.Type = "text"
		}

		// Map pre-question text to hint
		if v.Qstn.PreQTxt != "" {
			row.Hint = v.Qstn.PreQTxt
		}

		// Map interviewer instructions to parameters
		if v.Qstn.IvuInstr != "" {
			row.Parameters = "guidance_hint=" + v.Qstn.IvuInstr
		}
	}

	form.Survey = append(form.Survey, row)

	// Convert DDI categories to choice rows
	if len(v.Catgry) > 0 {
		for _, cat := range v.Catgry {
			// Skip categories marked as missing values
			if cat.Missing == "Y" {
				continue
			}
			form.Choices = append(form.Choices, ChoiceRow{
				ListName: v.Name,
				Name:     cat.CatValu,
				Label:    cat.Labl,
			})
		}
	}
}

// convertDDIVarGrpToXLSForm converts a DDI variable group to XLSForm begin_group/end_group rows.
func convertDDIVarGrpToXLSForm(vg DDIVarGrp, form *XLSForm) {
	form.Survey = append(form.Survey, SurveyRow{
		Type:  "begin_group",
		Name:  vg.Name,
		Label: vg.Concept.Value,
	})
	form.Survey = append(form.Survey, SurveyRow{
		Type: "end_group",
	})
}

// XLSFormToDDI converts XLSForm sheet-based JSON to DDI XML format.
//
// This function accepts XLSForm JSON with survey, choices, and settings sheets
// and converts each survey row to the corresponding DDI element:
//   - Regular question rows → <var> elements
//   - begin_group/end_group rows → <varGrp> elements
//   - select_one/select_multiple rows look up choices from the choices sheet
//
// For a form with a single question, returns a single <var> or <varGrp> XML element.
// For multiple survey rows, wraps them in a root element.
//
// Returns:
//   - XML bytes with <?xml...?> declaration and 2-space indentation
//   - Error if input is not valid JSON
func XLSFormToDDI(xlsformJSON []byte) ([]byte, error) {
	var form XLSForm
	if err := json.Unmarshal(xlsformJSON, &form); err != nil {
		return nil, fmt.Errorf("failed to parse XLSForm JSON: %w", err)
	}

	if len(form.Survey) == 0 {
		return nil, fmt.Errorf("survey sheet is empty")
	}

	// Build a lookup map: list_name → []ChoiceRow
	choiceMap := make(map[string][]ChoiceRow)
	for _, c := range form.Choices {
		choiceMap[c.ListName] = append(choiceMap[c.ListName], c)
	}

	// Process survey rows
	var vars []DDIVar
	var groups []DDIVarGrp
	var currentGroup *DDIVarGrp
	var groupVarIDs []string

	for _, row := range form.Survey {
		baseType, listName := parseXLSFormType(row.Type)

		switch baseType {
		case "begin_group", "begin_repeat":
			currentGroup = &DDIVarGrp{
				ID:      "VG_" + strings.ReplaceAll(row.Name, " ", "_"),
				Name:    row.Name,
				Concept: DDIConcept{Value: row.Label},
			}
			// Infer DDI group type from appearance, name, or label.
			// XLSForm "table-list" appearance on a group indicates a grid/matrix layout.
			appearanceLower := strings.ToLower(row.Appearance)
			nameLower := strings.ToLower(row.Name)
			labelLower := strings.ToLower(row.Label)
			if appearanceLower == "table-list" || strings.Contains(labelLower, "matrix") || strings.Contains(nameLower, "grid") {
				currentGroup.Type = "grid"
			} else {
				currentGroup.Type = "multipleResp"
			}
			groupVarIDs = nil

		case "end_group", "end_repeat":
			if currentGroup != nil {
				currentGroup.Var = strings.Join(groupVarIDs, " ")
				groups = append(groups, *currentGroup)
				currentGroup = nil
				groupVarIDs = nil
			}

		case "select_multiple":
			grp, binaryVars := convertSelectMultipleToDDI(row, listName, choiceMap)
			groups = append(groups, grp)
			vars = append(vars, binaryVars...)

		default:
			v := convertSurveyRowToDDIVar(row, baseType, listName, choiceMap)
			vars = append(vars, v)
			if currentGroup != nil {
				groupVarIDs = append(groupVarIDs, v.ID)
			}
		}
	}

	// If there's a single var and no groups, return just the var
	if len(vars) == 1 && len(groups) == 0 {
		output, err := xml.MarshalIndent(vars[0], "", "  ")
		if err != nil {
			return nil, err
		}
		return append([]byte(xml.Header), output...), nil
	}

	// If there's a single group and no standalone vars, return just the group
	if len(groups) == 1 && len(vars) == 0 {
		output, err := xml.MarshalIndent(groups[0], "", "  ")
		if err != nil {
			return nil, err
		}
		return append([]byte(xml.Header), output...), nil
	}

	// Multiple elements (e.g. varGrp + member vars): wrap in <dataDscr>
	if len(groups) > 0 || len(vars) > 1 {
		wrapper := DDIDataDscr{
			VarGrps: groups,
			Vars:    vars,
		}
		output, err := xml.MarshalIndent(wrapper, "", "  ")
		if err != nil {
			return nil, err
		}
		return append([]byte(xml.Header), output...), nil
	}

	// Single var fallback
	if len(vars) > 0 {
		output, err := xml.MarshalIndent(vars[0], "", "  ")
		if err != nil {
			return nil, err
		}
		return append([]byte(xml.Header), output...), nil
	}

	return nil, fmt.Errorf("no convertible survey rows found")
}

// parseXLSFormType splits an XLSForm type like "select_one gender" into
// the base type ("select_one") and the list_name ("gender").
func parseXLSFormType(t string) (baseType, listName string) {
	parts := strings.SplitN(t, " ", 2)
	baseType = parts[0]
	if len(parts) > 1 {
		listName = parts[1]
	}
	return
}

// convertSelectMultipleToDDI converts a select_multiple survey row into a
// DDI <varGrp type="multipleResp"> plus one binary <var> per choice option.
// Each binary var has categories 0 and 1 (no labels).
func convertSelectMultipleToDDI(row SurveyRow, listName string, choiceMap map[string][]ChoiceRow) (DDIVarGrp, []DDIVar) {
	groupName := row.Name
	groupID := "VG_" + strings.ReplaceAll(groupName, " ", "_")

	choices := choiceMap[listName]
	var varIDs []string
	var binaryVars []DDIVar

	for _, c := range choices {
		varName := groupName + "_" + c.Name
		varID := "V_" + varName
		varIDs = append(varIDs, varID)

		v := DDIVar{
			ID:     varID,
			Name:   varName,
			Intrvl: "discrete",
			Qstn: &DDIQstn{
				ResponseDomainType: "multiple",
				PreQTxt:            row.Label,
				QstnLit:            c.Label,
			},
			Catgry: []DDICategory{
				{CatValu: "0"},
				{CatValu: "1"},
			},
			Concept:   DDIConcept{Value: row.Label + ": " + c.Label},
			VarFormat: &DDIVarFormat{Type: "numeric", Schema: "other"},
		}

		binaryVars = append(binaryVars, v)
	}

	grp := DDIVarGrp{
		ID:      groupID,
		Name:    groupName,
		Type:    "multipleResp",
		Var:     strings.Join(varIDs, " "),
		Txt:     row.Label,
		Concept: DDIConcept{Value: row.Label},
	}

	return grp, binaryVars
}

// convertSurveyRowToDDIVar converts a single survey row to a DDI <var> element.
func convertSurveyRowToDDIVar(row SurveyRow, baseType, listName string, choiceMap map[string][]ChoiceRow) DDIVar {
	v := DDIVar{
		Name:    row.Name,
		ID:      "V_" + strings.ReplaceAll(row.Name, " ", "_"),
		Concept: DDIConcept{Value: row.Label},
	}

	var responseDomainType string
	var interval string
	var varFormatType string

	switch baseType {
	case "integer", "decimal", "range":
		responseDomainType = "numeric"
		interval = "discrete"
		varFormatType = "numeric"
	case "text", "note":
		responseDomainType = "text"
		interval = "contin"
		varFormatType = "character"
	case "select_one", "matrix":
		responseDomainType = "category"
		interval = "discrete"
		varFormatType = "numeric"
	case "select_one_from_file":
		responseDomainType = "category"
		interval = "discrete"
		varFormatType = "numeric"
		// listName is the CSV filename; strip .csv to get the standard code
		v.Concept.Vocab = strings.TrimSuffix(listName, ".csv")
	case "select_multiple_from_file":
		responseDomainType = "multiple"
		interval = "discrete"
		varFormatType = "numeric"
		v.Concept.Vocab = strings.TrimSuffix(listName, ".csv")
	default:
		responseDomainType = "text"
		interval = "contin"
		varFormatType = "character"
	}

	v.Intrvl = interval
	v.VarFormat = &DDIVarFormat{
		Type:   varFormatType,
		Schema: "other",
	}

	v.Qstn = &DDIQstn{
		ResponseDomainType: responseDomainType,
		QstnLit:            row.Label,
	}

	if row.Hint != "" {
		v.Qstn.PreQTxt = row.Hint
	}

	// Parse parameters string (e.g. "guidance_hint=some text")
	if row.Parameters != "" {
		for _, param := range strings.Split(row.Parameters, ";") {
			param = strings.TrimSpace(param)
			if kv := strings.SplitN(param, "=", 2); len(kv) == 2 {
				if strings.TrimSpace(kv[0]) == "guidance_hint" {
					v.Qstn.IvuInstr = strings.TrimSpace(kv[1])
				}
			}
		}
	}

	// Look up choices from the choices sheet
	if listName != "" {
		if choices, ok := choiceMap[listName]; ok {
			v.Catgry = make([]DDICategory, 0, len(choices))
			for _, c := range choices {
				v.Catgry = append(v.Catgry, DDICategory{
					CatValu: c.Name,
					Labl:    c.Label,
				})
			}
		}
	}

	return v
}

// ParseDDICodebookFragment parses a DDI XML fragment that may contain multiple var or varGrp elements
// and returns the result in XLSForm sheet-based format.
func ParseDDICodebookFragment(ddiXML []byte) ([]byte, error) {
	mv, err := mxj.NewMapXml(ddiXML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	form := XLSForm{
		Survey:  []SurveyRow{},
		Choices: []ChoiceRow{},
	}

	// Extract variables
	vars, err := mv.ValuesForPath("var")
	if err == nil {
		for _, v := range vars {
			vMap, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			vXML, err := mxj.Map(vMap).Xml()
			if err != nil {
				continue
			}
			var ddiVar DDIVar
			if err := xml.Unmarshal(vXML, &ddiVar); err != nil {
				continue
			}
			convertDDIVarToXLSForm(ddiVar, &form)
		}
	}

	// Extract variable groups
	grps, err := mv.ValuesForPath("varGrp")
	if err == nil {
		for _, g := range grps {
			gMap, ok := g.(map[string]interface{})
			if !ok {
				continue
			}
			gXML, err := mxj.Map(gMap).Xml()
			if err != nil {
				continue
			}
			var ddiGrp DDIVarGrp
			if err := xml.Unmarshal(gXML, &ddiGrp); err != nil {
				continue
			}
			convertDDIVarGrpToXLSForm(ddiGrp, &form)
		}
	}

	return json.MarshalIndent(form, "", "  ")
}
