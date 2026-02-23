package schematron

// ValidationRequest is sent to the Java worker via NATS.
type ValidationRequest struct {
	RequestID string `json:"request_id"`
	XML       string `json:"xml"` // base64 encoded
}

// ValidationResponse is received from the Java worker.
type ValidationResponse struct {
	RequestID string            `json:"request_id"`
	Valid     bool              `json:"valid"`
	Errors    []ValidationError `json:"errors"`
}

// ValidationError describes a single Schematron rule failure.
type ValidationError struct {
	Rule     string `json:"rule"`
	Test     string `json:"test"`
	Location string `json:"location"`
	Message  string `json:"message"`
}
