package mqtt

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

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
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

func (p *Publisher) PublishServiceCall(ctx context.Context, payload event.ServiceCallRequestedPayload) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("mqtt publisher is nil")
	}
	if strings.TrimSpace(payload.Topic) == "" {
		return fmt.Errorf("mqtt topic is required")
	}
	if payload.Input == nil {
		payload.Input = map[string]any{}
	}

	servicePayload, err := json.Marshal(map[string]any{
		"command_id":   payload.CommandID,
		"service_name": payload.ServiceName,
		"input":        payload.Input,
	})
	if err != nil {
		return fmt.Errorf("encode service call payload: %w", err)
	}

	body, err := json.Marshal(map[string]any{
		"topic":   payload.Topic,
		"payload": string(servicePayload),
		"qos":     1,
		"retain":  false,
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

func (p *Publisher) PublishPropertySet(ctx context.Context, payload event.PropertySetRequestedPayload) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("mqtt publisher is nil")
	}
	if strings.TrimSpace(payload.Topic) == "" {
		return fmt.Errorf("mqtt topic is required")
	}
	if len(payload.Properties) == 0 {
		return fmt.Errorf("mqtt properties are required")
	}

	propertyPayload, err := json.Marshal(map[string]any{
		"command_id": payload.CommandID,
		"properties": payload.Properties,
	})
	if err != nil {
		return fmt.Errorf("encode property set payload: %w", err)
	}

	body, err := json.Marshal(map[string]any{
		"topic":   payload.Topic,
		"payload": string(propertyPayload),
		"qos":     1,
		"retain":  false,
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
