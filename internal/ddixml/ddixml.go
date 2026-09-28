// Package ddixml copies DDI elements as XML token streams, so every element
// and attribute reaches the output, including ones qwacback doesn't model.
package ddixml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// DDINamespace is the default namespace of a DDI Codebook 2.5 document.
const DDINamespace = "ddi:codebook:2_5"

// XMLNamespace is the namespace of the xml: prefix (xml:lang).
const XMLNamespace = "http://www.w3.org/XML/1998/namespace"

// Element is one element with everything inside it, as tokens.
type Element struct {
	Name   string // local name
	ID     string // @ID
	Tokens []xml.Token
}

// Start returns the element's start tag.
func (e *Element) Start() xml.StartElement { return e.Tokens[0].(xml.StartElement) }

// Attr returns the value of an attribute without namespace, or "".
func (e *Element) Attr(name string) string {
	for _, a := range e.Start().Attr {
		if a.Name.Space == "" && a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// WithAttr returns a copy of the element with an attribute set.
func (e *Element) WithAttr(name, value string) *Element {
	start := e.Start()
	attrs := make([]xml.Attr, 0, len(start.Attr)+1)
	set := false
	for _, a := range start.Attr {
		if a.Name.Space == "" && a.Name.Local == name {
			a.Value, set = value, true
		}
		attrs = append(attrs, a)
	}
	if !set {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: name}, Value: value})
	}
	start.Attr = attrs
	out := &Element{Name: e.Name, ID: e.ID, Tokens: append([]xml.Token{start}, e.Tokens[1:]...)}
	return out
}

// Lang returns the xml:lang attribute of a start tag, or "".
func Lang(t xml.StartElement) string {
	for _, a := range t.Attr {
		if a.Name.Space == XMLNamespace && a.Name.Local == "lang" {
			return a.Value
		}
	}
	return ""
}

// DataDscr returns codeBook/@xml:lang and the children of <codeBook><dataDscr>
// in document order. Each is cleaned for use outside its codebook: the DDI
// default namespace and `files` attributes (IDREFs into <fileDscr>) are
// dropped, as is whitespace between elements.
func DataDscr(codebook []byte) (lang string, children []*Element, err error) {
	dec := xml.NewDecoder(bytes.NewReader(codebook))
	var path []string // local names of open elements
	found := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, fmt.Errorf("malformed XML: %w", err)
		}
		inDataDscr := len(path) >= 2 && path[0] == "codeBook" && path[1] == "dataDscr"

		switch t := tok.(type) {
		case xml.StartElement:
			if len(path) == 0 && t.Name.Local == "codeBook" {
				lang = Lang(t)
			}
			if len(path) == 1 && path[0] == "codeBook" && t.Name.Local == "dataDscr" {
				found = true
			}
			if inDataDscr {
				if len(path) == 2 {
					el := &Element{Name: t.Name.Local}
					for _, a := range t.Attr {
						if a.Name.Space == "" && a.Name.Local == "ID" {
							el.ID = a.Value
						}
					}
					children = append(children, el)
				}
				last := children[len(children)-1]
				last.Tokens = append(last.Tokens, CleanStart(t))
			}
			path = append(path, t.Name.Local)
		case xml.EndElement:
			path = path[:len(path)-1]
			// Still inside <dataDscr> after closing: t closed one of its descendants.
			if len(path) >= 2 && path[0] == "codeBook" && path[1] == "dataDscr" {
				last := children[len(children)-1]
				last.Tokens = append(last.Tokens, xml.EndElement{Name: CleanName(t.Name)})
			}
		case xml.CharData:
			if inDataDscr && len(path) > 2 && len(bytes.TrimSpace(t)) > 0 {
				last := children[len(children)-1]
				last.Tokens = append(last.Tokens, t.Copy())
			}
		}
	}
	if !found {
		return "", nil, fmt.Errorf("no <codeBook><dataDscr>")
	}
	return lang, children, nil
}

func CleanName(n xml.Name) xml.Name {
	if n.Space == DDINamespace {
		n.Space = ""
	}
	return n
}

