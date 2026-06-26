package fetch

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"tainanfire/event"

	"github.com/PuerkitoBio/goquery"
)

// FetchPingtung parses the 屏東 fire department page which uses a
// <table class="table news-table"> with positional columns in order:
// 案類, 發生地點, 派遣分隊, 執行狀況, 受理時間.
// Time format is "2006/06/27 15:04" (no seconds).
func (f *Fetcher) FetchPingtung(url, source string) (map[string]event.Event, error) {
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

	doc.Find("table.news-table tbody tr").Each(func(i int, tr *goquery.Selection) {
		e := event.Event{}
		tr.Find("td").Each(func(j int, td *goquery.Selection) {
			content := strings.TrimSpace(td.Text())

			switch j {
			case 0: // 案類
				e.Category = content
			case 1: // 發生地點
				e.Location = content
			case 2: // 派遣分隊
				e.Brigade = event.List(strings.Split(content, ","))
			case 3: // 執行狀況
				e.Status = content
			case 4: // 受理時間 — "2026/06/27 15:04"
				t, err := time.Parse("2006/01/02 15:04", content)
				if err == nil {
					e.Time = t
				}
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
