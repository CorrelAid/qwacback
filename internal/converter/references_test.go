package converter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDropOutsideReferences(t *testing.T) {
	in := `{"survey":[` +
		`{"type":"select_one ja_nein","name":"rente","label":"Rente? ${alter}","relevant":"${alter} > 60","hint":"h"},` +
		`{"type":"integer","name":"jahre","label":"Seit?","constraint":". < ${alter} and . > ${rente}","constraint_message":"zu viel","constraint_message::English (en)":"too much","required":"${rente} = '1'"}` +
		`],"choices":[{"list_name":"ja_nein","name":"1","label":"Ja"}],"settings":[],"warnings":[{"code":"ddi-field-missing","message":"m"}]}`
	out, err := DropOutsideReferences([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)

	// Dropped: cells naming alter; kept: the label piping it, references to rows present.
	for _, gone := range []string{`"relevant"`, `"constraint"`, `"constraint_message"`} {
		if strings.Contains(s, gone) {
			t.Errorf("%s kept:\n%s", gone, s)
		}
	}
	for _, kept := range []string{`"label":"Rente? ${alter}"`, `"required":"${rente} = '1'"`, `"hint":"h"`,
		`{"type":"select_one ja_nein","name":"rente","label":"Rente? ${alter}","hint":"h"}`, `"choices":[{"list_name"`} {
		if !strings.Contains(s, kept) {
			t.Errorf("missing %s:\n%s", kept, s)
		}
	}

	var form struct {
		Warnings []Warning `json:"warnings"`
	}
	if err := json.Unmarshal(out, &form); err != nil {
		t.Fatal(err)
	}
	if len(form.Warnings) != 3 || form.Warnings[1].Code != "reference-outside" ||
		!strings.Contains(form.Warnings[1].Message, `relevant of "rente"`) || !strings.Contains(form.Warnings[2].Message, "alter") {
		t.Errorf("warnings: %+v", form.Warnings)
	}
}

func TestDropOutsideReferencesUnchanged(t *testing.T) {
	in := `{"survey":[{"type":"integer","name":"a","label":"A"},{"type":"text","name":"b","relevant":"${a} > 1"}],"choices":[],"settings":[]}`
	out, err := DropOutsideReferences([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != in {
		t.Errorf("changed:\n%s", out)
	}
}

func TestDropOutsideReferencesAddsWarnings(t *testing.T) {
	out, err := DropOutsideReferences([]byte(`{"survey":[{"type":"text","name":"b","relevant":"${a} > 1"}],"choices":[],"settings":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(out), `"warnings":[{"code":"reference-outside","message":"relevant of \"b\" dropped: \"${a} \u003e 1\" refers to a, which isn't in this export"}]}`) {
		t.Errorf("got %s", out)
	}
}
