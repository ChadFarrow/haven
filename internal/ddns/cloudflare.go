package ddns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CloudflareProvider updates a DNS A record via the Cloudflare API.
type CloudflareProvider struct {
	Token    string
	ZoneID   string
	RecordID string
	Domain   string
}

type cloudflarePayload struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type cloudflareResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *CloudflareProvider) Update(ctx context.Context, ip string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", c.ZoneID, c.RecordID)

	payload := cloudflarePayload{
		Type:    "A",
		Name:    c.Domain,
		Content: ip,
		TTL:     1, // auto
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cloudflare: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cloudflare: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloudflare: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return fmt.Errorf("cloudflare: read response: %w", err)
	}

	var cfResp cloudflareResponse
	if err := json.Unmarshal(respBody, &cfResp); err != nil {
		return fmt.Errorf("cloudflare: parse response: %w", err)
	}

	if !cfResp.Success {
		errMsg := "unknown error"
		if len(cfResp.Errors) > 0 {
			errMsg = cfResp.Errors[0].Message
		}
		return fmt.Errorf("cloudflare: API error: %s", errMsg)
	}

	return nil
}

func (c *CloudflareProvider) Name() string {
	return "cloudflare"
}
