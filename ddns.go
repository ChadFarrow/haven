package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/barrydeen/haven/internal/ddns"
)

func startPeriodicDDNS(ctx context.Context) {
	provider, err := getDDNSProvider()
	if err != nil {
		log.Printf("🌐 DDNS disabled: %v", err)
		return
	}

	log.Printf("🌐 DDNS enabled (provider: %s, interval: %s)", provider.Name(), config.DDNSCheckInterval)

	var lastIP string

	// Check immediately on startup
	if ip, err := ddns.GetPublicIP(ctx); err != nil {
		log.Printf("⚠️ DDNS: failed to detect public IP: %v", err)
	} else {
		log.Printf("🌐 DDNS: detected public IP: %s", ip)
		if err := provider.Update(ctx, ip); err != nil {
			log.Printf("⚠️ DDNS: update failed: %v", err)
		} else {
			log.Printf("✅ DDNS: updated %s to %s", config.DDNSDomain, ip)
			lastIP = ip
		}
	}

	ticker := time.NewTicker(config.DDNSCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ip, err := ddns.GetPublicIP(ctx)
			if err != nil {
				log.Printf("⚠️ DDNS: failed to detect public IP: %v", err)
				continue
			}

			if ip == lastIP {
				continue
			}

			log.Printf("🌐 DDNS: IP changed from %s to %s", lastIP, ip)
			if err := provider.Update(ctx, ip); err != nil {
				log.Printf("⚠️ DDNS: update failed: %v", err)
				continue
			}

			log.Printf("✅ DDNS: updated %s to %s", config.DDNSDomain, ip)
			lastIP = ip
		}
	}
}

func getDDNSProvider() (ddns.Provider, error) {
	switch config.DDNSProvider {
	case "cloudflare":
		return &ddns.CloudflareProvider{
			Token:    config.DDNSToken,
			ZoneID:   config.DDNSCloudflareZoneID,
			RecordID: config.DDNSCloudflareRecordID,
			Domain:   config.DDNSDomain,
		}, nil
	case "duckdns":
		return &ddns.DuckDNSProvider{
			Token:  config.DDNSToken,
			Domain: config.DDNSDomain,
		}, nil
	case "noip":
		return &ddns.NoIPProvider{
			Username: config.DDNSUsername,
			Password: config.DDNSPassword,
			Hostname: config.DDNSDomain,
		}, nil
	case "none", "":
		return nil, fmt.Errorf("no DDNS provider configured")
	default:
		return nil, fmt.Errorf("unknown DDNS provider: %q", config.DDNSProvider)
	}
}
