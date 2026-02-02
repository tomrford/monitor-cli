package relay

type MonitorWebhookEvent struct {
	Type      string                  `json:"type"`
	Timestamp string                  `json:"timestamp"`
	Data      MonitorWebhookEventData `json:"data"`
}

type MonitorWebhookEventData struct {
	MonitorID string            `json:"monitor_id"`
	Event     MonitorEvent      `json:"event"`
	Metadata  map[string]string `json:"metadata"`
}

type MonitorEvent struct {
	EventGroupID string `json:"event_group_id"`
}
