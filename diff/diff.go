package diff

import (
	"sync"

	"tainanfire/event"
)

type EventDiff struct {
	Old     event.Event
	New     event.Event
	Changes []event.FieldChange
}

type DiffResult struct {
	New     []event.Event
	Updated []EventDiff
	Deleted []event.Event
}

type Differ struct {
	mu   sync.Mutex
	prev map[string]event.Event
}

func New() *Differ {
	return &Differ{prev: make(map[string]event.Event)}
}

func (d *Differ) Init(current map[string]event.Event) {
	d.mu.Lock()
	d.prev = make(map[string]event.Event, len(current))
	for k, v := range current {
		d.prev[k] = v
	}
	d.mu.Unlock()
}

func (d *Differ) Diff(current map[string]event.Event) DiffResult {
	d.mu.Lock()
	defer d.mu.Unlock()

	result := DiffResult{}

	for key, ev := range current {
		old, ok := d.prev[key]
		if !ok {
			result.New = append(result.New, ev)
			continue
		}
		changes := old.Changes(&ev)
		if len(changes) > 0 {
			result.Updated = append(result.Updated, EventDiff{
				Old:     old,
				New:     ev,
				Changes: changes,
			})
		}
	}

	for key, ev := range d.prev {
		if _, ok := current[key]; !ok {
			result.Deleted = append(result.Deleted, ev)
		}
	}

	d.prev = current
	return result
}
