package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DDIEmitterURL is the base URL of the ddi-emitter sidecar. Overridden by the
// DDI_EMITTER_URL environment variable; defaults to localhost on the standard
// port for both dev (go run) and docker-compose.
var DDIEmitterURL = func() string {
	if v := os.Getenv("DDI_EMITTER_URL"); v != "" {
		return v
	}
	return "http://127.0.0.1:8091"
}()

const ddiEmitterHTTPTimeout = 30 * time.Second

// ErrConverterUnavailable means the ddi-emitter sidecar couldn't be reached
// or answered with something other than a result or an input rejection. It is
// a server-side problem: callers should answer 503, not blame the input. Every
// other error from XLSFormToDDI describes a problem with the input and is safe
// to show to the API client.
var ErrConverterUnavailable = errors.New("XLSForm to DDI converter unavailable")

// ddiNamespace is the default namespace of formtransform's <codeBook>. It is
// dropped from the returned fragment, which is not a standalone document.
const ddiNamespace = "ddi:codebook:2_5"

// XLSFormToDDIRequest is the JSON payload POSTed to the ddi-emitter sidecar.
// Mirrors the shape formtransform's buildDdiXml expects internally.
type XLSFormToDDIRequest struct {
	Survey   []SurveyRow `json:"survey"`
	Choices  []ChoiceRow `json:"choices"`
	Settings SettingsRow `json:"settings"`
}

// DDIEmitterError is returned by the sidecar when it rejects input. The 400
// detail carries the library's error message (it names the question and the
// reason), so it surfaces to the API client unchanged.
type DDIEmitterError struct {
	Error string `json:"error"`
}

// XLSFormToDDI converts XLSForm sheet-based JSON to DDI XML by delegating to
// the ddi-emitter sidecar (which wraps @correlaid/formtransform).
//
// The sidecar returns a full <codeBook> document; this function cuts out the
// children of its <dataDscr> unchanged (see extractDataDscr) and applies the
// same single-element unwrapping the Go converter used to: a single <var> or
// <varGrp> is returned as a bare fragment, anything else is wrapped in
// <dataDscr>. The <codeBook> framing is never returned, so the public API of
// /api/convert/xlsform-to-ddi is unchanged.
func XLSFormToDDI(xlsformJSON []byte) ([]byte, error) {
	if len(bytes.TrimSpace(xlsformJSON)) == 0 {
		return nil, fmt.Errorf("empty request body")
	}

	var form XLSForm
	if err := json.Unmarshal(xlsformJSON, &form); err != nil {
		return nil, fmt.Errorf("failed to parse XLSForm JSON: %w", err)
	}
	if len(form.Survey) == 0 {
		return nil, fmt.Errorf("survey sheet is empty")
	}

	payload, err := json.Marshal(XLSFormToDDIRequest{
		Survey:   form.Survey,
		Choices:  form.Choices,
		Settings: form.Settings,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XLSForm payload: %w", err)
	}

	codebook, err := callDDIEmitter(payload)
	if err != nil {
		return nil, err
	}

	children, err := extractDataDscr(codebook)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConverterUnavailable, err)
	}
	if len(children) == 0 {
		// Notes store no response, so formtransform emits no <var> for them.
		return nil, fmt.Errorf("the form has no questions that store an answer (notes produce no DDI variables)")
	}

	return shapeDDIFragment(children)
}

func callDDIEmitter(payload []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ddiEmitterHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(DDIEmitterURL, "/")+"/xlsform-to-ddi",
		bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrConverterUnavailable, err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s unreachable: %v", ErrConverterUnavailable, DDIEmitterURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %v", ErrConverterUnavailable, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusBadRequest:
		// The library rejected the input; its message names the question.
		var apiErr DDIEmitterError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
			return nil, errors.New(apiErr.Error)
		}
		return nil, errors.New(strings.TrimSpace(string(body)))
	default:
		return nil, fmt.Errorf("%w: HTTP %d: %s", ErrConverterUnavailable, resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// fragmentChild is one element directly under <dataDscr>, as its token stream.
type fragmentChild struct {
	name   string
	tokens []xml.Token
}

// extractDataDscr returns the children of the <dataDscr> in formtransform's
// <codeBook> as token streams. They are copied, not parsed into Go structs, so
// whatever formtransform emits reaches the client, including elements qwacback
// doesn't model. Only three things change:
//   - the DDI default namespace is dropped (the fragment has no <codeBook> to
//     declare it; the deleted Go converter didn't emit one either),
//   - `files` attributes are dropped: they point at the <fileDscr> in the
//     <codeBook>, which isn't part of the fragment,
//   - whitespace between elements is dropped and re-indented.
func extractDataDscr(codebook []byte) ([]fragmentChild, error) {
	dec := xml.NewDecoder(bytes.NewReader(codebook))
	var children []fragmentChild
	var path []string // local names of open elements
	found := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ddi-emitter returned malformed XML: %w", err)
		}
		// Position of this token relative to <codeBook><dataDscr>.
		inDataDscr := len(path) >= 2 && path[0] == "codeBook" && path[1] == "dataDscr"

		switch t := tok.(type) {
		case xml.StartElement:
			if len(path) == 1 && path[0] == "codeBook" && t.Name.Local == "dataDscr" {
				found = true
			}
			if inDataDscr {
				if len(path) == 2 {
					children = append(children, fragmentChild{name: t.Name.Local})
				}
				last := &children[len(children)-1]
				last.tokens = append(last.tokens, cleanStart(t))
			}
			path = append(path, t.Name.Local)
		case xml.EndElement:
			path = path[:len(path)-1]
			// Still inside <dataDscr> after closing: t closed one of its descendants.
			if len(path) >= 2 && path[0] == "codeBook" && path[1] == "dataDscr" {
				last := &children[len(children)-1]
				last.tokens = append(last.tokens, xml.EndElement{Name: cleanName(t.Name)})
			}
		case xml.CharData:
			if inDataDscr && len(path) > 2 && len(bytes.TrimSpace(t)) > 0 {
				last := &children[len(children)-1]
				last.tokens = append(last.tokens, t.Copy())
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("ddi-emitter returned no <codeBook><dataDscr>")
	}
	return children, nil
}

func cleanName(n xml.Name) xml.Name {
	if n.Space == ddiNamespace {
		n.Space = ""
	}
	return n
}

func cleanStart(t xml.StartElement) xml.StartElement {
	out := xml.StartElement{Name: cleanName(t.Name)}
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

// shapeDDIFragment mirrors the unwrapping rules the deleted XLSFormToDDI used:
//   - a single <var> or <varGrp> → bare element
//   - everything else            → wrapped in <dataDscr>, original order kept
func shapeDDIFragment(children []fragmentChild) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")

	bare := len(children) == 1 && (children[0].name == "var" || children[0].name == "varGrp")
	wrapper := xml.StartElement{Name: xml.Name{Local: "dataDscr"}}
	if !bare {
		if err := enc.EncodeToken(wrapper); err != nil {
			return nil, err
		}
	}
	for _, c := range children {
		for _, tok := range c.tokens {
			if err := enc.EncodeToken(tok); err != nil {
				return nil, err
			}
		}
	}
	if !bare {
		if err := enc.EncodeToken(wrapper.End()); err != nil {
			return nil, err
		}
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
