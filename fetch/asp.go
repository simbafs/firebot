package fetch

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"tainanfire/event"

	"github.com/PuerkitoBio/goquery"
)

// FetchASP parses ASP-based pages that use <li>/<span data-th="..."> layouts.
// Used by 臺中 and 彰化.
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

	doc.Find("li:has(span[data-th])").Each(func(i int, s *goquery.Selection) {
		e := event.Event{}
		s.Find("span[data-th]").Each(func(i int, span *goquery.Selection) {
			th, _ := span.Attr("data-th")
			content := strings.TrimSpace(span.Text())

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
