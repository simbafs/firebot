package fetch

import (
	"tainanfire/event"
)

type Fetcher struct {
	Filter func(event.Event) bool
}

// Fetch dispatches to the correct parser based on kind.
func (f *Fetcher) Fetch(url, source, kind string) (map[string]event.Event, error) {
	switch kind {
	case "dts":
		return f.FetchDTS(url, source)
	case "asp":
		return f.FetchASP(url, source)
	case "taoyuan":
		return f.FetchTaoyuan(url, source)
	default:
		return f.FetchDTS(url, source)
	}
}
