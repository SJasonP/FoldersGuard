package progress

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNilTrackerIsNoOp(t *testing.T) {
	var tracker *Tracker
	// None of these must panic on a nil tracker.
	tracker.Begin()
	tracker.StartPhase("encrypting", true)
	tracker.SetTotalBytes(100)
	tracker.SetTotalItems(2)
	tracker.AddBytes(10)
	tracker.SetItem("file")
	tracker.ItemDone()
	tracker.Finish(nil, false)
}

func TestThrottleCoalescesUpdates(t *testing.T) {
	var events []Event
	clock := time.Unix(0, 0)
	tracker := New("op-1", "create", func(e Event) { events = append(events, e) })
	tracker.now = func() time.Time { return clock }

	tracker.Begin()                         // forced emit
	tracker.StartPhase("encrypting", true)  // forced emit
	tracker.SetTotalBytes(1 << 30)          // forced emit
	base := len(events)

	// Small, frequent updates inside one throttle window must coalesce.
	tracker.AddBytes(1024)
	tracker.AddBytes(1024)
	if len(events) != base {
		t.Fatalf("expected throttled updates to coalesce, got %d new events", len(events)-base)
	}

	// Advancing past the interval lets the next update through.
	clock = clock.Add(emitInterval + time.Millisecond)
	tracker.AddBytes(1024)
	if len(events) != base+1 {
		t.Fatalf("expected one emit after interval, got %d", len(events)-base)
	}

	// A large byte delta emits even within the interval.
	tracker.AddBytes(emitByteDelta)
	if len(events) != base+2 {
		t.Fatalf("expected emit after large byte delta, got %d", len(events)-base)
	}
}

func TestAddBytesIsMonotonicAndClamped(t *testing.T) {
	var last Event
	clock := time.Unix(0, 0)
	tracker := New("op-2", "verify", func(e Event) { last = e })
	tracker.now = func() time.Time { return clock }

	tracker.Begin()
	tracker.StartPhase("verifying", true)
	tracker.SetTotalBytes(100)
	tracker.AddBytes(60)
	clock = clock.Add(time.Second)
	tracker.AddBytes(60) // would exceed total

	if last.ProcessedBytes != 100 {
		t.Fatalf("processed bytes = %d, want clamped to 100", last.ProcessedBytes)
	}
}

func TestFinishCompletedFillsTotals(t *testing.T) {
	var last Event
	tracker := New("op-3", "create", func(e Event) { last = e })
	tracker.Begin()
	tracker.StartPhase("encrypting", true)
	tracker.SetTotalBytes(500)
	tracker.SetTotalItems(3)
	tracker.AddBytes(120)

	tracker.Finish(nil, false)
	if last.State != string(StateCompleted) {
		t.Fatalf("state = %q, want completed", last.State)
	}
	if last.ProcessedBytes != 500 || last.ProcessedItems != 3 {
		t.Fatalf("completed totals = %d bytes / %d items, want 500/3", last.ProcessedBytes, last.ProcessedItems)
	}
}

func TestFinishStates(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		cancelled bool
		want      State
	}{
		{"failure", errors.New("boom"), false, StateFailed},
		{"cancel", context.Canceled, true, StateCancelled},
		{"cancel wins over error", errors.New("boom"), true, StateCancelled},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var last Event
			tracker := New("op", "verify", func(e Event) { last = e })
			tracker.Begin()
			tracker.Finish(tc.err, tc.cancelled)
			if last.State != string(tc.want) {
				t.Fatalf("state = %q, want %q", last.State, tc.want)
			}
		})
	}
}

func TestContextRoundTrip(t *testing.T) {
	if FromContext(context.Background()) != nil {
		t.Fatal("expected nil tracker from empty context")
	}
	tracker := New("op", "create", nil)
	ctx := NewContext(context.Background(), tracker)
	if FromContext(ctx) != tracker {
		t.Fatal("expected tracker round-trip through context")
	}
}
