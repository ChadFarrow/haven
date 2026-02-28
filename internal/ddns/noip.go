package ddns

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// NoIPProvider updates a No-IP dynamic DNS hostname.
type NoIPProvider struct {
	Username string
	Password string
	Hostname string
}

func (n *NoIPProvider) Update(ctx context.Context, ip string) error {
	url := fmt.Sprintf("https://dynupdate.no-ip.com/nic/update?hostname=%s&myip=%s", n.Hostname, ip)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("noip: create request: %w", err)
	}

	req.SetBasicAuth(n.Username, n.Password)
	req.Header.Set("User-Agent", "Haven DDNS/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("noip: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return fmt.Errorf("noip: read response: %w", err)
	}

	result := strings.TrimSpace(string(body))
	if !strings.HasPrefix(result, "good") && !strings.HasPrefix(result, "nochg") {
		return fmt.Errorf("noip: update failed, response: %q", result)
	}

	return nil
}

func (n *NoIPProvider) Name() string {
	return "noip"
}
