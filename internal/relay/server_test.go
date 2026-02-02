package relay

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRelay_VerifyAndForward(t *testing.T) {
	t.Parallel()

	secret := "s3cr3t"
	now := time.Unix(1738450000, 0)

	var gotBody []byte
	fwd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = b
		w.WriteHeader(http.StatusOK)
	}))
	defer fwd.Close()

	cfg := Config{
		ListenAddr:    ":0",
		ForwardURL:    fwd.URL,
		WebhookSecret: secret,
		MaxBodyBytes:  1 << 20,
		ReplayWindow:  10 * time.Minute,
		RPS:           100,
		Burst:         100,
		Now:           func() time.Time { return now },
	}
	h := NewServer(cfg)

	body := []byte(`{"type":"monitor.event.detected","timestamp":"2025-02-01T16:40:34Z","data":{"monitor_id":"m1","event":{"event_group_id":"eg1"},"metadata":{"relay_token":"x"}}}`)
	id := "msg_1"
	ts := "1738450000"
	sig := computeSignature(secret, id, ts, body)

	req := httptest.NewRequest(http.MethodPost, "/parallel/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerWebhookID, id)
	req.Header.Set(headerWebhookTimestamp, ts)
	req.Header.Set(headerWebhookSignature, "v1,"+sig)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if string(gotBody) != string(body) {
		t.Fatalf("forward body mismatch")
	}

	// duplicate id -> OK but no forward
	gotBody = nil
	req2 := httptest.NewRequest(http.MethodPost, "/parallel/webhook", bytes.NewReader(body))
	req2.Header.Set(headerWebhookID, id)
	req2.Header.Set(headerWebhookTimestamp, ts)
	req2.Header.Set(headerWebhookSignature, "v1,"+sig)

	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("dup status=%d", rr2.Code)
	}
	if gotBody != nil {
		t.Fatalf("expected no forward on duplicate")
	}

}

func TestVerifyParallelSignature_ReplayWindow(t *testing.T) {
	t.Parallel()

	secret := "s"
	body := []byte(`{"x":1}`)
	id := "m"
	now := time.Unix(1000, 0)
	ts := "1"
	sig := computeSignature(secret, id, ts, body)

	err := VerifyParallelSignature(secret, SignatureHeaders{
		ID:        id,
		Timestamp: ts,
		Signature: "v1," + sig,
	}, body, now, 10*time.Second)
	if err == nil {
		t.Fatalf("expected replay window err")
	}
}
