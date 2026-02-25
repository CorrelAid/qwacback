package schematron

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	// WaitForWorker blocks until a worker is subscribed to the validation
	// subject or timeout elapses.
	WaitForWorker(timeout time.Duration) error
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

// WaitForWorker polls the validation subject until a worker subscribes or
// timeout elapses. NATS returns ErrNoResponders immediately when no subscriber
// exists, so we can retry cheaply without blocking for the full request timeout.
func (c *NatsClient) WaitForWorker(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		// Use a short probe timeout — ErrNoResponders comes back instantly,
		// ErrTimeout means a subscriber exists but is slow (worker is ready).
		_, err := c.conn.Request(Subject, []byte("{}"), 2*time.Second)
		if !errors.Is(err, nats.ErrNoResponders) {
			return nil
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return fmt.Errorf("timed out waiting %v for schematron worker to become available", timeout)
		}
		log.Printf("Schematron worker not ready yet, retrying in 2s (%.0fs remaining)...", remaining.Seconds())
		time.Sleep(min(2*time.Second, remaining))
	}
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
