package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// startDynamicDNS periodically checks the public IP and updates the DNS record
// if it has changed. It runs until the context is cancelled.
func startDynamicDNS(ctx context.Context, cfg *DDNSConfig) {
	if cfg == nil {
		return
	}

	log.Printf("🌐 Dynamic DNS enabled (provider: %s, domain: %s, interval: %s)",
		cfg.Provider, cfg.Domain, cfg.UpdateInterval)

	// Do an initial update immediately at startup
	lastIP := ""
	if ip, err := updateDDNSIfChanged(ctx, cfg, lastIP); err != nil {
		log.Printf("🌐 DDNS initial update failed: %v", err)
	} else {
		lastIP = ip
	}

	ticker := time.NewTicker(cfg.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🌐 Dynamic DNS updater stopped")
			return
		case <-ticker.C:
			if ip, err := updateDDNSIfChanged(ctx, cfg, lastIP); err != nil {
				log.Printf("🌐 DDNS update failed: %v", err)
			} else {
				lastIP = ip
			}
		}
	}
}

// updateDDNSIfChanged checks the current public IP and updates DNS if it differs
// from lastIP. Returns the current IP.
func updateDDNSIfChanged(ctx context.Context, cfg *DDNSConfig, lastIP string) (string, error) {
	currentIP, err := getPublicIP(ctx)
	if err != nil {
		return lastIP, fmt.Errorf("failed to detect public IP: %w", err)
	}

	if currentIP == lastIP {
		return currentIP, nil
	}

	if lastIP == "" {
		log.Printf("🌐 DDNS detected public IP: %s", currentIP)
	} else {
		log.Printf("🌐 DDNS IP changed: %s -> %s", lastIP, currentIP)
	}

	if err := updateDNSRecord(ctx, cfg, currentIP); err != nil {
		return lastIP, fmt.Errorf("failed to update DNS record: %w", err)
	}

	log.Printf("🌐 DDNS updated %s to %s", cfg.Domain, currentIP)
	return currentIP, nil
}

// getPublicIP queries external services to determine the current public IP.
// Tries multiple services for reliability.
func getPublicIP(ctx context.Context) (string, error) {
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	for _, svc := range services {
		ip, err := fetchIP(ctx, svc)
		if err == nil && ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("all IP detection services failed")
}

func fetchIP(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

// updateDNSRecord dispatches to the appropriate provider to update the DNS record.
func updateDNSRecord(ctx context.Context, cfg *DDNSConfig, ip string) error {
	switch cfg.Provider {
	case "duckdns":
		return updateDuckDNS(ctx, cfg, ip)
	case "cloudflare":
		return updateCloudflare(ctx, cfg, ip)
	case "noip":
		return updateNoIP(ctx, cfg, ip)
	default:
		return fmt.Errorf("unknown DDNS provider: %s", cfg.Provider)
	}
}

// updateDuckDNS updates a DuckDNS domain record.
// API: https://www.duckdns.org/spec.jsp
func updateDuckDNS(ctx context.Context, cfg *DDNSConfig, ip string) error {
	// DuckDNS wants just the subdomain part (without .duckdns.org)
	domain := strings.TrimSuffix(cfg.Domain, ".duckdns.org")

	url := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&ip=%s",
		domain, cfg.DuckDNSToken, ip)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return err
	}

	result := strings.TrimSpace(string(body))
	if result != "OK" {
		return fmt.Errorf("DuckDNS returned: %s", result)
	}

	return nil
}

// updateCloudflare updates a Cloudflare DNS A record.
// API: https://developers.cloudflare.com/api/resources/dns/subresources/records/methods/update/
func updateCloudflare(ctx context.Context, cfg *DDNSConfig, ip string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s",
		cfg.CloudflareZoneID, cfg.CloudflareRecordID)

	payload := map[string]interface{}{
		"type":    "A",
		"name":    cfg.Domain,
		"content": ip,
		"ttl":     120,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.CloudflareAPIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("Cloudflare API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse Cloudflare response: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("Cloudflare API returned success=false")
	}

	return nil
}

// updateNoIP updates a No-IP / DynDNS2 compatible hostname.
// Protocol: https://www.noip.com/integrate/request
func updateNoIP(ctx context.Context, cfg *DDNSConfig, ip string) error {
	url := fmt.Sprintf("https://dynupdate.no-ip.com/nic/update?hostname=%s&myip=%s",
		cfg.Domain, ip)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(cfg.NoIPUsername, cfg.NoIPPassword)
	req.Header.Set("User-Agent", config.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return err
	}

	result := strings.TrimSpace(string(body))
	if strings.HasPrefix(result, "good") || strings.HasPrefix(result, "nochg") {
		return nil
	}

	return fmt.Errorf("No-IP returned: %s", result)
}
