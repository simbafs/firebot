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
