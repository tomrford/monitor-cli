package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	headerWebhookID        = "webhook-id"
	headerWebhookTimestamp = "webhook-timestamp"
	headerWebhookSignature = "webhook-signature"
)

func NewServer(cfg Config) http.Handler {
	mux := http.NewServeMux()

	dedupe := NewDedupe(1 * time.Hour)
	lim := NewTokenBucket(cfg.RPS, cfg.Burst)
	client := cfg.ForwardHTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for now := range t.C {
			dedupe.Prune(now)
		}
	}()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/parallel/webhook", func(w http.ResponseWriter, r *http.Request) {
		now := nowFn()
		if !lim.Allow(now) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ct := r.Header.Get("Content-Type")
		if ct != "" && !strings.HasPrefix(strings.ToLower(ct), "application/json") {
			http.Error(w, "unsupported content-type", http.StatusUnsupportedMediaType)
			return
		}

		body, err := readBody(r.Body, cfg.MaxBodyBytes)
		if err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		hdr := SignatureHeaders{
			ID:        r.Header.Get(headerWebhookID),
			Timestamp: r.Header.Get(headerWebhookTimestamp),
			Signature: r.Header.Get(headerWebhookSignature),
		}
		if err := VerifyParallelSignature(cfg.WebhookSecret, hdr, body, now, cfg.ReplayWindow); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if id := strings.TrimSpace(hdr.ID); id != "" {
			if dedupe.Seen(id, now) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("duplicate"))
				return
			}
		}

		var evt MonitorWebhookEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if evt.Type != "monitor.event.detected" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if cfg.AllowedMonitors != nil {
			if _, ok := cfg.AllowedMonitors[evt.Data.MonitorID]; !ok {
				http.Error(w, "monitor not allowed", http.StatusForbidden)
				return
			}
		}

		if cfg.MetadataToken != "" {
			if evt.Data.Metadata == nil || evt.Data.Metadata["relay_token"] != cfg.MetadataToken {
				http.Error(w, "metadata token mismatch", http.StatusUnauthorized)
				return
			}
		}

		outReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, cfg.ForwardURL, bytes.NewReader(body))
		if err != nil {
			http.Error(w, "forward build failed", http.StatusInternalServerError)
			return
		}
		outReq.Header.Set("Content-Type", "application/json")
		outReq.Header.Set("X-Parallel-Webhook-Id", hdr.ID)
		outReq.Header.Set("X-Parallel-Webhook-Timestamp", hdr.Timestamp)
		outReq.Header.Set("X-Parallel-Webhook-Signature", hdr.Signature)
		outReq.Header.Set("X-Parallel-Event-Type", evt.Type)
		outReq.Header.Set("X-Parallel-Monitor-Id", evt.Data.MonitorID)
		outReq.Header.Set("X-Parallel-Event-Group-Id", evt.Data.Event.EventGroupID)
		outReq.Header.Set("X-Relay-Received-At", now.UTC().Format(time.RFC3339Nano))
		outReq.Header.Set("X-Relay-Source-IP", clientIP(r))
		if cfg.ForwardToken != "" {
			outReq.Header.Set("Authorization", "Bearer "+cfg.ForwardToken)
		}

		resp, err := client.Do(outReq)
		if err != nil {
			log.Printf("forward error monitor_id=%s err=%s", evt.Data.MonitorID, err.Error())
			http.Error(w, "forward failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("forward non2xx monitor_id=%s status=%d", evt.Data.MonitorID, resp.StatusCode)
			http.Error(w, fmt.Sprintf("forward status %d", resp.StatusCode), http.StatusBadGateway)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return mux
}

func readBody(r io.Reader, max int64) ([]byte, error) {
	lr := io.LimitReader(r, max+1)
	b, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("body too large")
	}
	return b, nil
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
