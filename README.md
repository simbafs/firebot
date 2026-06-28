# tainanfire

監控多縣市消防局案件頁面，比對前後差異，透過 Telegram Rich Messages (Bot API 10.1) 廣播變更。

## 執行

```bash
export API_KEY="your_telegram_bot_token"
go run .
```

無 `API_KEY` 時，`LocalBot` 會輸出到 stdout。透過環境變數 `{PREFIX}_CHAT` 啟用各縣市，例如 `TAINAN_CHAT=-123456`；或設定 `ALL_CHAT` 讓所有縣市共用同一個聊天室。

## 支援縣市

| 縣市   |   時間   | 案類 | 案別 | 發生地點 | 派遣分隊 | 執行狀況 | 編號 | 格式     |
| ------ | :------: | :--: | :--: | :------: | :------: | :------: | :--: | -------- |
| 臺南   |    O     |  O   |  O   |    O     |    O     |    O     |  O   | DTS HTML |
| 高雄   |    O     |  O   |  O   |    O     |    O     |    O     |  -   | DTS HTML |
| 新竹   |    O     |  O   |  O   |    O     |    O     |    O     |  -   | DTS HTML |
| 苗栗   |    O     |  O   |  O   |    O     |    O     |    O     |  -   | DTS HTML |
| 雲林   |    O     |  O   |  O   |    O     |    O     |    O     |  -   | DTS HTML |
| 臺中   |    O     |  O   |  O   |    O     |    O     |    O     |  -   | ASP HTML |
| 彰化   |    O     |  O   |  O   |    O     |    O     |    O     |  -   | ASP HTML |
| 桃園   |    O     |  O   |  O   |    O     |    O     |    O     |  -   | 桃園專用 |
| 新北   |    -     |  O   |  -   |    O     |    O     |    -     |  O   | JSON API |
| 嘉義縣 | O (民國) |  O   |  O   |    O     |    O     |    O     |  -   | 嘉義專用 |
| 屏東   |    O     |  O   |  -   |    O     |    O     |    O     |  -   | 屏東專用 |

- **O** 表示該欄位存在，**-** 表示該來源不提供此欄位
- 嘉義縣使用民國年格式（如 `115-06-27`），會自動轉換為西元年
- 新北的時間欄位來自 JSON API 不提供

## License

MIT
