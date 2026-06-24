// Package progress provides reliable, byte-weighted progress reporting for
// long-running FoldersGuard operations.
//
// A Tracker is the single source of truth for one operation's progress. It
// records phases and byte/item totals, keeps reported progress monotonic and
// clamped, derives throughput and an estimated time remaining, and forwards
// throttled and coalesced events to a Sink. The WebUI builds a Tracker whose
// Sink emits application events; the CLI runs without a Tracker, in which case
// every Tracker method is a safe no-op.
package progress

import (
	"context"
	"sync"
	"time"
)

// State is the lifecycle state of an operation.
type State string

const (
	StatePending   State = "pending"
	StateRunning   State = "running"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

// Event is one progress snapshot delivered to a Sink. All fields are safe to
// serialize for the frontend; no sensitive values are included.
type Event struct {
	OperationID    string  `json:"operationId"`
	Operation      string  `json:"operation"`
	State          string  `json:"state"`
	Phase          string  `json:"phase"`
	PhaseIndex     int     `json:"phaseIndex"`
	PhaseCount     int     `json:"phaseCount"`
	Determinate    bool    `json:"determinate"`
	ProcessedBytes int64   `json:"processedBytes"`
	TotalBytes     int64   `json:"totalBytes"`
	ProcessedItems int     `json:"processedItems"`
	TotalItems     int     `json:"totalItems"`
	CurrentItem    string  `json:"currentItem"`
	BytesPerSecond float64 `json:"bytesPerSecond"`
	ETASeconds     float64 `json:"etaSeconds"`
	Error          string  `json:"error"`
}

// Stable phase keys shared by the Go core and the frontend, which localizes
// them for display.
const (
	PhasePreparing  = "preparing"
	PhaseEncrypting = "encrypting"
	PhaseDecrypting = "decrypting"
	PhaseVerifying  = "verifying"
	PhaseCopying    = "copying"
	PhaseFinalizing = "finalizing"
)

// Sink receives progress events. Implementations must be safe to call from the
// goroutine that drives the operation and must not block.
type Sink func(Event)

const (
	emitInterval  = 100 * time.Millisecond
	emitByteDelta = 8 << 20 // 8 MiB
)

// Tracker measures one operation's progress and emits throttled events.
//
// The zero value is not usable; build one with New. A nil *Tracker is valid and
// makes every method a no-op, so core code can hold an optional tracker without
// nil checks.
type Tracker struct {
	mu    sync.Mutex
	sink  Sink
	now   func() time.Time
	id    string
	op    string
	state State

	phases     []string
	phaseIndex int

	determinate    bool
	totalBytes     int64
	processedBytes int64
	totalItems     int
	processedItems int
	currentItem    string

	startTime       time.Time
	phaseStart      time.Time
	phaseStartBytes int64

	lastEmit      time.Time
	lastEmitBytes int64
	emitted       bool
}

// New creates a Tracker for one operation. The phases are the ordered phase
// labels the operation will move through; they may be empty when phases are not
// known ahead of time.
func New(id, operation string, sink Sink, phases ...string) *Tracker {
	now := time.Now()
	return &Tracker{
		sink:       sink,
		now:        time.Now,
		id:         id,
		op:         operation,
		state:      StatePending,
		phases:     phases,
		phaseIndex: -1,
		startTime:  now,
		phaseStart: now,
	}
}

// Begin marks the operation as running and emits an initial event.
func (t *Tracker) Begin() {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.state = StateRunning
	t.startTime = t.now()
	t.phaseStart = t.startTime
	t.mu.Unlock()
	t.emit(true)
}

// SetPhases declares the ordered phase labels for the operation. It may be
// called once after the tracker is built, before the first StartPhase.
func (t *Tracker) SetPhases(phases ...string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.phases = phases
	t.mu.Unlock()
}

// StartPhase moves to the named phase. A determinate phase reports a byte
// percentage; an indeterminate phase reports only activity and counts.
// Byte and throughput accounting restart for the new phase.
func (t *Tracker) StartPhase(name string, determinate bool) {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.phaseIndex = indexOf(t.phases, name, t.phaseIndex)
	t.determinate = determinate
	t.totalBytes = 0
	t.processedBytes = 0
	t.totalItems = 0
	t.processedItems = 0
	t.currentItem = ""
	t.phaseStart = t.now()
	t.phaseStartBytes = 0
	t.lastEmitBytes = 0
	t.mu.Unlock()
	t.emit(true)
}

// SetTotalBytes sets the byte total for the current phase.
func (t *Tracker) SetTotalBytes(total int64) {
	if t == nil {
		return
	}
	t.mu.Lock()
	if total < 0 {
		total = 0
	}
	t.totalBytes = total
	t.mu.Unlock()
	t.emit(true)
}

// SetTotalItems sets the item total for the current phase.
func (t *Tracker) SetTotalItems(total int) {
	if t == nil {
		return
	}
	t.mu.Lock()
	if total < 0 {
		total = 0
	}
	t.totalItems = total
	t.mu.Unlock()
	t.emit(false)
}

// AddBytes advances processed bytes by n. Progress is monotonic and clamped to
// the known total.
func (t *Tracker) AddBytes(n int64) {
	if t == nil || n <= 0 {
		return
	}
	t.mu.Lock()
	t.processedBytes += n
	if t.totalBytes > 0 && t.processedBytes > t.totalBytes {
		t.processedBytes = t.totalBytes
	}
	t.mu.Unlock()
	t.emit(false)
}

// SetItem records the name of the item currently being processed.
func (t *Tracker) SetItem(name string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.currentItem = name
	t.mu.Unlock()
	t.emit(false)
}

// ItemDone advances the processed item count by one.
func (t *Tracker) ItemDone() {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.processedItems++
	if t.totalItems > 0 && t.processedItems > t.totalItems {
		t.processedItems = t.totalItems
	}
	t.mu.Unlock()
	t.emit(false)
}

// Finish marks the operation as terminal and emits a final event. When err is
// non-nil the state is failed, unless cancelled is true, in which case the
// state is cancelled. A successful finish reports byte and item totals as fully
// processed.
func (t *Tracker) Finish(err error, cancelled bool) {
	if t == nil {
		return
	}
	t.mu.Lock()
	switch {
	case cancelled:
		t.state = StateCancelled
	case err != nil:
		t.state = StateFailed
	default:
		t.state = StateCompleted
		if t.totalBytes > 0 {
			t.processedBytes = t.totalBytes
		}
		if t.totalItems > 0 {
			t.processedItems = t.totalItems
		}
	}
	t.currentItem = ""
	t.mu.Unlock()
	t.emit(true)
}

// emit sends a snapshot to the sink, honoring throttling unless force is set.
func (t *Tracker) emit(force bool) {
	if t == nil || t.sink == nil {
		return
	}
	t.mu.Lock()
	now := t.now()
	if !force && t.emitted {
		sinceTime := now.Sub(t.lastEmit)
		sinceBytes := t.processedBytes - t.lastEmitBytes
		if sinceTime < emitInterval && sinceBytes < emitByteDelta {
			t.mu.Unlock()
			return
		}
	}
	t.lastEmit = now
	t.lastEmitBytes = t.processedBytes
	t.emitted = true
	event := t.snapshotLocked(now)
	sink := t.sink
	t.mu.Unlock()
	sink(event)
}

func (t *Tracker) snapshotLocked(now time.Time) Event {
	event := Event{
		OperationID:    t.id,
		Operation:      t.op,
		State:          string(t.state),
		PhaseCount:     len(t.phases),
		Determinate:    t.determinate,
		ProcessedBytes: t.processedBytes,
		TotalBytes:     t.totalBytes,
		ProcessedItems: t.processedItems,
		TotalItems:     t.totalItems,
		CurrentItem:    t.currentItem,
	}
	if t.phaseIndex >= 0 && t.phaseIndex < len(t.phases) {
		event.Phase = t.phases[t.phaseIndex]
		event.PhaseIndex = t.phaseIndex
	}
	if t.determinate {
		elapsed := now.Sub(t.phaseStart).Seconds()
		done := t.processedBytes - t.phaseStartBytes
		if elapsed > 0 && done > 0 {
			bps := float64(done) / elapsed
			event.BytesPerSecond = bps
			if bps > 0 && t.totalBytes > t.processedBytes {
				event.ETASeconds = float64(t.totalBytes-t.processedBytes) / bps
			}
		}
	}
	return event
}

func indexOf(phases []string, name string, fallback int) int {
	for i, phase := range phases {
		if phase == name {
			return i
		}
	}
	return fallback
}

type contextKey struct{}

// NewContext returns a copy of ctx carrying the tracker.
func NewContext(ctx context.Context, tracker *Tracker) context.Context {
	return context.WithValue(ctx, contextKey{}, tracker)
}

// FromContext returns the tracker carried by ctx, or nil when none is present.
// A nil tracker is safe to use: all methods are no-ops.
func FromContext(ctx context.Context) *Tracker {
	if ctx == nil {
		return nil
	}
	tracker, _ := ctx.Value(contextKey{}).(*Tracker)
	return tracker
}
