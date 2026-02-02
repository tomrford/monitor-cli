# monitor-cli

Tiny Go CLI for Parallel Monitor API. Manage monitors + webhook delivery.

## Env

- `PARALLEL_API_KEY` (required)
- `PARALLEL_WEBHOOK_URL` (required for `enable`, optional for `add`)
- `PARALLEL_API_BASE` (optional, default `https://api.parallel.ai`)

Webhook event types: always `monitor.event.detected`.

## Usage

```bash
monitor list [--json]

monitor add --query "..." [--cadence daily|weekly|hourly|every_two_weeks] [--metadata-json '{}'] [--disabled] [--json]

monitor edit <monitor_id> [--query "..."] [--cadence daily|weekly|hourly|every_two_weeks] [--metadata-json '{}'] [--json]

monitor enable <monitor_id> [--json]
monitor disable <monitor_id> [--json]

monitor remove <monitor_id> [--json]
```
