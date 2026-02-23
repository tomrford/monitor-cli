# monitor-cli

Tiny Go CLI for Parallel Monitor API. Manage monitors + webhook delivery.

## Env

- `PARALLEL_API_KEY` (required)
- `PARALLEL_WEBHOOK_URL` (required for `enable`, optional for `add`)
- `PARALLEL_RELAY_TOKEN` (optional; auto-add `metadata.relay_token`)
- `PARALLEL_API_BASE` (optional, default `https://api.parallel.ai`)

Webhook event types: always `monitor.event.detected`.

## Usage

```bash
monitor list [--json]

monitor add --query "..." [--cadence daily|weekly|hourly] [--metadata-json '{}'] [--disabled] [--json]

monitor edit <monitor_id> [--query "..."] [--cadence daily|weekly|hourly] [--metadata-json '{}'] [--json]

monitor enable <monitor_id> [--json]
monitor disable <monitor_id> [--json]

monitor remove <monitor_id> [--json]
```

## Relay

Public webhook relay. Verifies Parallel signature, rate-limits, dedupes, filters, then forwards to tailnet target.

Env:

- `PARALLEL_WEBHOOK_SECRET` (required; from Parallel dashboard)
- `RELAY_FORWARD_URL` (required; mac mini handler URL)
- `RELAY_FORWARD_TOKEN` (optional; adds `Authorization: Bearer ...` to forwarded request)
- `RELAY_LISTEN_ADDR` (default `:8080`)
- `RELAY_METADATA_TOKEN` (optional; require payload `data.metadata.relay_token` match)
- `RELAY_ALLOW_MONITOR_IDS` (optional; comma allowlist)
- `RELAY_MAX_BODY_BYTES` (default `1048576`)
- `RELAY_REPLAY_WINDOW_SECONDS` (default `600`)
- `RELAY_RPS` (default `5`)
- `RELAY_BURST` (default `20`)
- `RELAY_TS_AUTHKEY` (optional; enables outbound tailnet dialing when set with `RELAY_TS_HOSTNAME`)
- `RELAY_TS_HOSTNAME` (optional; tsnet hostname for this relay node)
- `RELAY_TS_STATE_DIR` (optional; tsnet state path for persistent identity)
- `RELAY_TS_EPHEMERAL` (optional; default `true`)

Run:

```bash
relay
```

With Tailscale enabled, set `RELAY_FORWARD_URL` to your tailnet host, e.g. `http://macmini.tail91b66e.ts.net:18789/hooks/agent`.
