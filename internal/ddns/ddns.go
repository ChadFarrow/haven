package ddns

import "context"

// Provider is an interface for updating DNS records with a new IP address.
type Provider interface {
	Update(ctx context.Context, ip string) error
	Name() string
}
