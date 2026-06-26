# PLAN.md

## Roadmap: Multi-City Fire Department Dispatch Integration

Add support for 8 more cities/counties beyond the current 3 (臺南、高雄、新竹).

---

## Target Sources (11 total)

### Phase 1 — DTS System (zero code change, existing `fetch.go` works)

These use the exact same `<table id="dataTable">` structure as the current fetcher.

| City | URL | Has ID | Has 案別 | Env Prefix |
|------|-----|:------:|:--------:|------------|
| 苗栗 (Miaoli) | `https://119mlfire.mlfd.gov.tw/DTS/caselist/html` | ✗ | ✗ | `MIAOLI` |
| 雲林 (Yunlin) | `https://119.ylfire.gov.tw/DTS/caselist/html` | ✗ | ✓ | `YUNLIN` |

**Work:** Add two entries to `cityDefs` in `config.go`. No parser changes.

---

### Phase 2 — ASP Table (new parser, same column structure)

These use `<table>` but different CSS selectors than DTS. Column structure is the same (受理時間, 案類, 案別, 發生地點, 派遣分隊, 執行狀況).

| City | URL | Time Format | Env Prefix |
|------|-----|-------------|------------|
| 臺中 (Taichung) | `https://www.fire.taichung.gov.tw/caselist/index.asp?Parser=99,8,226` | `2006/01/02 15:04:05` | `TAICHUNG` |
| 彰化 (Changhua) | `https://www.chfd.gov.tw/RealInfo/index.aspx?Parser=99,3,29` | `2006/01/02 15:04:05` | `CHANGHUA` |

**Work:**
- Add `fetchASP()` method in `fetch/` package (generalized ASP/ASPX table parser)
- Both pages use standard `<tr><td>` table rows with column headers
- Key difference from DTS: different CSS selector for the table, different surrounding HTML

---

### Phase 3 — Custom HTML Parsers (one parser per city)

#### 3a. 桃園 (Taoyuan)
- **URL:** `https://www.tyfd.gov.tw/cht/index.php?act=caselist`
- **Format:** PHP CMS, div-based layout (not `<table>`)
- **Columns:** 發生時間, 案類, 案別, 發生地點, 派遣分隊, 案件狀態
- **Time format:** `2006-06-27 15:04` (dash-separated date, no seconds)
- **Env Prefix:** `TAOYUAN`

#### 3b. 新北 (New Taipei)
- **URL:** `https://e.ntpc.gov.tw/v3/api/map/dynamic/layer/rescue`
- **Format:** JSON API (GeoJSON FeatureCollection), no HTML scraping
- **Fields:** `fireType` (案類), `endPointInfo` (地點), `caseList[].startPointInfo` (分隊), `featureId` (unique ID)
- **Note:** Only shows fire events, no other categories. Time not directly available in JSON.
- **Env Prefix:** `NEWTAIPEI`

#### 3c. 嘉義縣 (Chiayi County)
- **URL:** `https://cycfb.cyhg.gov.tw/DisasterPrevent.aspx?n=5F10482409025004&sms=ED4E0CDDC2EA92E6`
- **Format:** ASPX custom layout
- **Columns:** 受理時間, 案類-細項 (merged with dash), 案發地點, 派遣分隊, 案件狀態
- **Time format:** `115-06-27 15:04:05` (ROC year — 115 = 2026)
- **Env Prefix:** `CHIAYI`

#### 3d. 屏東 (Pingtung)
- **URL:** `https://pteoc.pthg.gov.tw/News119`
- **Format:** ASPX custom layout, different column order
- **Columns:** 案類, 發生地點, 派遣分隊, 執行狀況, 受理時間 (reordered)
- **Time format:** `2026/06/27 15:04` (no seconds)
- **Env Prefix:** `PINGTUNG`

---

## Architecture Changes

### `config.go`

URLs are hardcoded; only chat IDs come from ENV.

```go
var cityDefs = []CityDef{
    {Source: "臺南",    Prefix: "TAINAN",    URL: "https://119dts.tncfd.gov.tw/DTS/caselist/html",                        Type: "dts"},
    {Source: "高雄",    Prefix: "KAOHSIUNG", URL: "https://119dts.fdkc.gov.tw/DTS/caselist/html",                        Type: "dts"},
    {Source: "新竹",    Prefix: "HSINCHU",   URL: "https://119.hcfd.gov.tw/DTS/caselist/html",                           Type: "dts"},
    {Source: "苗栗",    Prefix: "MIAOLI",    URL: "https://119mlfire.mlfd.gov.tw/DTS/caselist/html",                     Type: "dts"},
    {Source: "雲林",    Prefix: "YUNLIN",    URL: "https://119.ylfire.gov.tw/DTS/caselist/html",                         Type: "dts"},
    {Source: "臺中",    Prefix: "TAICHUNG",  URL: "https://www.fire.taichung.gov.tw/caselist/index.asp?Parser=99,8,226",  Type: "asp"},
    {Source: "彰化",    Prefix: "CHANGHUA",  URL: "https://www.chfd.gov.tw/RealInfo/index.aspx?Parser=99,3,29",          Type: "asp"},
    {Source: "桃園",    Prefix: "TAOYUAN",   URL: "https://www.tyfd.gov.tw/cht/index.php?act=caselist",                  Type: "taoyuan"},
    {Source: "新北",    Prefix: "NEWTAIPEI", URL: "https://e.ntpc.gov.tw/v3/api/map/dynamic/layer/rescue",                Type: "ntpc_json"},
    {Source: "嘉義縣",  Prefix: "CHIAYI",    URL: "https://cycfb.cyhg.gov.tw/DisasterPrevent.aspx?n=5F10482409025004&sms=ED4E0CDDC2EA92E6", Type: "chiayi"},
    {Source: "屏東",    Prefix: "PINGTUNG",  URL: "https://pteoc.pthg.gov.tw/News119",                                    Type: "pingtung"},
}
```

ENV: `{PREFIX}_CHAT` only (URLs are fixed in code). Optional `{PREFIX}_CHAT` to enable a city.

### `fetch/` package

```go
func (f *Fetcher) Fetch(url, source, fetcherType string) (map[string]event.Event, error) {
    switch fetcherType {
    case "dts":       return f.fetchDTS(url, source)
    case "asp":       return f.fetchASP(url, source)
    case "taoyuan":   return f.fetchTaoyuan(url, source)
    case "ntpc_json": return f.fetchNTPCJSON(url, source)
    case "chiayi":    return f.fetchChiayi(url, source)
    case "pingtung":  return f.fetchPingtung(url, source)
    }
}
```

### `main.go`

- `Fetcher.Fetch` signature gains a `fetcherType` parameter
- Pass `cityDef.Type` when calling `Fetcher.Fetch`

### compose.yaml

Add environment variable blocks for each new city (commented out by default):

```yaml
environment:
  # ... existing ...
  # - MIAOLI_CHAT=-100XXXX
  # - YUNLIN_CHAT=-100XXXX
```

---

## Implementation Order

1. **Phase 1 — 苗栗 + 雲林** (DTS, zero code change)
2. **Phase 2 — 臺中 + 彰化** (ASP parser, ~50 lines)
3. **Phase 3a — 桃園** (PHP div parser, ~60 lines)
4. **Phase 3b — 新北** (JSON parser, ~40 lines)
5. **Phase 3c — 嘉義** (ROC year converter + parser, ~50 lines)
6. **Phase 3d — 屏東** (reordered columns parser, ~50 lines)
