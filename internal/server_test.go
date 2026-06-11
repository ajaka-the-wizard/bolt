package internal

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

// ---- Test doubles ------------------------------------------------------------

// callRecorder tracks the order in which methods were called across goroutines.
type callRecorder struct {
	mu    sync.Mutex
	calls []string
}

func (r *callRecorder) record(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, name)
}

func (r *callRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.calls))
	copy(out, r.calls)
	return out
}

// mockApp implements appShutdowner.
type mockApp struct {
	rec        *callRecorder
	shutdownFn func() error
}

func (a *mockApp) Shutdown() error {
	a.rec.record("app.Shutdown")
	if a.shutdownFn != nil {
		return a.shutdownFn()
	}
	return nil
}

// mockRDB implements redisConnCloser.
type mockRDB struct {
	rec         *callRecorder
	closeConnFn func() error
}

func (r *mockRDB) CloseConn() error {
	r.rec.record("rdb.CloseConn")
	if r.closeConnFn != nil {
		return r.closeConnFn()
	}
	return nil
}

// mockDB implements dbConnCloser.
type mockDB struct {
	rec         *callRecorder
	closeConnFn func()
}

func (d *mockDB) CloseConn() {
	d.rec.record("db.CloseConn")
	if d.closeConnFn != nil {
		d.closeConnFn()
	}
}

// ---- Helpers -----------------------------------------------------------------

// sendSignalAndWait sends sig to ch and blocks until done is closed or timeout
// elapses. Fails the test on timeout.
func sendSignalAndWait(t *testing.T, ch chan<- os.Signal, done <-chan struct{}, timeout time.Duration) {
	t.Helper()
	ch <- os.Interrupt
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatal("shutdown goroutine did not complete within timeout")
	}
}

// ---- Tests -------------------------------------------------------------------

// TestDoShutdown_OrderingAppBeforeConnections is the primary regression test
// for the PR change.  It verifies that app.Shutdown() is called BEFORE
// rdb.CloseConn() and db.CloseConn().
func TestDoShutdown_OrderingAppBeforeConnections(t *testing.T) {
	rec := &callRecorder{}
	done := make(chan struct{})

	app := &mockApp{rec: rec}
	rdb := &mockRDB{rec: rec}
	db := &mockDB{
		rec: rec,
		closeConnFn: func() { close(done) }, // signals test that goroutine finished
	}

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)

	doShutdown(sig, slog.Default(), app, db, rdb, cancel)
	sendSignalAndWait(t, sig, done, 3*time.Second)

	calls := rec.snapshot()
	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d: %v", len(calls), calls)
	}

	want := []string{"app.Shutdown", "rdb.CloseConn", "db.CloseConn"}
	for i, w := range want {
		if calls[i] != w {
			t.Errorf("call[%d] = %q, want %q (full order: %v)", i, calls[i], w, calls)
		}
	}

	// Context must have been cancelled.
	select {
	case <-ctx.Done():
	default:
		t.Error("context was not cancelled after shutdown signal")
	}
}

// TestDoShutdown_CancelCalledBeforeAppShutdown verifies that cancel() is
// invoked before app.Shutdown(), ensuring in-flight request handlers observe
// the cancelled context promptly.
func TestDoShutdown_CancelCalledBeforeAppShutdown(t *testing.T) {
	rec := &callRecorder{}
	done := make(chan struct{})

	cancelCalled := false
	cancelFn := context.CancelFunc(func() {
		rec.record("cancel")
		cancelCalled = true
	})

	app := &mockApp{rec: rec}
	rdb := &mockRDB{rec: rec}
	db := &mockDB{
		rec:         rec,
		closeConnFn: func() { close(done) },
	}

	sig := make(chan os.Signal, 1)
	doShutdown(sig, slog.Default(), app, db, rdb, cancelFn)
	sendSignalAndWait(t, sig, done, 3*time.Second)

	if !cancelCalled {
		t.Fatal("cancel was never called")
	}

	calls := rec.snapshot()
	// cancel must appear before app.Shutdown
	cancelIdx, appIdx := -1, -1
	for i, c := range calls {
		switch c {
		case "cancel":
			cancelIdx = i
		case "app.Shutdown":
			appIdx = i
		}
	}
	if cancelIdx == -1 {
		t.Fatal("cancel not recorded")
	}
	if appIdx == -1 {
		t.Fatal("app.Shutdown not recorded")
	}
	if cancelIdx >= appIdx {
		t.Errorf("cancel (idx %d) must come before app.Shutdown (idx %d); order: %v", cancelIdx, appIdx, calls)
	}
}