func CleanStart(t xml.StartElement) xml.StartElement {
	out := xml.StartElement{Name: CleanName(t.Name)}
	for _, a := range t.Attr {
		if a.Name.Space == "xmlns" || (a.Name.Space == "" && a.Name.Local == "xmlns") {
			continue
		}
		if a.Name.Space == "" && a.Name.Local == "files" {
			continue
		}
		out.Attr = append(out.Attr, a)
	}
	return out
}

// Fragment writes elements as a DDI fragment: a single <var> or <varGrp>
// bare, anything else wrapped in <dataDscr>, order kept. A non-empty lang
// (the codebook's base language) goes on the root as xml:lang, so it isn't
// lost with the <codeBook>.
func Fragment(children []*Element, lang string) ([]byte, error) {
	bare := len(children) == 1 && (children[0].Name == "var" || children[0].Name == "varGrp")
	return write(children, lang, bare)
}

// Wrapped is Fragment, but always wraps in <dataDscr>, even one element.
func Wrapped(children []*Element, lang string) ([]byte, error) {
	return write(children, lang, false)
}

func write(children []*Element, lang string, bare bool) ([]byte, error) {
	langAttr := xml.Attr{Name: xml.Name{Space: XMLNamespace, Local: "lang"}, Value: lang}
	wrapper := xml.StartElement{Name: xml.Name{Local: "dataDscr"}}
	first := 0 // tokens of children[0] to skip when its start tag is replaced
	var toks []xml.Token
	if lang != "" {
		if bare {
			root := children[0].Start()
			root.Attr = append([]xml.Attr{langAttr}, root.Attr...)
			toks = append(toks, root)
			first = 1
		} else {
			wrapper.Attr = []xml.Attr{langAttr}
		}
	}
	if !bare {
		toks = append([]xml.Token{wrapper}, toks...)
	}
	for i, c := range children {
		if i == 0 {
			toks = append(toks, c.Tokens[first:]...)
		} else {
			toks = append(toks, c.Tokens...)
		}
	}
	if !bare {
		toks = append(toks, wrapper.End())
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	for _, tok := range indent(toks) {
		if err := enc.EncodeToken(tok); err != nil {
			return nil, err
		}
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// indent lays tokens out with two-space indentation between elements, but
// leaves elements with text in them (mixed content, like the XHTML
// <p>…<em>…</em>…</p> of a question text) exactly as they are. Encoder.Indent
// would add a line break before the <em> and so change the question text.
func indent(toks []xml.Token) []xml.Token {
	// Which elements have text of their own, by the index of their start tag.
	mixed := map[int]bool{}
	var open []int
	for i, tok := range toks {
		switch t := tok.(type) {
		case xml.StartElement:
			open = append(open, i)
		case xml.EndElement:
			open = open[:len(open)-1]
		case xml.CharData:
			if len(open) > 0 && len(bytes.TrimSpace(t)) > 0 {
				mixed[open[len(open)-1]] = true
			}
		}
	}

	type frame struct{ inline, hasChildren bool }
	var out []xml.Token
	var stack []frame
	newline := func(depth int) xml.CharData {
		return xml.CharData("\n" + string(bytes.Repeat([]byte("  "), depth)))
	}
	for i, tok := range toks {
		inline := len(stack) > 0 && stack[len(stack)-1].inline
		switch t := tok.(type) {
		case xml.StartElement:
			if !inline {
				if len(stack) > 0 {
					stack[len(stack)-1].hasChildren = true
				}
				if len(out) > 0 {
					out = append(out, newline(len(stack)))
				}
			}
			stack = append(stack, frame{inline: inline || mixed[i]})
			out = append(out, t)
		case xml.EndElement:
			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !f.inline && f.hasChildren {
				out = append(out, newline(len(stack)))
			}
			out = append(out, t)
		case xml.CharData:
			// Whitespace between elements is replaced by the indentation.
			if inline || len(bytes.TrimSpace(t)) > 0 {
				out = append(out, t)
			}
		default:
			out = append(out, tok)
		}
	}
	return out
}
