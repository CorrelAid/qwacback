package schematron

// MockClient is a test double for the Schematron client.
type MockClient struct {
	ValidateFunc func(xmlBytes []byte) (*ValidationResponse, error)
}

func (m *MockClient) Validate(xmlBytes []byte) (*ValidationResponse, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(xmlBytes)
	}
	return &ValidationResponse{Valid: true, Errors: nil}, nil
}

func (m *MockClient) Close() {}
