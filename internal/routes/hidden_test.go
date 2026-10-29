package routes

import (
	"errors"
	"testing"

	"qwacback/internal/importer"
	_ "qwacback/migrations"

	"github.com/clbanning/mxj/v2"
	"github.com/pocketbase/pocketbase/tests"
)

// A hidden field as formtransform 0.10.0 writes it, next to a question.
const hiddenCodebook = `<?xml version="1.0"?>
<codeBook xmlns="ddi:codebook:2_5" version="2.5" xml:lang="de">
  <stdyDscr><citation><titlStmt><titl>Hidden fields</titl></titlStmt></citation></stdyDscr>
  <dataDscr>
    <var ID="V_partner" name="partner" intrvl="discrete">
      <qstn responseDomainType="other" otherResponseDomainType="hidden" seqNo="1"><qstnLit>Partner</qstnLit></qstn>
      <varFormat type="character" schema="other"/>
    </var>
    <var ID="V_age" name="age" intrvl="contin" dcml="0">
      <qstn responseDomainType="numeric" seqNo="2"><qstnLit>Wie alt sind Sie?</qstnLit></qstn>
      <varFormat type="numeric" schema="other"/>
    </var>
  </dataDscr>
</codeBook>`

// #78: a hidden variable is kept as a variable but is no question: not in
// the list, the search or the detail.
func TestHiddenVariableIsNoQuestion(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	mv, err := mxj.NewMapXml([]byte(hiddenCodebook))
	if err != nil {
		t.Fatal(err)
	}
	studyID, err := importer.ImportCodebook(app, mv, []byte(hiddenCodebook))
	if err != nil {
		t.Fatal(err)
	}

	hidden, err := app.FindFirstRecordByFilter("variables", "name = 'partner'")
	if err != nil {
		t.Fatalf("hidden variable not stored: %v", err)
	}
	if got := hidden.GetString("answer_type"); got != "hidden" {
		t.Fatalf("answer_type = %q, want hidden", got)
	}

	qs, err := AssembleQuestions(app, studyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 || qs[0].Name != "age" {
		t.Errorf("questions = %+v, want only age", qs)
	}

	if _, err := QuestionDetail(app, hidden.Id); !errors.Is(err, ErrQuestionNotFound) {
		t.Errorf("detail of hidden variable: err = %v, want ErrQuestionNotFound", err)
	}

	found, err := SearchQuestions(app, "Partner", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range found {
		if q.ID == hidden.Id {
			t.Errorf("search found the hidden variable: %+v", q)
		}
	}
}
