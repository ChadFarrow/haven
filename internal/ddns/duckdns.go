package ddns

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DuckDNSProvider updates a DuckDNS hostname.
type DuckDNSProvider struct {
	Token  string
	Domain string
}

func (d *DuckDNSProvider) Update(ctx context.Context, ip string) error {
	url := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&ip=%s", d.Domain, d.Token, ip)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("duckdns: create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("duckdns: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return fmt.Errorf("duckdns: read response: %w", err)
	}

	if strings.TrimSpace(string(body)) != "OK" {
		return fmt.Errorf("duckdns: update failed, response: %q", string(body))
	}

	return nil
}

func (d *DuckDNSProvider) Name() string {
	return "duckdns"
}
