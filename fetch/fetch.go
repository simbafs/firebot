package fetch

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"tainanfire/event"

	"github.com/PuerkitoBio/goquery"
)

type Fetcher struct {
	Filter func(event.Event) bool
}

func (f *Fetcher) Fetch(url, source string) (map[string]event.Event, error) {
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

	titles := doc.Find("tbody tr:first-child th").Map(func(i int, s *goquery.Selection) string {
		return strings.Trim(s.Text(), " \n\t")
	})
	results := map[string]event.Event{}

	doc.Find("tbody tr").Not(":first-child").Each(func(i int, s *goquery.Selection) {
		e := event.Event{}
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			content := strings.Trim(s.Text(), " \n\t")
			switch titles[i] {
			case "受理時間":
				t, err := time.Parse("2006/01/02 15:04:05", content)
				if err == nil {
					t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
				}
				e.Time = t
			case "派遣分隊":
				e.Brigade = strings.Split(content, ",")
			case "編號":
				e.ID = content
			case "案類":
				e.Category = content
			case "案別":
				e.Subcategory = content
			case "發生地點":
				e.Location = content
			case "執行狀況":
				e.Status = content
			}
		})

		e.Source = source
		e.GenerateKey()

		if f.Filter(e) {
			results[e.Key] = e
		}
	})

	return results, nil
}

// FetchASP parses ASP-based pages that use <li>/<span data-th="..."> layouts
// instead of <table>. Used by 臺中 and 彰化.
func (f *Fetcher) FetchASP(url, source string) (map[string]event.Event, error) {
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

	// Each row is a <li> containing <span data-th="..."> elements.
	// Find all such <li> containers and extract data by data-th attribute.
	doc.Find("li:has(span[data-th])").Each(func(i int, s *goquery.Selection) {
		e := event.Event{}
		s.Find("span[data-th]").Each(func(i int, span *goquery.Selection) {
			th, _ := span.Attr("data-th")
			content := strings.TrimSpace(span.Text())

			// The label may have a trailing "：" (全形 colon) or not.
			th = strings.TrimRight(th, "：")

			switch th {
			case "受理時間":
				t, err := time.Parse("2006/01/02 15:04:05", content)
				if err == nil {
					t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
				}
				e.Time = t
			case "派遣分隊":
				e.Brigade = event.List(strings.Split(content, ","))
			case "案類":
				e.Category = content
			case "案別":
				e.Subcategory = content
			case "發生地點":
				e.Location = content
			case "執行狀況":
				e.Status = content
			}
		})

	// Skip header rows (no time parsed)
		if e.Time.IsZero() {
			return
		}

		e.Source = source
		e.GenerateKey()

		if f.Filter(e) {
			results[e.Key] = e
		}
	})
	// End FetchASP.

	return results, nil
}

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

			// Strip the <span class="title"> prefix (e.g. "案類火災" → "火災").
			td.Find("span.title").Each(func(k int, title *goquery.Selection) {
				prefix := strings.TrimSpace(title.Text())
				content = strings.TrimPrefix(content, prefix)
			})
			content = strings.TrimSpace(content)

			switch j {
			case 0: // 發生時間 — "2006-06-27 15:04"
				t, err := time.Parse("2006-01-02 15:04", content)
				if err == nil {
					e.Time = t
				}
			case 1: // 案類
				e.Category = content
			case 2: // 案別
				e.Subcategory = content
			case 3: // 發生地點
				e.Location = content
			case 4: // 派遣分隊
				e.Brigade = event.List(strings.Split(content, ","))
			case 5: // 案件狀態
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
