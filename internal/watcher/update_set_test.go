package watcher_test

import (
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/stretchr/testify/require"

	"github.com/buglloc/cf-ddns/internal/watcher"
)

func TestBuildUpdateSet(t *testing.T) {
	cases := []struct {
		name            string
		actualRecords   []watcher.DNSRecord
		expectedRecords []watcher.DNSRecord
		expectedSet     watcher.UpdateSet
	}{
		{
			name: "from-empty-state",
			expectedRecords: []watcher.DNSRecord{
				{
					DNSRecord: cloudflare.DNSRecord{
						Type:    "A",
						Name:    "kek",
						Content: "1.1.1.1",
						TTL:     300,
					},
				},
				{
					DNSRecord: cloudflare.DNSRecord{
						Type:    "AAAA",
						Name:    "kek",
						Content: "2606:4700:4700::1111",
						TTL:     300,
					},
				},
			},
			expectedSet: watcher.UpdateSet{
				ToAdd: []watcher.DNSRecord{
					{
						DNSRecord: cloudflare.DNSRecord{
							Type:    "A",
							Name:    "kek",
							Content: "1.1.1.1",
							TTL:     300,
						},
					},
					{
						DNSRecord: cloudflare.DNSRecord{
							Type:    "AAAA",
							Name:    "kek",
							Content: "2606:4700:4700::1111",
							TTL:     300,
						},
					},
				},
			},
		},
		{
			name: "remove-stale-add-ipv6",
			actualRecords: []watcher.DNSRecord{
				{
					DNSRecord: cloudflare.DNSRecord{
						ID:      "1",
						Type:    "A",
						Name:    "kek",
						Content: "1.1.1.1",
						TTL:     300,
					},
				},
				{
					DNSRecord: cloudflare.DNSRecord{
						ID:      "2",
						Type:    "A",
						Name:    "kek",
						Content: "2.2.2.2",
						TTL:     300,
					},
				},
			},
			expectedRecords: []watcher.DNSRecord{
				{
					DNSRecord: cloudflare.DNSRecord{
						Type:    "A",
						Name:    "kek",
						Content: "1.1.1.1",
						TTL:     300,
					},
				},
				{
					DNSRecord: cloudflare.DNSRecord{
						Type:    "AAAA",
						Content: "2606:4700:4700::1111",
						TTL:     300,
					},
				},
			},
			expectedSet: watcher.UpdateSet{
				ToAdd: []watcher.DNSRecord{
					{
						DNSRecord: cloudflare.DNSRecord{
							Type:    "AAAA",
							Content: "2606:4700:4700::1111",
							TTL:     300,
						},
					},
				},
				ToDelete: []watcher.DNSRecord{
					{
						DNSRecord: cloudflare.DNSRecord{
							ID:      "2",
							Type:    "A",
							Name:    "kek",
							Content: "2.2.2.2",
							TTL:     300,
						},
					},
				},
			},
		},
		{
			name: "complex",
			actualRecords: []watcher.DNSRecord{
				{
					DNSRecord: cloudflare.DNSRecord{
						ID:      "1",
						Type:    "A",
						Name:    "kek",
						Content: "4.4.4.4",
						TTL:     300,
					},
				},
				{
					DNSRecord: cloudflare.DNSRecord{
						ID:      "2",
						Type:    "A",
						Name:    "kek",
						Content: "2.2.2.2",
						TTL:     300,
					},
				},
			},
			expectedRecords: []watcher.DNSRecord{
				{
					DNSRecord: cloudflare.DNSRecord{
						Type:    "A",
						Name:    "kek",
						Content: "1.1.1.1",
						TTL:     300,
					},
				},
				{
					DNSRecord: cloudflare.DNSRecord{
						Type:    "AAAA",
						Content: "2606:4700:4700::1111",
						TTL:     300,
					},
				},
			},
			expectedSet: watcher.UpdateSet{
				ToAdd: []watcher.DNSRecord{
					{
						DNSRecord: cloudflare.DNSRecord{
							Type:    "AAAA",
							Content: "2606:4700:4700::1111",
							TTL:     300,
						},
					},
				},
				ToUpdate: []watcher.DNSRecord{
					{
						DNSRecord: cloudflare.DNSRecord{
							ID:      "1",
							Type:    "A",
							Name:    "kek",
							Content: "1.1.1.1",
							TTL:     300,
						},
					},
				},
				ToDelete: []watcher.DNSRecord{
					{
						DNSRecord: cloudflare.DNSRecord{
							ID:      "2",
							Type:    "A",
							Name:    "kek",
							Content: "2.2.2.2",
							TTL:     300,
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := watcher.BuildUpdateSet(tc.actualRecords, tc.expectedRecords)
			require.Equal(t, tc.expectedSet, actual)
		})
	}
}
