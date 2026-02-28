package ddns

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var ipServices = []string{
	"https://api.ipify.org",
	"https://icanhazip.com",
	"https://ifconfig.me/ip",
}

// GetPublicIP queries multiple services to detect the current public IP address.
// Returns the first valid IP found, with a 5s timeout per service.
func GetPublicIP(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	for _, svc := range ipServices {
		ip, err := queryIPService(ctx, client, svc)
		if err == nil {
			return ip, nil
		}
	}

	return "", fmt.Errorf("failed to detect public IP from all services")
}

func queryIPService(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP from %s: %q", url, ip)
	}

	return ip, nil
}
