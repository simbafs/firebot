# AGENTS.md

## Project

`tainanfire` monitors Tainan/Kaohsiung/Hsinchu fire department incident pages, diffs events between fetches, and broadcasts changes to Telegram chats as Rich Messages (Bot API 10.1) with pinned, editable tables.

## Run

```bash
export API_KEY="your_telegram_bot_token"
go run .
```

Without `API_KEY`, a `LocalBot` prints to stdout instead of sending to Telegram.

Configuration lives in `config.yaml` (override path with `CONFIG_PATH` env var):
```yaml
chats:
  - source: 臺南
    chat_id: -1002309286627
    url: https://119dts.tncfd.gov.tw/DTS/caselist/html
```

## Architecture

```
config.yaml → Config → main loop (every 10s):
  per city: Fetcher.Fetch(url) → map[string]Event (goquery HTML scrape)
           → Differ.Diff(events) → DiffResult{New, Updated, Deleted}
           → Bot.Broadcast(chat, result, silent) → Telegram
```

**First fetch** calls `Differ.Init()` to seed the baseline without broadcasting. Subsequent fetches call `Differ.Diff()` and broadcast normally.

**Event identity** (`events.go:GenerateKey`):
- Tainan: `Key = "臺南-{編號}"` (has dedicated incident ID column)
- Kaohsiung: `Key = "{Source}-{Time}-{Category}-{Subcategory}-{Location}"` (no ID)
- Hsinchu: `Key = "{Source}-{Time}-{Category}-{Location}"` (no ID, no subcategory)

**Filter** (`main.go:filter`): only events with `Category == "火災"` OR `len(Brigade) >= 2` are kept.

## Rich Messages (Bot API 10.1)

`richmsg.go` makes raw HTTP POSTs to `sendRichMessage` and `editMessageText?rich_message=` because `gotgbot` v2.0.0-rc.33 does **not** support Bot API 10.1. Do not use gotgbot's `SendMessage` or `EditMessageText` for rich content — they lack the `rich_message` parameter.

When gotgbot adds 10.1 support, `richmsg.go` can be replaced with native calls.

## Message lifecycle

| State | Action |
|-------|--------|
| New | `sendRichMessage` + `PinChatMessage` (silent pin) |
| Updated | `editRichMessage` in-place (accumulates table rows); fallback to new message if edit fails |
| Closed | `editRichMessage` (final "已結案" row) → `UnpinChatMessage` |

Each event gets **one message**, edited in-place. Table rows accumulate — the first row uses the event receipt time, subsequent rows use `time.Now()` at change detection time. Active events are pinned (newest pin first). Startup calls `UnpinAllChatMessages` to clear leftover pins.

## Build / test

```bash
go build -o fire .
go test ./...       # one test in events_test.go
go vet ./...
```

## Gotchas

- **HTML column mapping in `fetch.go` is fragile**: it matches `<th>` text to field names. The three sites have different column sets (see table above). Changing the scraper must handle all three schemas.
- **gotgbot is pinned to rc.33** — do not upgrade casually; the rich message workaround depends on the gap between gotgbot's API level and Bot API 10.1.
- **`bucket/` package is unused** — kept for potential future use. Safe to ignore or remove.
- **No persistent state** — all data (differ prev map, msgIDs, eventRows) is in-memory. Restart clears everything.
- **Concurrency**: one Differ per city (each protected by its own mutex), goroutines per fetch iteration may overlap. TGBot maps are mutex-protected.
