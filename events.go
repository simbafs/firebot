package main

import (
	"fmt"
	"time"
)

type Event struct {
	Key         string
	ID          string
	Time        time.Time
	Category    string
	Subcategory string
	Location    string
	Brigade     List
	Status      string
}

const timeLayout = "2006/01/02 15:04:05"

func (e *Event) String() string {
	s := ""

	if len(e.Brigade) == 0 {
		s += fmt.Sprintf("`%s\n%s %s\n%s %s`", e.Time.Format(timeLayout), e.Category, e.Subcategory, e.Location, e.Status)
	} else {
		s += fmt.Sprintf("`%s\n%s %s\n%s %s\n%s`", e.Time.Format(timeLayout), e.Category, e.Subcategory, e.Location, e.Status, e.Brigade)
	}

	// debug //
	// s += fmt.Sprintf("\n||---debug---\n%s||", e.ID)

	return s
}

func (e *Event) Diff(other *Event) string {
	s := ""
	if e.Time != other.Time {
		s += fmt.Sprintf("時間: %s -> %s\n", e.Time.Format(timeLayout), other.Time.Format(timeLayout))
	}
	if e.Category != other.Category {
		s += fmt.Sprintf("類型: %s -> %s\n", e.Category, other.Category)
	}
	if e.Subcategory != other.Subcategory {
		s += fmt.Sprintf("案別: %s -> %s\n", e.Subcategory, other.Subcategory)
	}
	if e.Location != other.Location {
		s += fmt.Sprintf("地點: %s -> %s\n", e.Location, other.Location)
	}
	if !e.Brigade.Equal(other.Brigade) {
		s += fmt.Sprintf("分隊：\n%s", e.Brigade.Diff(other.Brigade))
	}
	if e.Status != other.Status {
		s += fmt.Sprintf("狀態: %s -> %s\n", e.Status, other.Status)
	}
	return s
}
