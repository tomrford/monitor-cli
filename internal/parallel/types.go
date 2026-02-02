package parallel

type MonitorWebhook struct {
	URL        string   `json:"url"`
	EventTypes []string `json:"event_types,omitempty"`
}

func DetectedWebhook(url string) *MonitorWebhook {
	return &MonitorWebhook{
		URL:        url,
		EventTypes: []string{"monitor.event.detected"},
	}
}

type MonitorResponse struct {
	MonitorID string            `json:"monitor_id"`
	Query     string            `json:"query"`
	Status    string            `json:"status"`
	Cadence   string            `json:"cadence"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Webhook   *MonitorWebhook   `json:"webhook"`
	CreatedAt string            `json:"created_at"`
	LastRunAt *string           `json:"last_run_at"`
}

type ListMonitorsResponse = []MonitorResponse

type CreateMonitorRequest struct {
	Query    string            `json:"query"`
	Cadence  string            `json:"cadence"`
	Webhook  *MonitorWebhook   `json:"webhook,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type UpdateMonitorRequest struct {
	Query    *string           `json:"query,omitempty"`
	Cadence  *string           `json:"cadence,omitempty"`
	Webhook  *MonitorWebhook   `json:"webhook,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type UpdateMonitorWebhookRequest struct {
	Webhook *MonitorWebhook `json:"webhook"`
}
