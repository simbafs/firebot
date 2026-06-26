package fetch

import (
	"encoding/json"
	"net/http"
	"strings"

	"tainanfire/event"
)

// ntpcResponse wraps the outer NTPC API envelope.
type ntpcResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    string `json:"data"` // double-escaped JSON string
}

// ntpcGeoJSON is the inner GeoJSON FeatureCollection.
type ntpcGeoJSON struct {
	Type     string        `json:"type"`
	Features []ntpcFeature `json:"features"`
}

type ntpcFeature struct {
	Properties ntpcProperties `json:"properties"`
}

type ntpcProperties struct {
	FeatureID    string          `json:"featureId"`
	FireType     string          `json:"fireType"`
	Type         string          `json:"type"`
	EndPointInfo string          `json:"endPointInfo"`
	CaseList     []ntpcCaseEntry `json:"caseList"`
}

type ntpcCaseEntry struct {
	StartPointInfo string `json:"startPointInfo"`
}

// FetchNTPCJSON parses the 新北 fire/rescue JSON API (GeoJSON FeatureCollection).
// Includes all event types (fire + ambulance). Key = "新北-{featureId}".
// Time is not available from the API; Status is empty.
func (f *Fetcher) FetchNTPCJSON(url, source string) (map[string]event.Event, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var envelope ntpcResponse
	if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
		return nil, err
	}

	var geojson ntpcGeoJSON
	if err := json.Unmarshal([]byte(envelope.Data), &geojson); err != nil {
		return nil, err
	}

	results := map[string]event.Event{}
	for _, feat := range geojson.Features {
		p := feat.Properties

		// Determine category: fireType for fire events, derived from type for others.
		category := strings.TrimSpace(p.FireType)
		if category == "" {
			switch p.Type {
			case "AmbulanceBack":
				category = "緊急救護"
			}
		}
		if category == "" {
			category = "未知"
		}

		// Collect brigade names from caseList.
		var brigades []string
		for _, entry := range p.CaseList {
			name := strings.TrimSpace(entry.StartPointInfo)
			if name != "" {
				brigades = append(brigades, name)
			}
		}

		e := event.Event{
			Category: category,
			Location: strings.TrimSpace(p.EndPointInfo),
			Brigade:  event.List(brigades),
			Source:   source,
		}

		if p.FeatureID != "" {
			e.ID = p.FeatureID
		}
		e.GenerateKey()

		if f.Filter(e) {
			results[e.Key] = e
		}
	}

	return results, nil
}
