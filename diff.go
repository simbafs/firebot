package main

import "sync"

// EventDiff represents a single event that changed between fetches.
type EventDiff struct {
	Old     Event
	New     Event
	Changes []FieldChange
}

// DiffResult captures all changes between two consecutive fetches for a source.
// Callers can consume New, Updated, and Deleted independently — the Differ
// produces a stable, complete diff each cycle.
type DiffResult struct {
	New     []Event
	Updated []EventDiff
	Deleted []Event
}

// Differ compares consecutive fetch results for a single source.
// It maintains the previous snapshot internally and produces a DiffResult
// on each call to Diff. Safe for concurrent use.
type Differ struct {
	mu   sync.Mutex
	prev map[string]Event
}

func NewDiffer() *Differ {
	return &Differ{prev: make(map[string]Event)}
}

// Diff compares the current fetch result against the previous one.
// On the first call (prev is empty), everything is treated as new.
func (d *Differ) Diff(current map[string]Event) DiffResult {
	d.mu.Lock()
	defer d.mu.Unlock()

	result := DiffResult{}

	for key, event := range current {
		old, ok := d.prev[key]
		if !ok {
			result.New = append(result.New, event)
			continue
		}
		changes := old.Changes(&event)
		if len(changes) > 0 {
			result.Updated = append(result.Updated, EventDiff{
				Old:     old,
				New:     event,
				Changes: changes,
			})
		}
	}

	for key, event := range d.prev {
		if _, ok := current[key]; !ok {
			result.Deleted = append(result.Deleted, event)
		}
	}

	d.prev = current
	return result
}

// Init seeds the differ with a baseline snapshot without producing a diff.
// Use this on the very first fetch to avoid broadcasting everything as new.
func (d *Differ) Init(current map[string]Event) {
	d.mu.Lock()
	d.prev = make(map[string]Event, len(current))
	for k, v := range current {
		d.prev[k] = v
	}
	d.mu.Unlock()
}