// TestDoShutdown_AppShutdownErrorDoesNotPreventConnCleanup verifies that if
// app.Shutdown() returns an error, the function still proceeds to close redis
// and database connections.
func TestDoShutdown_AppShutdownErrorDoesNotPreventConnCleanup(t *testing.T) {
	rec := &callRecorder{}
	done := make(chan struct{})

	app := &mockApp{
		rec:        rec,
		shutdownFn: func() error { return errors.New("shutdown failed") },
	}
	rdb := &mockRDB{rec: rec}
	db := &mockDB{
		rec:         rec,
		closeConnFn: func() { close(done) },
	}

	_, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	doShutdown(sig, slog.Default(), app, db, rdb, cancel)
	sendSignalAndWait(t, sig, done, 3*time.Second)

	calls := rec.snapshot()
	hasRDB, hasDB := false, false
	for _, c := range calls {
		if c == "rdb.CloseConn" {
			hasRDB = true
		}
		if c == "db.CloseConn" {
			hasDB = true
		}
	}
	if !hasRDB {
		t.Error("rdb.CloseConn was not called after app.Shutdown error")
	}
	if !hasDB {
		t.Error("db.CloseConn was not called after app.Shutdown error")
	}
}

// TestDoShutdown_RedisCloseErrorDoesNotPreventDBCleanup verifies that an error
// from rdb.CloseConn() does not prevent db.CloseConn() from running.
func TestDoShutdown_RedisCloseErrorDoesNotPreventDBCleanup(t *testing.T) {
	rec := &callRecorder{}
	done := make(chan struct{})

	app := &mockApp{rec: rec}
	rdb := &mockRDB{
		rec:         rec,
		closeConnFn: func() error { return errors.New("redis close error") },
	}
	db := &mockDB{
		rec:         rec,
		closeConnFn: func() { close(done) },
	}

	_, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	doShutdown(sig, slog.Default(), app, db, rdb, cancel)
	sendSignalAndWait(t, sig, done, 3*time.Second)

	calls := rec.snapshot()
	hasDB := false
	for _, c := range calls {
		if c == "db.CloseConn" {
			hasDB = true
		}
	}
	if !hasDB {
		t.Error("db.CloseConn was not called after rdb.CloseConn error")
	}
}

// TestDoShutdown_AllThreeResourcesClosedOnSignal verifies that all three
// cleanup operations (app, rdb, db) are executed exactly once per signal.
func TestDoShutdown_AllThreeResourcesClosedOnSignal(t *testing.T) {
	rec := &callRecorder{}
	done := make(chan struct{})

	app := &mockApp{rec: rec}
	rdb := &mockRDB{rec: rec}
	db := &mockDB{
		rec:         rec,
		closeConnFn: func() { close(done) },
	}

	_, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	doShutdown(sig, slog.Default(), app, db, rdb, cancel)
	sendSignalAndWait(t, sig, done, 3*time.Second)

	counts := map[string]int{}
	for _, c := range rec.snapshot() {
		counts[c]++
	}
	for _, name := range []string{"app.Shutdown", "rdb.CloseConn", "db.CloseConn"} {
		if counts[name] != 1 {
			t.Errorf("%s called %d times, want 1", name, counts[name])
		}
	}
}

// TestDoShutdown_BlocksUntilSignal verifies that before a signal is sent, none
// of the cleanup methods are called (the goroutine is waiting).
func TestDoShutdown_BlocksUntilSignal(t *testing.T) {
	rec := &callRecorder{}

	app := &mockApp{rec: rec}
	rdb := &mockRDB{rec: rec}
	db := &mockDB{rec: rec}

	_, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	doShutdown(sig, slog.Default(), app, db, rdb, cancel)

	// Give the goroutine a moment to proceed if it were to do so incorrectly.
	time.Sleep(50 * time.Millisecond)

	calls := rec.snapshot()
	if len(calls) != 0 {
		t.Errorf("cleanup methods called before signal: %v", calls)
	}
}