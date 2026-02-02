package relay

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SignatureHeaders struct {
	ID        string
	Timestamp string
	Signature string
}

func VerifyParallelSignature(secret string, hdr SignatureHeaders, body []byte, now time.Time, replayWindow time.Duration) error {
	id := strings.TrimSpace(hdr.ID)
	ts := strings.TrimSpace(hdr.Timestamp)
	sigHeader := strings.TrimSpace(hdr.Signature)
	if id == "" || ts == "" || sigHeader == "" {
		return fmt.Errorf("missing signature headers")
	}

	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook-timestamp")
	}
	t := time.Unix(sec, 0)
	if replayWindow > 0 {
		if now.Sub(t) > replayWindow || t.Sub(now) > replayWindow {
			return fmt.Errorf("timestamp outside replay window")
		}
	}

	want := computeSignature(secret, id, ts, body)

	// header format: "v1,<sig> v1,<sig2> ..."
	parts := strings.Fields(sigHeader)
	for _, p := range parts {
		v, got, ok := strings.Cut(p, ",")
		if !ok {
			continue
		}
		if v != "v1" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1 {
			return nil
		}
	}

	return fmt.Errorf("signature mismatch")
}

func computeSignature(secret, id, ts string, body []byte) string {
	msg := id + "." + ts + "." + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	sum := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(sum)
}
