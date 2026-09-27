package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// logicColumns are the survey columns whose ${name} references make a form
// invalid when the named question is missing. Labels and hints may pipe
// ${name} too, but those only print the value, so they stay.
var logicColumns = map[string]bool{
	"relevant":      true,
	"constraint":    true,
	"required":      true,
	"calculation":   true,
	"choice_filter": true,
	"default":       true,
	"read_only":     true,
	"readonly":      true,
	"repeat_count":  true,
	"trigger":       true,
}

var referenceRe = regexp.MustCompile(`\$\{([^}\s]+)\}`)

// Warning is one entry of the `warnings` list in XLSForm JSON.
type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// field is one key/value of a JSON object, in document order.
type field struct {
	key   string
	value json.RawMessage
}

// DropOutsideReferences removes logic cells (relevant, constraint, …) that
// refer to a question not in the form, and appends a `reference-outside`
// warning for each (#46). A single question exported from a study keeps its
// skip logic, e.g. relevant "${alter} > 60" without an `alter` row, which
// pyxform and Kobo reject. The condition stays in the DDI (cdl:relevant,
// <universe>). A constraint_message goes with its constraint.
//
// Rows and keys keep their order; rows without such a cell are unchanged.
func DropOutsideReferences(xlsformJSON []byte) ([]byte, error) {
	top, err := objectFields(xlsformJSON)
	if err != nil {
		return nil, err
	}
	var survey []json.RawMessage
	var warnings []json.RawMessage
	for _, f := range top {
		switch f.key {
		case "survey":
			if err := json.Unmarshal(f.value, &survey); err != nil {
				return nil, fmt.Errorf("survey: %w", err)
			}
		case "warnings":
			if err := json.Unmarshal(f.value, &warnings); err != nil {
				return nil, fmt.Errorf("warnings: %w", err)
			}
		}
	}

	names := map[string]bool{}
	rows := make([][]field, len(survey))
	for i, raw := range survey {
		if rows[i], err = objectFields(raw); err != nil {
			return nil, fmt.Errorf("survey row %d: %w", i, err)
		}
		for _, f := range rows[i] {
			if f.key == "name" {
				var n string
				_ = json.Unmarshal(f.value, &n)
				names[n] = true
			}
		}
	}

	changed := false
	for i, row := range rows {
		var name string
		dropped := map[string]bool{}
		for _, f := range row {
			if f.key == "name" {
				_ = json.Unmarshal(f.value, &name)
			}
		}
		for _, f := range row {
			if !logicColumns[baseColumn(f.key)] {
				continue
			}
			var cell string
			if json.Unmarshal(f.value, &cell) != nil {
				continue
			}
			var missing []string
			for _, m := range referenceRe.FindAllStringSubmatch(cell, -1) {
				if !names[m[1]] {
					missing = append(missing, m[1])
				}
			}
			if len(missing) == 0 {
				continue
			}
			dropped[f.key] = true
			w, _ := json.Marshal(Warning{
				Code: "reference-outside",
				Message: fmt.Sprintf("%s of %q dropped: %q refers to %s, which isn't in this export",
					f.key, name, cell, strings.Join(missing, ", ")),
			})
			warnings = append(warnings, w)
		}
		if len(dropped) == 0 {
			continue
		}
		changed = true
		kept := row[:0:0]
		for _, f := range row {
			if dropped[f.key] || (dropped["constraint"] && baseColumn(f.key) == "constraint_message") {
				continue
			}
			kept = append(kept, f)
		}
		rows[i] = kept
	}
	if !changed {
		return xlsformJSON, nil
	}

	for i, row := range rows {
		survey[i] = encodeObject(row)
	}
	surveyJSON, _ := json.Marshal(survey)
	warningsJSON, _ := json.Marshal(warnings)
	hasWarnings := false
	for i, f := range top {
		switch f.key {
		case "survey":
			top[i].value = surveyJSON
		case "warnings":
			top[i].value = warningsJSON
			hasWarnings = true
		}
	}
	if !hasWarnings {
		top = append(top, field{"warnings", warningsJSON})
	}
	return encodeObject(top), nil
}

// baseColumn strips a language suffix: "constraint_message::English (en)".
func baseColumn(key string) string {
	if i := strings.Index(key, "::"); i >= 0 {
		return key[:i]
	}
	return key
}

// objectFields decodes a JSON object into its fields, in order.
func objectFields(raw []byte) ([]field, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("not a JSON object")
	}
	var out []field
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := tok.(string)
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return nil, err
		}
		out = append(out, field{key, v})
	}
	return out, nil
}

func encodeObject(fields []field) []byte {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, f := range fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		k, _ := json.Marshal(f.key)
		buf.Write(k)
		buf.WriteByte(':')
		buf.Write(f.value)
	}
	buf.WriteByte('}')
	return buf.Bytes()
}
