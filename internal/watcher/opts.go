package watcher

import (
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type Option func(*Watcher)

func WithCloudflareAPI(cfc *cloudflare.API) Option {
	return func(watcher *Watcher) {
		watcher.cfc = cfc
	}
}

func WithZoneID(zoneID string) Option {
	return func(watcher *Watcher) {
		watcher.zoneID = zoneID
	}
}

func WithDNSName(name string) Option {
	return func(watcher *Watcher) {
		watcher.dnsName = name
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(watcher *Watcher) {
		watcher.ttl = int(ttl.Seconds())
	}
}

func WithProxied(proxied bool) Option {
	return func(watcher *Watcher) {
		watcher.proxied = proxied
	}
}

func WithIPv4(enabled bool) Option {
	return func(watcher *Watcher) {
		watcher.useIPv4 = enabled
	}
}

func WithIPv6(enabled bool) Option {
	return func(watcher *Watcher) {
		watcher.useIPv6 = enabled
	}
}
