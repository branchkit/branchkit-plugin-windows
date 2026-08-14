package main

import "testing"

// The mouse latch must release exactly once, and the release must not move to
// function exit.
//
// handleMoveToSpace presses the left button, drags to latch a window onto the
// cursor, switches desktop, then releases. The button is REAL system state that
// outlives the function: a panic between press and release is caught by the
// SDK's per-request recovery, so the process keeps running and nothing ever
// posts the matching up event — every subsequent click reads as a drag until
// the user physically clicks, and the plugin never notices.
//
// The fix is releaseOnce + `defer`, not a plain deferred release, because the
// drop has to land BEFORE the return-hop. These call the REAL releaseOnce from
// spaces.go rather than restating its logic — a test that reimplements the thing
// it checks certifies nothing.

func TestReleaseOnceFiresExactlyOnceOnTheHappyPath(t *testing.T) {
	calls := 0
	releaseMouse := releaseOnce(func() { calls++ })

	releaseMouse() // the in-sequence release, before the return-hop
	releaseMouse() // the deferred backstop

	if calls != 1 {
		t.Fatalf("release fired %d times, want exactly 1 — a second up event "+
			"on an unlatched button is a stray click", calls)
	}
}

func TestReleaseOnceStillFiresWhenTheSequenceIsAbandoned(t *testing.T) {
	calls := 0
	releaseMouse := releaseOnce(func() { calls++ })

	// Press happened; the in-sequence release never ran (panic recovered by the
	// SDK, or an early return added later). Only the defer is left.
	releaseMouse()

	if calls != 1 {
		t.Fatal("the backstop must release a latch the main path abandoned — " +
			"otherwise the left button stays down system-wide")
	}
}
