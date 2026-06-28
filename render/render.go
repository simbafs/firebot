package render

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"tainanfire/event"
)

type EventRow struct {
	Time       string
	Status     string
	Brigade    string
	CurBrigade string
}

func Heading(location, category string) string {
	if category != "" {
		return location + " - " + category
	}
	return location
}

func InitialRow(e *event.Event) EventRow {
	raw := e.Brigade.String()
	return EventRow{
		Time:       e.Time.Format("2006/01/02 15:04"),
		Status:     e.Status,
		Brigade:    raw,
		CurBrigade: raw,
	}
}

func SnapshotRow(status, prevBrigade, curBrigade string) EventRow {
	formatted, _ := diffBrigadeMarkdown(prevBrigade, curBrigade)
	return EventRow{
		Time:       time.Now().In(event.TWLoc).Format("2006/01/02 15:04"),
		Status:     status,
		Brigade:    formatted,
		CurBrigade: curBrigade,
	}
}

func diffBrigadeMarkdown(prev, cur string) (formatted, live string) {
	prevList := splitAndTrim(prev)
	curList := splitAndTrim(cur)

	curSet := make(map[string]bool, len(curList))
	for _, n := range curList {
		curSet[n] = true
	}

	var parts []string
	for _, p := range prevList {
		if curSet[p] {
			parts = append(parts, p)
		} else {
			parts = append(parts, "~~"+p+"~~")
		}
	}

	prevSet := make(map[string]bool, len(prevList))
	for _, p := range prevList {
		prevSet[p] = true
	}
	for _, c := range curList {
		if !prevSet[c] {
			parts = append(parts, "**"+c+"**")
		}
	}

	return strings.Join(parts, ", "), cur
}

func RenderRows(heading, activity string, rows []EventRow) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## %s\n\n", escapeRich(heading))

	if activity != "" {
		fmt.Fprintf(&b, "%s\n\n", activity)
	}

	b.WriteString("| 時間 | 狀態 | 分隊 |\n")
	b.WriteString("|:-----|:-----|:-----|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s |\n",
			r.Time,
			escapeRich(r.Status),
			escapeRich(r.Brigade),
		)
	}

	return b.String()
}

func ActivityLine(changes []event.FieldChange) string {
	var parts []string
	for _, c := range changes {
		switch c.Field {
		case "狀態":
			if c.Old == "" {
				parts = append(parts, fmt.Sprintf("狀態：%s", c.New))
			} else {
				parts = append(parts, fmt.Sprintf("狀態：%s → %s", c.Old, c.New))
			}
		case "分隊":
			added, removed := brigadeDiff(c.Old, c.New)
			for _, a := range added {
				parts = append(parts, fmt.Sprintf("新增 %s", a))
			}
			for _, r := range removed {
				parts = append(parts, fmt.Sprintf("移除 %s", r))
			}
		case "時間":
			parts = append(parts, fmt.Sprintf("時間：%s → %s", c.Old, c.New))
		case "地點":
			parts = append(parts, fmt.Sprintf("地點：%s → %s", c.Old, c.New))
		case "類型":
			parts = append(parts, fmt.Sprintf("類型：%s → %s", c.Old, c.New))
		default:
			if c.Old == "" {
				parts = append(parts, fmt.Sprintf("%s：%s", c.Field, c.New))
			} else {
				parts = append(parts, fmt.Sprintf("%s：%s → %s", c.Field, c.Old, c.New))
			}
		}
	}
	return strings.Join(parts, "\n")
}

func brigadeDiff(old, new string) (added, removed []string) {
	oldList := splitAndTrim(old)
	newList := splitAndTrim(new)

	for _, n := range newList {
		if !slices.Contains(oldList, n) {
			added = append(added, n)
		}
	}
	for _, o := range oldList {
		if !slices.Contains(newList, o) {
			removed = append(removed, o)
		}
	}
	return
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func escapeRich(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
