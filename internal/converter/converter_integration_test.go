package converter

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"qwacback/internal/schematron"
)

// stripXMLDeclaration removes the <?xml ...?> declaration from XML bytes.
func stripXMLDeclaration(xmlBytes []byte) []byte {
	return []byte(strings.TrimPrefix(string(xmlBytes), xml.Header))
}

// wrapFragmentInCodebook wraps a DDI XML fragment in a minimal valid codebook document
// so it can be validated by the Schematron worker.
// If the fragment is already a <dataDscr>, it is placed directly (not nested).
func wrapFragmentInCodebook(fragment []byte) []byte {
	stripped := string(stripXMLDeclaration(fragment))
	trimmed := strings.TrimSpace(stripped)

	// If the fragment is already a <dataDscr>, use it directly instead of wrapping in another <dataDscr>
	if strings.HasPrefix(trimmed, "<dataDscr") {
		return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <stdyDscr>
    <citation><titlStmt><titl>Fragment Validation</titl></titlStmt></citation>
  </stdyDscr>
  %s
</codeBook>`, stripped))
	}

	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <stdyDscr>
    <citation><titlStmt><titl>Fragment Validation</titl></titlStmt></citation>
  </stdyDscr>
  <dataDscr>
    %s
  </dataDscr>
</codeBook>`, stripped))
}

// getSchematronClient returns a connected schematron client or skips the test.
func getSchematronClient(t *testing.T) schematron.Client {
	t.Helper()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set; skipping integration test")
	}

	client, err := schematron.NewNatsClient(natsURL, os.Getenv("NATS_TOKEN"))
	if err != nil {
		t.Skipf("Could not connect to NATS: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	if err := client.WaitForWorker(5 * time.Second); err != nil {
		t.Skipf("Schematron worker not available: %v", err)
	}

	return client
}

func TestIntegration_XLSFormToDDI_ValidatesSelectOne(t *testing.T) {
	client := getSchematronClient(t)

	xlsformJSON := `{
		"survey": [
			{"type": "select_one gender", "name": "gender", "label": "What is your gender?"}
		],
		"choices": [
			{"list_name": "gender", "name": "1", "label": "Male"},
			{"list_name": "gender", "name": "2", "label": "Female"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	codebook := wrapFragmentInCodebook(ddiXML)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI, got errors: %+v", resp.Errors)
	}
}

func TestIntegration_XLSFormToDDI_ValidatesInteger(t *testing.T) {
	client := getSchematronClient(t)

	xlsformJSON := `{
		"survey": [
			{"type": "integer", "name": "age", "label": "What is your age?"}
		],
		"choices": [],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	codebook := wrapFragmentInCodebook(ddiXML)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI, got errors: %+v", resp.Errors)
	}
}

