# Agent Driver

`agent-driver` is a small JSONL client for the FluffyUI agent server. It can optionally launch a child command, feed it an input script, and wire up recording env vars.

## Usage

Run a Candy Wars demo (sim backend) and record an asciicast:

```bash
go run ./scripts/agent-driver \
  --addr unix:/tmp/fluffyui-candy.sock \
  --script scripts/agent-driver/scripts/candy-wars.jsonl \
  --backend sim --width 120 --height 36 \
  --record docs/demos/candy-wars.cast \
  --record-title "Candy Wars" \
  -- go run ./examples/candy-wars
```

## Script format

Each line is a JSON object. Blank lines and lines starting with `#` are ignored.

Common request types:
- `key`: `{"type":"key","key":"enter"}`
- `text`: `{"type":"text","text":"5"}`
- `mouse`: `{"type":"mouse","x":10,"y":5,"button":"left","action":"press"}`
- `snapshot`: `{"type":"snapshot"}`
- `sleep`: `{"type":"sleep","ms":250}`
- `wait_label`: `{"type":"wait_label","label":"Candy","timeout_ms":2000}`

Use `delay_ms` to pause after a request.
