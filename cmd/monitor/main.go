package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tomrford/monitor-cli/internal/parallel"
)

const (
	envAPIKey     = "PARALLEL_API_KEY"
	envWebhookURL = "PARALLEL_WEBHOOK_URL"
	envAPIBase    = "PARALLEL_API_BASE"
	envRelayToken = "PARALLEL_RELAY_TOKEN"

	defaultAPIBase = "https://api.parallel.ai"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return usageError("")
	}

	cmd := args[1]
	switch cmd {
	case "help", "-h", "--help":
		return usageError("")
	case "list":
		return cmdList(args[2:])
	case "add":
		return cmdAdd(args[2:])
	case "edit":
		return cmdEdit(args[2:])
	case "enable":
		return cmdEnable(args[2:])
	case "disable":
		return cmdDisable(args[2:])
	case "remove":
		return cmdRemove(args[2:])
	default:
		return usageError("unknown command: " + cmd)
	}
}

func usageError(extra string) error {
	var b strings.Builder
	if extra != "" {
		b.WriteString(extra)
		b.WriteString("\n\n")
	}
	b.WriteString("usage:\n")
	b.WriteString("  monitor list [--json]\n")
	b.WriteString("  monitor add --query \"...\" [--cadence daily|weekly|hourly] [--metadata-json '{}'] [--disabled] [--json]\n")
	b.WriteString("  monitor edit <monitor_id> [--query \"...\"] [--cadence daily|weekly|hourly] [--metadata-json '{}'] [--json]\n")
	b.WriteString("  monitor enable <monitor_id> [--json]\n")
	b.WriteString("  monitor disable <monitor_id> [--json]\n")
	b.WriteString("  monitor remove <monitor_id> [--json]\n\n")
	b.WriteString("env:\n")
	b.WriteString("  " + envAPIKey + " (required)\n")
	b.WriteString("  " + envWebhookURL + " (required for enable; optional for add)\n")
	b.WriteString("  " + envAPIBase + " (optional; default " + defaultAPIBase + ")\n")
	b.WriteString("  " + envRelayToken + " (optional; auto-add metadata.relay_token)\n")
	return errors.New(b.String())
}

func mustClient() (*parallel.Client, error) {
	apiKey := strings.TrimSpace(os.Getenv(envAPIKey))
	if apiKey == "" {
		return nil, fmt.Errorf("missing env %s", envAPIKey)
	}

	baseURL := strings.TrimSpace(os.Getenv(envAPIBase))
	if baseURL == "" {
		baseURL = defaultAPIBase
	}

	return parallel.NewClient(parallel.Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
	}), nil
}

func cmdList(argv []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(argv); err != nil {
		return usageError(err.Error())
	}

	client, err := mustClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	resp, err := client.ListMonitors(ctx)
	if err != nil {
		return err
	}

	if *jsonOut {
		return writeJSON(os.Stdout, resp)
	}

	for _, m := range *resp {
		url := ""
		enabled := false
		if m.Webhook != nil {
			enabled = true
			url = m.Webhook.URL
		}
		fmt.Printf("%s\t%s\t%s\tenabled=%t\t%s\n", m.MonitorID, m.Status, m.Cadence, enabled, url)
	}
	return nil
}

func cmdAdd(argv []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	query := fs.String("query", "", "")
	cadence := fs.String("cadence", "daily", "")
	metadataJSON := fs.String("metadata-json", "", "")
	disabled := fs.Bool("disabled", false, "")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(argv); err != nil {
		return usageError(err.Error())
	}
	if strings.TrimSpace(*query) == "" {
		return usageError("missing --query")
	}
	normalizedCadence, err := normalizeCadence(*cadence)
	if err != nil {
		return usageError(err.Error())
	}

	var metadata map[string]string
	if strings.TrimSpace(*metadataJSON) != "" {
		if err := json.Unmarshal([]byte(*metadataJSON), &metadata); err != nil {
			return fmt.Errorf("invalid --metadata-json: %w", err)
		}
	}
	metadata = withRelayToken(metadata, os.Getenv(envRelayToken))

	var webhook *parallel.MonitorWebhook
	if !*disabled {
		webhookURL := strings.TrimSpace(os.Getenv(envWebhookURL))
		if webhookURL == "" {
			return fmt.Errorf("missing env %s (or pass --disabled)", envWebhookURL)
		}
		webhook = parallel.DetectedWebhook(webhookURL)
	}

	client, err := mustClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	created, err := client.CreateMonitor(ctx, parallel.CreateMonitorRequest{
		Query:    *query,
		Cadence:  normalizedCadence,
		Webhook:  webhook,
		Metadata: metadata,
	})
	if err != nil {
		return err
	}

	if *jsonOut {
		return writeJSON(os.Stdout, created)
	}
	fmt.Println(created.MonitorID)
	return nil
}

