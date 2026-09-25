package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
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

// XLSFormToDDIRequest is the JSON payload POSTed to the ddi-emitter sidecar.
// Mirrors the shape formtransform's buildDdiXml expects internally.
type XLSFormToDDIRequest struct {
	Survey   []SurveyRow    `json:"survey"`
	Choices  []ChoiceRow    `json:"choices"`
	Settings SettingsRow    `json:"settings"`
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
// The sidecar returns a full <codeBook> document; this function extracts the
// <dataDscr> and applies the same single-element unwrapping the Go converter
// used to: a single <var> or <varGrp> is returned as a bare fragment, anything
// else is wrapped in <dataDscr>. The <codeBook> framing is never returned, so
// the public API of /api/convert/xlsform-to-ddi is unchanged.
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

	dataDscr, err := extractDataDscr(codebook)
	if err != nil {
		return nil, err
	}

	return shapeDDIFragment(dataDscr)
}

func callDDIEmitter(payload []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ddiEmitterHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(DDIEmitterURL, "/")+"/xlsform-to-ddi",
		bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("ddi-emitter: build request: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ddi-emitter unreachable at %s: %w", DDIEmitterURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("ddi-emitter: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr DDIEmitterError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
			return nil, fmt.Errorf("ddi-emitter: %s", apiErr.Error)
		}
		return nil, fmt.Errorf("ddi-emitter: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

// codeBookEnvelope is a minimal view of formtransform's full <codeBook> output
// large enough to extract its <dataDscr> child for re-shaping.
type codeBookEnvelope struct {
	XMLName xml.Name    `xml:"codeBook"`
	DataDscr DDIDataDscr `xml:"dataDscr"`
}

func extractDataDscr(codebook []byte) (DDIDataDscr, error) {
	var cb codeBookEnvelope
	dec := xml.NewDecoder(bytes.NewReader(codebook))
	dec.Strict = false
	if err := dec.Decode(&cb); err != nil {
		return DDIDataDscr{}, fmt.Errorf("ddi-emitter returned malformed codeBook: %w", err)
	}
	return cb.DataDscr, nil
}

// shapeDDIFragment mirrors the unwrapping rules the deleted XLSFormToDDI used:
//   - 1 var, 0 varGrps  → bare <var>
//   - 0 vars, 1 varGrp  → bare <varGrp>
//   - everything else   → wrapped in <dataDscr>
func shapeDDIFragment(dd DDIDataDscr) ([]byte, error) {
	switch {
	case len(dd.Vars) == 1 && len(dd.VarGrps) == 0:
		out, err := xml.MarshalIndent(dd.Vars[0], "", "  ")
		if err != nil {
			return nil, err
		}
		return append([]byte(xml.Header), out...), nil
	case len(dd.VarGrps) == 1 && len(dd.Vars) == 0:
		out, err := xml.MarshalIndent(dd.VarGrps[0], "", "  ")
		if err != nil {
			return nil, err
		}
		return append([]byte(xml.Header), out...), nil
	default:
		out, err := xml.MarshalIndent(dd, "", "  ")
		if err != nil {
			return nil, err
		}
		return append([]byte(xml.Header), out...), nil
	}
}
