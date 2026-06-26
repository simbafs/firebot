package fetch

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tainanfire/event"

	"github.com/PuerkitoBio/goquery"
)

// FetchChiayi parses the 嘉義縣 fire department page which uses a
// <table class="disaster_form"> with positional columns:
// 受理時間, 案類-細項, 案發地點, 派遣分隊, 案件狀態.
// Time uses ROC year (115 = 2026), format: "115-06-27 01:36:23".
// 案類-細項 is merged with a dash (e.g. "緊急救護-急病").
func (f *Fetcher) FetchChiayi(url, source string) (map[string]event.Event, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("failed to fetch data: " + res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	results := map[string]event.Event{}

	doc.Find("table.disaster_form tr").Not(":first-child").Each(func(i int, tr *goquery.Selection) {
		e := event.Event{}
		tr.Find("td").Each(func(j int, td *goquery.Selection) {
			content := strings.TrimSpace(td.Text())

			switch j {
			case 0: // 受理時間 — ROC year "115-06-27 01:36:23"
				e.Time = parseROC(content)

			case 1: // 案類-細項 — "緊急救護-急病" or "火災"
				cat, sub, _ := strings.Cut(content, "-")
				e.Category = strings.TrimSpace(cat)
				e.Subcategory = strings.TrimSpace(sub)

			case 2: // 案發地點
				e.Location = content

			case 3: // 派遣分隊
				e.Brigade = event.List(strings.Split(content, ","))

			case 4: // 案件狀態
				e.Status = content
			}
		})

		if e.Time.IsZero() {
			return
		}

		e.Source = source
		e.GenerateKey()

		if f.Filter(e) {
			results[e.Key] = e
		}
	})

	return results, nil
}

// parseROC parses a ROC year datetime like "115-06-27 01:36:23".
// ROC year = Western year - 1911, so 115 → 2026.
func parseROC(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}

	// Split "115-06-27" from "01:36:23"
	dateTime := strings.SplitN(s, " ", 2)
	if len(dateTime) != 2 {
		return time.Time{}
	}

	// Split "115-06-27"
	parts := strings.Split(dateTime[0], "-")
	if len(parts) != 3 {
		return time.Time{}
	}

	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])

	// Parse "01:36:23"
	timeParts := strings.Split(dateTime[1], ":")
	if len(timeParts) != 3 {
		return time.Time{}
	}

	hour, _ := strconv.Atoi(timeParts[0])
	min, _ := strconv.Atoi(timeParts[1])
	sec, _ := strconv.Atoi(timeParts[2])

	return time.Date(year+1911, time.Month(month), day, hour, min, sec, 0, time.Local)
}
