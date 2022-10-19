package watcher

import (
	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog"
)

var _ zerolog.LogArrayMarshaler = (*loggableDNSRecords)(nil)
var _ zerolog.LogObjectMarshaler = (*DNSRecord)(nil)

type DNSRecord struct {
	cloudflare.DNSRecord
}

func (rr DNSRecord) IsProxied() bool {
	return rr.Proxied != nil && *rr.Proxied
}

func (rr DNSRecord) IsEqual(o DNSRecord) bool {
	return rr.Type == o.Type &&
		rr.Name == o.Name &&
		rr.Content == o.Content &&
		rr.TTL == o.TTL &&
		rr.IsProxied() == o.IsProxied()
}

func (rr DNSRecord) MarshalZerologObject(e *zerolog.Event) {
	e.
		Str("id", rr.ID).
		Str("type", rr.Type).
		Str("content", rr.Content).
		Int("ttl", rr.TTL).
		Bool("proxied", rr.IsProxied())
}

type loggableDNSRecords []DNSRecord

func (rrs loggableDNSRecords) MarshalZerologArray(a *zerolog.Array) {
	for _, rr := range rrs {
		a.Object(rr)
	}
}