func cmdEnable(argv []string) error {
	fs := flag.NewFlagSet("enable", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(argv); err != nil {
		return usageError(err.Error())
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return usageError("missing <monitor_id>")
	}

	webhookURL := strings.TrimSpace(os.Getenv(envWebhookURL))
	if webhookURL == "" {
		return fmt.Errorf("missing env %s", envWebhookURL)
	}

	client, err := mustClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	updated, err := client.UpdateMonitorWebhook(ctx, rest[0], parallel.DetectedWebhook(webhookURL))
	if err != nil {
		return err
	}

	if *jsonOut {
		return writeJSON(os.Stdout, updated)
	}
	fmt.Printf("%s\tenabled=true\t%s\n", updated.MonitorID, webhookURL)
	return nil
}

func cmdEdit(argv []string) error {
	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	query := fs.String("query", "", "")
	cadence := fs.String("cadence", "", "")
	metadataJSON := fs.String("metadata-json", "", "")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(argv); err != nil {
		return usageError(err.Error())
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return usageError("missing <monitor_id>")
	}

	var req parallel.UpdateMonitorRequest
	changed := false

	if strings.TrimSpace(*query) != "" {
		q := *query
		req.Query = &q
		changed = true
	}
	if strings.TrimSpace(*cadence) != "" {
		normalizedCadence, err := normalizeCadence(*cadence)
		if err != nil {
			return usageError(err.Error())
		}
		req.Cadence = &normalizedCadence
		changed = true
	}
	if strings.TrimSpace(*metadataJSON) != "" {
		var metadata map[string]string
		if err := json.Unmarshal([]byte(*metadataJSON), &metadata); err != nil {
			return fmt.Errorf("invalid --metadata-json: %w", err)
		}
		req.Metadata = withRelayToken(metadata, os.Getenv(envRelayToken))
		changed = true
	}

	if !changed {
		return usageError("no changes (provide --query and/or --cadence and/or --metadata-json)")
	}

	client, err := mustClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	updated, err := client.UpdateMonitor(ctx, rest[0], req)
	if err != nil {
		return err
	}

	if *jsonOut {
		return writeJSON(os.Stdout, updated)
	}
	fmt.Printf("%s\tupdated\n", updated.MonitorID)
	return nil
}

func withRelayToken(metadata map[string]string, token string) map[string]string {
	t := strings.TrimSpace(token)
	if t == "" {
		return metadata
	}
	if metadata == nil {
		metadata = map[string]string{}
	}
	if _, ok := metadata["relay_token"]; !ok {
		metadata["relay_token"] = t
	}
	return metadata
}

func normalizeCadence(cadence string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(cadence))
	switch normalized {
	case "hourly", "daily", "weekly":
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid --cadence (allowed: hourly, daily, weekly)")
	}
}

func cmdDisable(argv []string) error {
	fs := flag.NewFlagSet("disable", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(argv); err != nil {
		return usageError(err.Error())
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return usageError("missing <monitor_id>")
	}

	client, err := mustClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	updated, err := client.UpdateMonitorWebhook(ctx, rest[0], nil)
	if err != nil {
		return err
	}

	if *jsonOut {
		return writeJSON(os.Stdout, updated)
	}
	fmt.Printf("%s\tenabled=false\n", updated.MonitorID)
	return nil
}

func cmdRemove(argv []string) error {
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(argv); err != nil {
		return usageError(err.Error())
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return usageError("missing <monitor_id>")
	}

	client, err := mustClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	deleted, err := client.DeleteMonitor(ctx, rest[0])
	if err != nil {
		return err
	}

	if *jsonOut {
		return writeJSON(os.Stdout, deleted)
	}
	fmt.Printf("%s\tdeleted\n", deleted.MonitorID)
	return nil
}

func writeJSON(w *os.File, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
