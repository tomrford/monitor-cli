package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"tailscale.com/tsnet"
)

const (
	envTailscaleAuthKey   = "RELAY_TS_AUTHKEY"
	envTailscaleHostname  = "RELAY_TS_HOSTNAME"
	envTailscaleStateDir  = "RELAY_TS_STATE_DIR"
	envTailscaleEphemeral = "RELAY_TS_EPHEMERAL"
)

type forwardClientSetup struct {
	Client           *http.Client
	Close            func() error
	TailscaleEnabled bool
}

func newForwardHTTPClientFromEnv() (forwardClientSetup, error) {
	setup := forwardClientSetup{
		Client: &http.Client{Timeout: 15 * time.Second},
	}

	authKey := strings.TrimSpace(os.Getenv(envTailscaleAuthKey))
	hostname := strings.TrimSpace(os.Getenv(envTailscaleHostname))
	stateDir := strings.TrimSpace(os.Getenv(envTailscaleStateDir))
	ephemeralRaw := strings.TrimSpace(os.Getenv(envTailscaleEphemeral))

	if authKey == "" && hostname == "" && stateDir == "" && ephemeralRaw == "" {
		return setup, nil
	}

	if authKey == "" || hostname == "" {
		return forwardClientSetup{}, fmt.Errorf("set both %s and %s to enable tailscale forwarding", envTailscaleAuthKey, envTailscaleHostname)
	}

	ephemeral := true
	if ephemeralRaw != "" {
		b, err := strconv.ParseBool(ephemeralRaw)
		if err != nil {
			return forwardClientSetup{}, fmt.Errorf("invalid env %s", envTailscaleEphemeral)
		}
		ephemeral = b
	}

	ts := &tsnet.Server{
		Hostname:     hostname,
		AuthKey:      authKey,
		Dir:          stateDir,
		Ephemeral:    ephemeral,
		RunWebClient: false,
	}
	if err := ts.Start(); err != nil {
		return forwardClientSetup{}, fmt.Errorf("tailscale start: %w", err)
	}

	dialer := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	setup.Client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if shouldDialViaTailscale(addr) {
					return ts.Dial(ctx, network, addr)
				}
				return dialer.DialContext(ctx, network, addr)
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	setup.Close = ts.Close
	setup.TailscaleEnabled = true
	return setup, nil
}

func shouldDialViaTailscale(addr string) bool {
	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	lowerHost := strings.ToLower(host)
	if strings.HasSuffix(lowerHost, ".ts.net") {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127
	}

	ip16 := ip.To16()
	if ip16 == nil {
		return false
	}

	return ip16[0] == 0xfd &&
		ip16[1] == 0x7a &&
		ip16[2] == 0x11 &&
		ip16[3] == 0x5c &&
		ip16[4] == 0xa1 &&
		ip16[5] == 0xe0
}
