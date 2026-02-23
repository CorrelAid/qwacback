package schematron

import (
	"fmt"
	"testing"
)

func TestMockClientDefaultPasses(t *testing.T) {
	mock := &MockClient{}
	resp, err := mock.Validate([]byte("<xml/>"))
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Valid {
		t.Error("expected mock to return valid by default")
	}
}

func TestMockClientCanReturnErrors(t *testing.T) {
	mock := &MockClient{
		ValidateFunc: func(xmlBytes []byte) (*ValidationResponse, error) {
			return &ValidationResponse{
				Valid: false,
				Errors: []ValidationError{
					{Rule: "test", Message: "something wrong"},
				},
			}, nil
		},
	}
	resp, err := mock.Validate([]byte("<xml/>"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Valid {
		t.Error("expected invalid")
	}
	if len(resp.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(resp.Errors))
	}
}

func TestMockClientCanReturnError(t *testing.T) {
	mock := &MockClient{
		ValidateFunc: func(xmlBytes []byte) (*ValidationResponse, error) {
			return nil, fmt.Errorf("nats timeout")
		},
	}
	_, err := mock.Validate([]byte("<xml/>"))
	if err == nil {
		t.Error("expected error")
	}
}
