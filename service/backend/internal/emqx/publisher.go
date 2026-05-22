package emqx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Publisher struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

type PublisherOptions struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type PublishInput struct {
	Topic   string
	Payload []byte
	QOS     int
	Retain  bool
}

func NewPublisher(opt PublisherOptions) (*Publisher, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(opt.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("emqx api url is required")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid emqx api url %q: %w", baseURL, err)
	}
	opt.APIKey = strings.TrimSpace(opt.APIKey)
	opt.APISecret = strings.TrimSpace(opt.APISecret)
	if opt.APIKey == "" || opt.APISecret == "" {
		return nil, fmt.Errorf("emqx api credentials are required")
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 3 * time.Second
	}
	client := opt.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: opt.Timeout}
	}
	return &Publisher{
		baseURL:    baseURL,
		apiKey:     opt.APIKey,
		apiSecret:  opt.APISecret,
		httpClient: client,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, in PublishInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("emqx publisher is nil")
	}
	in.Topic = strings.TrimSpace(in.Topic)
	if in.Topic == "" {
		return fmt.Errorf("mqtt topic is required")
	}
	if len(in.Payload) == 0 {
		return fmt.Errorf("mqtt payload is required")
	}
	if in.QOS < 0 || in.QOS > 2 {
		return fmt.Errorf("mqtt qos must be between 0 and 2")
	}

	body, err := json.Marshal(map[string]any{
		"topic":   in.Topic,
		"payload": string(in.Payload),
		"qos":     in.QOS,
		"retain":  in.Retain,
	})
	if err != nil {
		return fmt.Errorf("encode emqx publish request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v5/publish", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create emqx publish request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(p.apiKey, p.apiSecret)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call emqx publish api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("emqx publish api returned %s: %s", resp.Status, strings.TrimSpace(string(raw)))
}
