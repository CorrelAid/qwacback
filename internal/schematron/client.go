package schematron

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	Subject        = "schematron.validate"
	DefaultTimeout = 30 * time.Second
)

// Client defines the interface for Schematron validation.
type Client interface {
	Validate(xmlBytes []byte) (*ValidationResponse, error)
	Close()
}

// NatsClient implements Client using NATS request-reply.
type NatsClient struct {
	conn    *nats.Conn
	timeout time.Duration
}

// NewNatsClient creates a new NATS-backed Schematron client.
// Pass a non-empty token to authenticate against the NATS server.
func NewNatsClient(natsURL, token string) (*NatsClient, error) {
	opts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
	}
	if token != "" {
		opts = append(opts, nats.Token(token))
	}
	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	return &NatsClient{conn: nc, timeout: DefaultTimeout}, nil
}

func (c *NatsClient) Validate(xmlBytes []byte) (*ValidationResponse, error) {
	req := ValidationRequest{
		RequestID: uuid.New().String(),
		XML:       base64.StdEncoding.EncodeToString(xmlBytes),
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	msg, err := c.conn.Request(Subject, reqBytes, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("schematron validation request failed: %w", err)
	}

	var resp ValidationResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *NatsClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
