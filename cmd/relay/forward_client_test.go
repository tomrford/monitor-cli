package main

import "testing"

func TestShouldDialViaTailscale(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		addr string
		want bool
	}{
		{name: "tailnet hostname", addr: "macmini.tail91b66e.ts.net:18789", want: true},
		{name: "tailscale ipv4", addr: "100.101.102.103:443", want: true},
		{name: "tailscale ipv6", addr: "[fd7a:115c:a1e0::1]:443", want: true},
		{name: "railway internal", addr: "tailscale-forwarder.railway.internal:18789", want: false},
		{name: "public host", addr: "example.com:443", want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := shouldDialViaTailscale(tc.addr)
			if got != tc.want {
				t.Fatalf("addr=%s got=%t want=%t", tc.addr, got, tc.want)
			}
		})
	}
}

func TestNewForwardHTTPClientFromEnv_RequiresAuthAndHostname(t *testing.T) {
	t.Setenv(envTailscaleAuthKey, "tskey-abc")
	t.Setenv(envTailscaleHostname, "")

	_, err := newForwardHTTPClientFromEnv()
	if err == nil {
		t.Fatalf("expected error")
	}
}
