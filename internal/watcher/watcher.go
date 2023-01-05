package watcher

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"github.com/buglloc/cf-ddns/internal/xcf"
	"github.com/buglloc/cf-ddns/internal/xhttp"
)

type Watcher struct {
	httpc     *resty.Client
	cfc       *cloudflare.API
	zoneID    string
	dnsName   string
	proxied   bool
	useIPv4   bool
	useIPv6   bool
	ttl       int
	ctx       context.Context
	cancelCtx context.CancelFunc
	closed    chan struct{}
}

func NewWatcher(opts ...Option) (*Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())

	out := &Watcher{
		httpc: resty.New().
			SetRedirectPolicy(resty.NoRedirectPolicy()).
			SetRetryCount(3).
			SetTLSClientConfig(xhttp.NewTLSClientConfig()),
		proxied:   false,
		useIPv4:   true,
		useIPv6:   false,
		ttl:       300,
		ctx:       ctx,
		cancelCtx: cancel,
		closed:    make(chan struct{}),
	}

	for _, opt := range opts {
		opt(out)
	}

	if out.cfc == nil {
		return nil, errors.New("no Cloudflare client provided, use watcher.WithCloudflareClient option")
	}

	if out.zoneID == "" {
		return nil, errors.New("no target domain provided, use watcher.WithZoneID option")
	}

	if out.dnsName == "" {
		return nil, errors.New("no target domain provided, use watcher.WithDNSName option")
	}

	if !out.useIPv6 && !out.useIPv4 {
		return nil, errors.New("disabled both IPv4 and IPv6 proto, nothing to do with it")
	}

	return out, nil
}

func (w *Watcher) Watch() error {
	defer close(w.closed)

	log.Info().Msg("starts initial sync")
	if err := w.Sync(w.ctx); err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}

	log.Info().Msg("synced")
	ticker := time.NewTicker(time.Duration(w.ttl) * time.Second)
	for {
		select {
		case <-w.ctx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			log.Info().Msg("starts syncing")
			if err := w.Sync(w.ctx); err != nil {
				log.Error().Err(err).Msg("sync failed")
				continue
			}

			log.Info().Msg("synced")
		}
	}
}

func (w *Watcher) Sync(ctx context.Context) error {
	expectedRecs, err := w.expectedRecords(ctx)
	if err != nil {
		return fmt.Errorf("unable to build expected DNS records: %w", err)
	}
	log.Info().Array("records", loggableDNSRecords(expectedRecs)).Msg("collected expected records")

	actualRecs, err := w.actualRecords(ctx)
	if err != nil {
		return fmt.Errorf("unable to get actual DNS records: %w", err)
	}
	log.Info().Array("records", loggableDNSRecords(actualRecs)).Msg("collected actual records")

	updateSet := BuildUpdateSet(actualRecs, expectedRecs)
	if updateSet.IsEmpty() {
		log.Info().Msg("records are up to date, nothing to do")
		return nil
	}

	w.processDeletes(ctx, updateSet.ToDelete...)
	w.processUpdates(ctx, updateSet.ToUpdate...)
	w.processAdds(ctx, updateSet.ToAdd...)
	return nil
}

func (w *Watcher) processDeletes(ctx context.Context, recs ...DNSRecord) {
	for _, rr := range recs {
		if rr.ID == "" {
			log.Error().Object("rr", rr).Msg("unable to delete record w/o ID")
			continue
		}

		if err := w.cfc.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(w.zoneID), rr.ID); err != nil {
			log.Error().Object("rr", rr).Err(err).Msg("unable to delete record from CF")
			continue
		}

		log.Info().Object("rr", rr).Msg("record deleted")
	}
}

func (w *Watcher) processUpdates(ctx context.Context, recs ...DNSRecord) {
	for _, rr := range recs {
		if rr.ID == "" {
			log.Error().Object("rr", rr).Msg("unable to update record w/o ID")
			continue
		}

		record := cloudflare.UpdateDNSRecordParams{
			ID:       rr.ID,
			Name:     rr.Name,
			Type:     strings.ToUpper(rr.Type),
			Content:  rr.Content,
			TTL:      rr.TTL,
			Proxied:  rr.Proxied,
			Priority: rr.Priority,
		}
		err := w.cfc.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(w.zoneID), record)
		if err != nil {
			log.Error().Object("rr", rr).Err(err).Msg("unable to update record in CF")
			continue
		}

		log.Info().Object("rr", rr).Msg("record updated")
	}
}

func (w *Watcher) processAdds(ctx context.Context, recs ...DNSRecord) {
	for _, rr := range recs {
		record := cloudflare.CreateDNSRecordParams{
			Name:     rr.Name,
			Type:     strings.ToUpper(rr.Type),
			Content:  rr.Content,
			TTL:      rr.TTL,
			Proxied:  rr.Proxied,
			Priority: rr.Priority,
		}
		rsp, err := w.cfc.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(w.zoneID), record)
		if err != nil {
			log.Error().Object("rr", rr).Err(err).Msg("unable to update record in CF")
			continue
		}

		log.Info().Object("rr", DNSRecord{rsp.Result}).Msg("record added")
	}
}

func (w *Watcher) actualRecords(ctx context.Context) ([]DNSRecord, error) {
	cfRecs, _, err := w.cfc.ListDNSRecords(
		ctx,
		cloudflare.ZoneIdentifier(w.zoneID),
		cloudflare.ListDNSRecordsParams{Name: w.dnsName},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get actual DNS records: %w", err)
	}

	out := make([]DNSRecord, 0, len(cfRecs))
	for _, rr := range cfRecs {
		switch rr.Type {
		case "A", "AAAA":
		default:
			continue
		}

		out = append(out, DNSRecord{
			DNSRecord: rr,
		})
	}

	return out, nil
}

func (w *Watcher) expectedRecords(ctx context.Context) ([]DNSRecord, error) {
	var out []DNSRecord
	addIP := func(typ, ip string) {
		out = append(out, DNSRecord{
			cloudflare.DNSRecord{
				Type:    typ,
				Name:    w.dnsName,
				Content: ip,
				TTL:     w.ttl,
				Proxied: &w.proxied,
			},
		})
	}

	if w.useIPv4 {
		extIP, err := w.externalIP(ctx, "https://1.1.1.1/cdn-cgi/trace")
		if err != nil {
			return nil, fmt.Errorf("unable to get external IPv4: %w", err)
		}
		addIP("A", extIP)
	}

	if w.useIPv6 {
		extIP, err := w.externalIP(ctx, "https://[2606:4700:4700::1111]/cdn-cgi/trace")
		if err != nil {
			return nil, fmt.Errorf("unable to get external IPv6: %w", err)
		}
		addIP("AAAA", extIP)
	}

	return out, nil
}

func (w *Watcher) externalIP(ctx context.Context, traceURL string) (string, error) {
	rsp, err := w.httpc.R().
		SetContext(ctx).
		Get(traceURL)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	body := bytes.TrimSpace(rsp.Body())
	if rsp.IsError() {
		return "", fmt.Errorf("non-200 status code %q: %s", rsp.Status(), string(body))
	}

	out, err := xcf.IPFromTrace(body)
	if err != nil {
		return "", fmt.Errorf("unable to parse CF trace: %w", err)
	}

	return out, nil
}

func (w *Watcher) Shutdown(ctx context.Context) {
	w.cancelCtx()

	select {
	case <-ctx.Done():
	case <-w.closed:
	}
}
