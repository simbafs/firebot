package fetch

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"tainanfire/event"

	"github.com/PuerkitoBio/goquery"
)

// FetchTaoyuan parses the 桃園 fire department page which uses a div-based
// table layout: .table-content.table-classic .table > .tr > .td.
// Columns are positional: 發生時間, 案類, 案別, 發生地點, 派遣分隊, 案件狀態.
// Time format is "2006-06-27 15:04" (dash-separated date, no seconds).
func (f *Fetcher) FetchTaoyuan(url, source string) (map[string]event.Event, error) {
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

	doc.Find(".table-content.table-classic .table .tr").Not(":first-child").Each(func(i int, tr *goquery.Selection) {
		e := event.Event{}
		tr.Find(".td").Each(func(j int, td *goquery.Selection) {
			content := strings.TrimSpace(td.Text())

			td.Find("span.title").Each(func(k int, title *goquery.Selection) {
				prefix := strings.TrimSpace(title.Text())
				content = strings.TrimPrefix(content, prefix)
			})
			content = strings.TrimSpace(content)

			switch j {
			case 0:
				t, err := time.Parse("2006-01-02 15:04", content)
				if err == nil {
					e.Time = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, event.TWLoc)
				}
			case 1:
				e.Category = content
			case 2:
				e.Subcategory = content
			case 3:
				e.Location = content
			case 4:
				e.Brigade = event.List(strings.Split(content, ","))
			case 5:
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
