package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"qwacback/internal/ddixml"
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
// other error from XLSFormToDDI or DDIToXLSForm describes a problem with the
// input and is safe to show to the API client.
var ErrConverterUnavailable = errors.New("DDI/XLSForm converter unavailable")

// XLSFormToDDIRequest is the JSON payload POSTed to the ddi-emitter sidecar.
// The sheets are forwarded as sent, not decoded into fixed-column structs:
// those have fixed columns and would drop label::<lang> and hint::<lang>
// (multilingual forms), default_language and anything else formtransform
// reads (#33).
type XLSFormToDDIRequest struct {
	Survey   []json.RawMessage `json:"survey"`
	Choices  []json.RawMessage `json:"choices"`
	Settings json.RawMessage   `json:"settings,omitempty"`
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
// children of its <dataDscr> unchanged (see ddixml.DataDscr): a single <var>
// or <varGrp> is returned as a bare fragment, anything else is wrapped in
// <dataDscr> (ddixml.Fragment).
func XLSFormToDDI(xlsformJSON []byte) ([]byte, error) {
	if len(bytes.TrimSpace(xlsformJSON)) == 0 {
		return nil, fmt.Errorf("empty request body")
	}

	var form XLSFormToDDIRequest
	if err := json.Unmarshal(xlsformJSON, &form); err != nil {
		return nil, fmt.Errorf("failed to parse XLSForm JSON: %w", err)
	}
	if len(form.Survey) == 0 {
		return nil, fmt.Errorf("survey sheet is empty")
	}
	if form.Choices == nil {
		form.Choices = []json.RawMessage{}
	}

	payload, err := json.Marshal(form)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XLSForm payload: %w", err)
	}

	codebook, err := callDDIEmitter("/xlsform-to-ddi", "application/json", payload)
	if err != nil {
		return nil, err
	}

	// Copied, not parsed into Go structs: whatever formtransform emits
	// reaches the client, including elements qwacback doesn't model. The
	// codeBook's xml:lang (the base language of a multilingual form) moves
	// to the fragment's root.
	lang, children, err := ddixml.DataDscr(codebook)
	if err != nil {
		return nil, fmt.Errorf("%w: ddi-emitter returned %v", ErrConverterUnavailable, err)
	}
	if len(children) == 0 {
		// Notes store no response, so formtransform emits no <var> for them.
		return nil, fmt.Errorf("the form has no questions that store an answer (notes produce no DDI variables)")
	}

	return ddixml.Fragment(children, lang)
}

// DDIToXLSForm converts DDI XML (a whole codebook, a <dataDscr>, or bare
// <var>/<varGrp> elements) to XLSForm JSON with formtransform's ddiToXlsform,
// via the ddi-emitter sidecar (#37):
//
//	{"survey": [...], "choices": [...], "settings": [...], "warnings": [{"code", "message"}]}
//
// The sheets are formtransform's, passed through unchanged. A codebook
// formtransform wrote gives back its form; other DDI converts as far as its
// standard elements go, with a warning for each field it can't supply.
func DDIToXLSForm(ddiXML []byte) ([]byte, error) {
	if len(bytes.TrimSpace(ddiXML)) == 0 {
		return nil, fmt.Errorf("empty request body")
	}
	return callDDIEmitter("/ddi-to-xlsform", "application/xml", ddiXML)
}

func callDDIEmitter(path, contentType string, payload []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ddiEmitterHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(DDIEmitterURL, "/")+path,
		bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrConverterUnavailable, err)
	}
	req.Header.Set("content-type", contentType)

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