func TestIntegration_XLSFormToDDI_ValidatesText(t *testing.T) {
	client := getSchematronClient(t)

	xlsformJSON := `{
		"survey": [
			{"type": "text", "name": "comments", "label": "Please provide any additional comments", "hint": "Be specific"}
		],
		"choices": [],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	codebook := wrapFragmentInCodebook(ddiXML)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI, got errors: %+v", resp.Errors)
	}
}

func TestIntegration_XLSFormToDDI_ValidatesSelectMultiple(t *testing.T) {
	client := getSchematronClient(t)

	xlsformJSON := `{
		"survey": [
			{"type": "select_multiple hobbies", "name": "hobbies", "label": "What are your hobbies?"}
		],
		"choices": [
			{"list_name": "hobbies", "name": "1", "label": "Reading"},
			{"list_name": "hobbies", "name": "2", "label": "Sports"},
			{"list_name": "hobbies", "name": "3", "label": "Music"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	codebook := wrapFragmentInCodebook(ddiXML)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI, got errors: %+v", resp.Errors)
	}
}

func TestIntegration_XLSFormToDDI_ValidatesSection(t *testing.T) {
	client := getSchematronClient(t)

	// A non-grid begin_group ("section") has no valid DDI representation —
	// qwacback's Schematron restricts varGrp/@type to grid/multipleResp/other.
	// The converter drops the wrapper; members flatten to top-level vars.
	xlsformJSON := `{
		"survey": [
			{"type": "begin_group", "name": "demographics", "label": "Demographics"},
			{"type": "integer", "name": "age", "label": "What is your age?"},
			{"type": "end_group", "name": ""}
		],
		"choices": [],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	codebook := wrapFragmentInCodebook(ddiXML)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI for dropped-section output, got errors: %+v\nXML:\n%s", resp.Errors, string(codebook))
	}
}

func TestIntegration_XLSFormToDDI_ValidatesGridGroup(t *testing.T) {
	client := getSchematronClient(t)

	// Grid group: begin_group with table-list appearance + member select_one vars
	xlsformJSON := `{
		"survey": [
			{"type": "begin_group", "name": "trust_grid", "label": "How much do you trust the following?", "appearance": "table-list"},
			{"type": "select_one trust_scale", "name": "trust_police", "label": "Police"},
			{"type": "select_one trust_scale", "name": "trust_courts", "label": "Courts"},
			{"type": "end_group", "name": ""}
		],
		"choices": [
			{"list_name": "trust_scale", "name": "1", "label": "Not at all"},
			{"list_name": "trust_scale", "name": "2", "label": "Somewhat"},
			{"list_name": "trust_scale", "name": "3", "label": "Very much"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	codebook := wrapFragmentInCodebook(ddiXML)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI for grid group, got errors: %+v\nXML:\n%s", resp.Errors, string(codebook))
	}
}

func TestIntegration_XLSFormToDDI_ValidatesMultipleRespGroup(t *testing.T) {
	client := getSchematronClient(t)

	xlsformJSON := `{
		"survey": [
			{"type": "select_multiple geraete", "name": "geraetebesitz", "label": "Welche Geräte besitzen Sie?"}
		],
		"choices": [
			{"list_name": "geraete", "name": "smartphone", "label": "Smartphone"},
			{"list_name": "geraete", "name": "laptop", "label": "Laptop"},
			{"list_name": "geraete", "name": "tablet", "label": "Tablet"}
		],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	codebook := wrapFragmentInCodebook(ddiXML)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI for multipleResp group, got errors: %+v\nXML:\n%s", resp.Errors, string(codebook))
	}
}

func TestIntegration_GroupCodebook_DDIToXLSFormRoundTrip(t *testing.T) {
	client := getSchematronClient(t)

	// Start with a grid group codebook fragment (varGrp + member vars)
	originalDDI := `<dataDscr>
		<varGrp ID="VG1" name="contact_knowledge" type="grid" var="V1 V2">
			<concept>Civic network awareness</concept>
			<txt>Do you know who to contact in the following groups?</txt>
		</varGrp>
		<var ID="V1" name="contact_community_groups" intrvl="discrete">
			<concept>Civic network awareness</concept>
			<qstn responseDomainType="category">
				<preQTxt>Do you know who to contact in the following groups?</preQTxt>
				<qstnLit>Local Community Groups</qstnLit>
			</qstn>
			<catgry><catValu>1</catValu><labl>Yes</labl></catgry>
			<catgry><catValu>2</catValu><labl>No</labl></catgry>
			<catgry><catValu>3</catValu><labl>Don't Know</labl></catgry>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V2" name="contact_local_council" intrvl="discrete">
			<concept>Civic network awareness</concept>
			<qstn responseDomainType="category">
				<preQTxt>Do you know who to contact in the following groups?</preQTxt>
				<qstnLit>Local Council</qstnLit>
			</qstn>
			<catgry><catValu>1</catValu><labl>Yes</labl></catgry>
			<catgry><catValu>2</catValu><labl>No</labl></catgry>
			<catgry><catValu>3</catValu><labl>Don't Know</labl></catgry>
			<varFormat type="numeric" schema="other"/>
		</var>
	</dataDscr>`

	// DDI → XLSForm
	xlsformJSON, err := DDIToXLSForm([]byte(originalDDI))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	// XLSForm → DDI
	newDDI, err := XLSFormToDDI(xlsformJSON)
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	// Validate round-tripped DDI
	codebook := wrapFragmentInCodebook(newDDI)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI after grid group round-trip, got errors: %+v\nXML:\n%s", resp.Errors, string(codebook))
	}
}

func TestIntegration_MultipleRespCodebook_DDIToXLSFormRoundTrip(t *testing.T) {
	client := getSchematronClient(t)

	// Start with a multipleResp group codebook fragment
	originalDDI := `<dataDscr>
		<varGrp ID="VG1" name="hobbies" type="multipleResp" var="V1 V2 V3">
			<txt>What are your hobbies?</txt>
			<concept>What are your hobbies?</concept>
		</varGrp>
		<var ID="V1" name="hobbies_reading" intrvl="discrete">
			<concept>What are your hobbies?: Reading</concept>
			<qstn responseDomainType="multiple">
				<preQTxt>What are your hobbies?</preQTxt>
				<qstnLit>Reading</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V2" name="hobbies_sports" intrvl="discrete">
			<concept>What are your hobbies?: Sports</concept>
			<qstn responseDomainType="multiple">
				<preQTxt>What are your hobbies?</preQTxt>
				<qstnLit>Sports</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<varFormat type="numeric" schema="other"/>
		</var>
		<var ID="V3" name="hobbies_music" intrvl="discrete">
			<concept>What are your hobbies?: Music</concept>
			<qstn responseDomainType="multiple">
				<preQTxt>What are your hobbies?</preQTxt>
				<qstnLit>Music</qstnLit>
			</qstn>
			<catgry><catValu>0</catValu></catgry>
			<catgry><catValu>1</catValu></catgry>
			<varFormat type="numeric" schema="other"/>
		</var>
	</dataDscr>`

	// DDI → XLSForm
	xlsformJSON, err := DDIToXLSForm([]byte(originalDDI))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	// XLSForm → DDI
	newDDI, err := XLSFormToDDI(xlsformJSON)
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	// Validate round-tripped DDI
	codebook := wrapFragmentInCodebook(newDDI)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI after multipleResp round-trip, got errors: %+v\nXML:\n%s", resp.Errors, string(codebook))
	}
}

func TestIntegration_RoundTrip_ValidatesOutput(t *testing.T) {
	client := getSchematronClient(t)

	originalDDI := `<var ID="V1" name="gender" intrvl="discrete">
		<concept>Gender</concept>
		<qstn responseDomainType="category">
			<qstnLit>What is your gender?</qstnLit>
		</qstn>
		<catgry>
			<catValu>1</catValu>
			<labl>Male</labl>
		</catgry>
		<catgry>
			<catValu>2</catValu>
			<labl>Female</labl>
		</catgry>
		<varFormat type="numeric" schema="other"/>
	</var>`

	// DDI → XLSForm → DDI
	xlsformJSON, err := DDIToXLSForm([]byte(originalDDI))
	if err != nil {
		t.Fatalf("DDIToXLSForm failed: %v", err)
	}

	newDDI, err := XLSFormToDDI(xlsformJSON)
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	codebook := wrapFragmentInCodebook(newDDI)
	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI after round-trip, got errors: %+v", resp.Errors)
	}
}
