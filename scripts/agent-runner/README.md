# Agent Runner

`agent-runner` connects to the FluffyUI agent socket and runs a policy loop:

1. Snapshot the accessibility tree
2. Decide on actions
3. Send inputs
4. Repeat

It can optionally launch a child command and set recording env vars.

## Usage

List available policies:

```bash
go run ./scripts/agent-runner --list-policies
```

Run Candy Wars with the demo policy and record an asciicast:

```bash
go run ./scripts/agent-runner \
  --addr unix:/tmp/fluffyui-candy.sock \
  --policy candy-wars-demo \
  --interval 300ms \
  --backend sim --width 120 --height 36 \
  --record docs/demos/candy-wars.cast \
  --record-title "Candy Wars" \
  -- go run ./examples/candy-wars
```

Write a JSONL log of snapshots, decisions, and actions:

```bash
go run ./scripts/agent-runner \
  --addr unix:/tmp/fluffyui-candy.sock \
  --policy candy-wars-demo \
  --log /tmp/candy-wars-agent-log.jsonl \
  -- go run ./examples/candy-wars
```

## Raw screen text (opt-in)

The runner uses accessibility snapshots by default. To request raw screen text, pass `--include-text` and ensure the server allows it (for example, set `FLUFFYUI_AGENT_ALLOW_TEXT=1` in the app).
