package xcf_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/buglloc/cf-ddns/internal/xcf"
)

func TestIPFromTrace(t *testing.T) {
	cases := []struct {
		trace    string
		expected string
		err      bool
	}{
		{
			trace: `fl=87f397
h=[2606:4700:4700::1111]
ip=2a02:6b8:b081:7229::1:1c
ts=1666169316.225
visit_scheme=https
uag=curl/7.85.0
colo=DME
sliver=none
http=http/2
loc=RU
tls=TLSv1.3
sni=off
warp=off
gateway=off
kex=X25519`,
			expected: "2a02:6b8:b081:7229::1:1c",
		},
		{
			trace: `fl=87f524
h=1.1.1.1
ip=109.252.50.29
ts=1666169306.987
visit_scheme=https
uag=curl/7.85.0
colo=DME
sliver=none
http=http/2
loc=RU
tls=TLSv1.3
sni=off
warp=off
gateway=off
kex=X25519`,
			expected: "109.252.50.29",
		},
		{
			trace: `fl=87f524
h=1.1.1.1
ts=1666169306.987
visit_scheme=https
uag=curl/7.85.0
colo=DME
sliver=none
http=http/2
loc=RU
tls=TLSv1.3
sni=off
warp=off
gateway=off
kex=X25519`,
			expected: "invalid",
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			actual, err := xcf.IPFromTrace([]byte(tc.trace))
			if tc.err {
				require.Error(t, err)
				return
			}

			require.Equal(t, tc.expected, actual)
		})
	}
}
