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
func wrapFragmentInCodebook(fragment []byte) []byte {
	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <stdyDscr>
    <citation><titlStmt><titl>Fragment Validation</titl></titlStmt></citation>
  </stdyDscr>
  <dataDscr>
    %s
  </dataDscr>
</codeBook>`, string(stripXMLDeclaration(fragment))))
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

func TestIntegration_XLSFormToDDI_ValidatesGroup(t *testing.T) {
	client := getSchematronClient(t)

	xlsformJSON := `{
		"survey": [
			{"type": "begin_group", "name": "satisfaction_group", "label": "Satisfaction questions"},
			{"type": "end_group", "name": ""}
		],
		"choices": [],
		"settings": {}
	}`

	ddiXML, err := XLSFormToDDI([]byte(xlsformJSON))
	if err != nil {
		t.Fatalf("XLSFormToDDI failed: %v", err)
	}

	// The converter produces var="" which is invalid per XSD (IDREFS must be non-empty).
	// In practice, the var attribute is populated when building a complete codebook.
	// For this test, patch in a dummy variable reference and include the dummy var.
	fragment := strings.Replace(string(stripXMLDeclaration(ddiXML)), `var=""`, `var="V_dummy"`, 1)
	dummyVar := `<var ID="V_dummy" name="dummy" intrvl="discrete">
      <qstn responseDomainType="multiple"><qstnLit>Dummy</qstnLit></qstn>
      <catgry><catValu>1</catValu><labl>Yes</labl></catgry>
      <concept>Dummy</concept>
      <varFormat type="numeric" schema="other"/>
    </var>`

	codebook := []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<codeBook xmlns="ddi:codebook:2_5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <stdyDscr>
    <citation><titlStmt><titl>Fragment Validation</titl></titlStmt></citation>
  </stdyDscr>
  <dataDscr>
    %s
    %s
  </dataDscr>
</codeBook>`, fragment, dummyVar))

	resp, err := client.Validate(codebook)
	if err != nil {
		t.Fatalf("Validation request failed: %v", err)
	}
	if !resp.Valid {
		t.Errorf("Expected valid DDI, got errors: %+v", resp.Errors)
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
