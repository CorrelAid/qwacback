// Package exporter writes DDI for a study or a question from the codebook as
// imported (studies.codebook), not from the records: a CDL codebook carries
// the whole form, much of it in typed <notes> the records don't model (skip
// logic, sections, question order, cdl: notes; formtransform v0.7.0). Read
// from the codebook, all of it comes back out, and formtransform's
// ddiToXlsform can rebuild the form (#37).
package exporter

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"qwacback/internal/ddixml"
)

// ErrNoCodebook means the study has no stored codebook, e.g. a record
// created by hand instead of by an import.
var ErrNoCodebook = errors.New("study has no stored codebook")

// ErrNotInCodebook means the record's DDI element isn't in its study's codebook.
var ErrNotInCodebook = errors.New("element not in the study's codebook")

// StudyXML returns the study's codebook as imported, with an XML declaration.
func StudyXML(study *core.Record) ([]byte, error) {
	cb := study.GetString("codebook")
	if strings.TrimSpace(cb) == "" {
		return nil, ErrNoCodebook
	}
	if strings.HasPrefix(strings.TrimSpace(cb), "<?xml") {
		return []byte(cb), nil
	}
	return []byte(xml.Header + cb), nil
}

// codebook is the parsed <dataDscr> of a study's codebook.
type codebook struct {
	lang     string
	children []*ddixml.Element
	byID     map[string]*ddixml.Element
}

func loadCodebook(app core.App, studyID string) (*codebook, error) {
	study, err := app.FindRecordById("studies", studyID)
	if err != nil {
		return nil, err
	}
	raw, err := StudyXML(study)
	if err != nil {
		return nil, err
	}
	lang, children, err := ddixml.DataDscr(raw)
	if err != nil {
		return nil, fmt.Errorf("study %s: %w", studyID, err)
	}
	cb := &codebook{lang: lang, children: children, byID: map[string]*ddixml.Element{}}
	for _, c := range children {
		if c.ID != "" {
			cb.byID[c.ID] = c
		}
	}
	return cb, nil
}

// GroupXML returns a <dataDscr> with the group's <varGrp>, the groups nested
// in it (@varGrp, e.g. the choices of a type="other" group) and every <var>
// they refer to (@var), in codebook order.
func GroupXML(app core.App, g *core.Record) ([]byte, error) {
	cb, err := loadCodebook(app, g.GetString("study"))
	if err != nil {
		return nil, err
	}
	root := cb.byID[g.GetString("ddi_id")]
	if root == nil || root.Name != "varGrp" {
		return nil, fmt.Errorf("group %s: %w", g.Id, ErrNotInCodebook)
	}

	keep := map[string]bool{}
	queue := []*ddixml.Element{root}
	for len(queue) > 0 {
		grp := queue[0]
		queue = queue[1:]
		if keep[grp.ID] {
			continue
		}
		keep[grp.ID] = true
		for _, id := range strings.Fields(grp.Attr("var")) {
			keep[id] = true
		}
		for _, id := range strings.Fields(grp.Attr("varGrp")) {
			if child := cb.byID[id]; child != nil && child.Name == "varGrp" {
				queue = append(queue, child)
			}
		}
	}

	var out []*ddixml.Element
	for _, c := range cb.children {
		if keep[c.ID] {
			out = append(out, c)
		}
	}
	return ddixml.Wrapped(out, cb.lang)
}

// VariableXML returns the variable's <var>. A member of a grid comes with
// the grid's <varGrp>, narrowed to this one variable, in a <dataDscr>: the
// grid's text is the question's lead-in.
func VariableXML(app core.App, v *core.Record) ([]byte, error) {
	cb, err := loadCodebook(app, v.GetString("study"))
	if err != nil {
		return nil, err
	}
	el := cb.byID[v.GetString("ddi_id")]
	if el == nil || el.Name != "var" {
		return nil, fmt.Errorf("variable %s: %w", v.Id, ErrNotInCodebook)
	}

	if groupID := v.GetString("group"); groupID != "" {
		if grp, err := app.FindRecordById("variable_groups", groupID); err == nil && grp.GetString("type") == "grid" {
			if gel := cb.byID[grp.GetString("ddi_id")]; gel != nil {
				return ddixml.Wrapped([]*ddixml.Element{gel.WithAttr("var", el.ID), el}, cb.lang)
			}
		}
	}
	return ddixml.Fragment([]*ddixml.Element{el}, cb.lang)
}
