package event

import (
	"fmt"
	"time"
)

type Event struct {
	UID         string
	Key         string
	Source      string
	ID          string
	Time        time.Time
	Category    string
	Subcategory string
	Location    string
	Brigade     List
	Status      string
}

const TimeLayout = "2006/01/02 15:04:05"

// TWLoc is Asia/Taipei (UTC+8) — all event times come from Taiwan fire departments
// and snapshot times (time.Now()) should use this timezone regardless of server location.
var TWLoc *time.Location

func init() {
	var err error
	TWLoc, err = time.LoadLocation("Asia/Taipei")
	if err != nil {
		TWLoc = time.FixedZone("CST", 8*60*60)
	}
}

func (e *Event) GenerateKey() {
	if e.ID != "" {
		e.Key = fmt.Sprintf("%s-%s", e.Source, e.ID)
	} else {
		e.Key = fmt.Sprintf("%s-%s-%s-%s-%s", e.Source, e.Time.Format(TimeLayout), e.Category, e.Subcategory, e.Location)
	}
	e.UID = e.Key
}

func (e *Event) String() string {
	s := ""

	if len(e.Brigade) == 0 {
		s += fmt.Sprintf("`%s\n%s %s\n%s %s`", e.Time.Format(TimeLayout), e.Category, e.Subcategory, e.Location, e.Status)
	} else {
		s += fmt.Sprintf("`%s\n%s %s\n%s %s\n%s`", e.Time.Format(TimeLayout), e.Category, e.Subcategory, e.Location, e.Status, e.Brigade)
	}

	return s
}

func (e *Event) Diff(other *Event) string {
	s := ""
	if e.Time != other.Time {
		s += fmt.Sprintf("時間: %s -> %s\n", e.Time.Format(TimeLayout), other.Time.Format(TimeLayout))
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

type FieldChange struct {
	Field string
	Old   string
	New   string
}

func (e *Event) Changes(other *Event) []FieldChange {
	var changes []FieldChange
	if e.Time != other.Time {
		changes = append(changes, FieldChange{
			Field: "時間",
			Old:   e.Time.Format(TimeLayout),
			New:   other.Time.Format(TimeLayout),
		})
	}
	if e.Category != other.Category {
		changes = append(changes, FieldChange{
			Field: "類型",
			Old:   e.Category,
			New:   other.Category,
		})
	}
	if e.Subcategory != other.Subcategory {
		changes = append(changes, FieldChange{
			Field: "案別",
			Old:   e.Subcategory,
			New:   other.Subcategory,
		})
	}
	if e.Location != other.Location {
		changes = append(changes, FieldChange{
			Field: "地點",
			Old:   e.Location,
			New:   other.Location,
		})
	}
	if !e.Brigade.Equal(other.Brigade) {
		changes = append(changes, FieldChange{
			Field: "分隊",
			Old:   e.Brigade.String(),
			New:   other.Brigade.String(),
		})
	}
	if e.Status != other.Status {
		changes = append(changes, FieldChange{
			Field: "狀態",
			Old:   e.Status,
			New:   other.Status,
		})
	}
	return changes
}
