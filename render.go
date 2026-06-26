package main

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type eventRow struct {
	Time    string // formatted time, e.g., "08:03:55"
	Status  string
	Brigade string
}

// heading builds the heading line for a rich message, e.g. "中西區民族路二段 - 火災"
func heading(location, category string) string {
	if category != "" {
		return location + " - " + category
	}
	return location
}

// initialRow creates the first table row from the event's receipt time.
func (e *Event) initialRow() eventRow {
	return eventRow{
		Time:    e.Time.Format("15:04:05"),
		Status:  e.Status,
		Brigade: e.Brigade.String(),
	}
}

// snapshotRow creates a new table row with the given status and brigade at the current time.
func snapshotRow(status, brigade string) eventRow {
	return eventRow{
		Time:    time.Now().Format("15:04"),
		Status:  status,
		Brigade: brigade,
	}
}

// renderRows builds a Rich Markdown message from the accumulated rows.
// activity is a change summary like "狀態：已出動 → 已到達" (may be empty).
func renderRows(heading, activity string, rows []eventRow) string {
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

// activityLine generates a concise summary of all field changes,
// e.g. "狀態：已出動 → 已到達、新增 仁德分隊"
func activityLine(changes []FieldChange) string {
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
	return strings.Join(parts, "、")
}

// brigadeDiff compares two comma-separated brigade lists and returns added and removed entries.
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

// escapeRich escapes characters that conflict with Rich Markdown pipe-table syntax.
func escapeRich(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
