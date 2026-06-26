package main

import (
	"fmt"
	"slices"
	"strings"
)

// RichMarkdown renders a new event as a Rich Markdown table.
// Location is the heading, followed by activity line and a single-row table.
func (e *Event) RichMarkdown() string {
	var b strings.Builder

	heading := e.Location
	if e.Category != "" {
		heading += " - " + e.Category
	}
	fmt.Fprintf(&b, "## %s\n\n", escapeRich(heading))

	fmt.Fprintf(&b, "🆕 新事件\n\n")

	b.WriteString("| 時間 | 狀態 | 分隊 |\n")
	b.WriteString("|:-----|:-----|:-----|\n")
	fmt.Fprintf(&b, "| %s | %s | %s |\n",
		e.Time.Format("15:04:05"),
		escapeRich(e.Status),
		escapeRich(e.Brigade.String()),
	)

	return b.String()
}

// RichDiffMarkdown renders an updated event as a Rich Markdown table with
// a change summary line between the heading and the table.
func (ed *EventDiff) RichDiffMarkdown() string {
	var b strings.Builder

	heading := ed.New.Location
	if ed.New.Category != "" {
		heading += " - " + ed.New.Category
	}
	fmt.Fprintf(&b, "## %s\n\n", escapeRich(heading))

	line := activityLine(ed.Changes)
	if line != "" {
		fmt.Fprintf(&b, "%s\n\n", line)
	}

	b.WriteString("| 時間 | 狀態 | 分隊 |\n")
	b.WriteString("|:-----|:-----|:-----|\n")
	fmt.Fprintf(&b, "| %s | %s | %s |\n",
		ed.New.Time.Format("15:04:05"),
		escapeRich(ed.New.Status),
		escapeRich(ed.New.Brigade.String()),
	)

	return b.String()
}

// activityLine generates a concise summary of all field changes,
// e.g. "狀態更新：已出動 → 已到達、新增 仁德分隊"
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
	// Pipe and newline break the table structure.
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
