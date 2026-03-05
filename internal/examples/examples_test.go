package examples

import (
	"strings"
	"testing"
)

func TestGetAll(t *testing.T) {
	all := GetAll()
	if len(all) == 0 {
		t.Fatal("Expected at least one example")
	}

	expectedTypes := []string{
		"single_choice",
		"multiple_choice",
		"single_choice_other",
		"multiple_choice_other",
		"grid",
		"integer",
		"text",
		"single_choice_long_list",
		"multiple_choice_long_list",
	}

	if len(all) != len(expectedTypes) {
		t.Fatalf("Expected %d examples, got %d", len(expectedTypes), len(all))
	}

	for i, exp := range expectedTypes {
		if all[i].Type != exp {
			t.Errorf("Example %d: expected type %s, got %s", i, exp, all[i].Type)
		}
		if all[i].Label == "" {
			t.Errorf("Example %d (%s): label is empty", i, exp)
		}
		if len(all[i].XLSForm.Survey) == 0 {
			t.Errorf("Example %d (%s): survey is empty", i, exp)
		}
		if all[i].DDI == "" {
			t.Errorf("Example %d (%s): DDI is empty", i, exp)
		}
		if !strings.Contains(all[i].DDI, "<?xml") {
			t.Errorf("Example %d (%s): DDI missing XML declaration", i, exp)
		}
	}
}

func TestGetByType(t *testing.T) {
	ex := GetByType("single_choice")
	if ex == nil {
		t.Fatal("Expected single_choice example")
	}
	if ex.Type != "single_choice" {
		t.Errorf("Expected type single_choice, got %s", ex.Type)
	}

	missing := GetByType("nonexistent")
	if missing != nil {
		t.Error("Expected nil for nonexistent type")
	}
}
