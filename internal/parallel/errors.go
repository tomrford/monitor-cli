package parallel

import (
	"encoding/json"
	"fmt"
)

type apiErrorEnvelope struct {
	Type  string `json:"type"`
	Error struct {
		RefID   string `json:"ref_id"`
		Message string `json:"message"`
	} `json:"error"`
}

type validationErrorEnvelope struct {
	Detail any `json:"detail"`
}

func decodeAPIError(status int, body []byte) error {
	var e apiErrorEnvelope
	if err := json.Unmarshal(body, &e); err == nil && e.Error.Message != "" {
		if e.Error.RefID != "" {
			return fmt.Errorf("parallel api %d: %s (ref_id=%s)", status, e.Error.Message, e.Error.RefID)
		}
		return fmt.Errorf("parallel api %d: %s", status, e.Error.Message)
	}

	var v validationErrorEnvelope
	if err := json.Unmarshal(body, &v); err == nil && v.Detail != nil {
		return fmt.Errorf("parallel api %d: validation error: %s", status, string(body))
	}

	if len(body) > 0 {
		return fmt.Errorf("parallel api %d: %s", status, string(body))
	}
	return fmt.Errorf("parallel api %d", status)
}
