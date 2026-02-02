package parallel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUpdateMonitorWebhook_EnableSendsDetectedOnly(t *testing.T) {
	t.Parallel()

	var gotMethod, gotPath, gotAPIKey string
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("x-api-key")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"monitor_id":"m1","query":"q","status":"active","cadence":"daily","created_at":"2025-01-01T00:00:00Z","webhook":{"url":"https://x","event_types":["monitor.event.detected"]}}`)
	}))
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, APIKey: "k", Timeout: 5 * time.Second})
	_, err := c.UpdateMonitorWebhook(context.Background(), "m1", DetectedWebhook("https://x"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method=%s", gotMethod)
	}
	if gotPath != "/v1alpha/monitors/m1" {
		t.Fatalf("path=%s", gotPath)
	}
	if gotAPIKey != "k" {
		t.Fatalf("x-api-key=%q", gotAPIKey)
	}

	webhook, ok := gotBody["webhook"].(map[string]any)
	if !ok {
		t.Fatalf("webhook missing: %#v", gotBody)
	}
	if webhook["url"] != "https://x" {
		t.Fatalf("url=%v", webhook["url"])
	}
	ev, _ := webhook["event_types"].([]any)
	if len(ev) != 1 || ev[0] != "monitor.event.detected" {
		t.Fatalf("event_types=%v", webhook["event_types"])
	}
}

func TestUpdateMonitorWebhook_DisableSendsNullWebhook(t *testing.T) {
	t.Parallel()

	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"monitor_id":"m1","query":"q","status":"active","cadence":"daily","created_at":"2025-01-01T00:00:00Z","webhook":null}`)
	}))
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, APIKey: "k", Timeout: 5 * time.Second})
	_, err := c.UpdateMonitorWebhook(context.Background(), "m1", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if !strings.Contains(gotBody, `"webhook":null`) {
		t.Fatalf("body=%s", gotBody)
	}
}

func TestUpdateMonitor_EditOmitsWebhook(t *testing.T) {
	t.Parallel()

	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"monitor_id":"m1","query":"q2","status":"active","cadence":"weekly","created_at":"2025-01-01T00:00:00Z","webhook":null}`)
	}))
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, APIKey: "k", Timeout: 5 * time.Second})
	q := "q2"
	cad := "weekly"
	_, err := c.UpdateMonitor(context.Background(), "m1", UpdateMonitorRequest{Query: &q, Cadence: &cad})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if strings.Contains(gotBody, `"webhook"`) {
		t.Fatalf("body includes webhook: %s", gotBody)
	}
}
