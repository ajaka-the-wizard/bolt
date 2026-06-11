package store

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// mockRedisClient is a test double for the redisClient interface.
// It lets tests control what GetNextOnQueue returns and how many times it was called.
type mockRedisClient struct {
	// getNextOnQueueFn is called by GetNextOnQueue. If nil, returns nil, nil.
	getNextOnQueueFn func(ctx context.Context, id string, stream string, group string) ([]goredis.XStream, error)
	callCount        atomic.Int32
}

func (m *mockRedisClient) GetNextOnQueue(ctx context.Context, id string, stream string, group string) ([]goredis.XStream, error) {
	m.callCount.Add(1)
	if m.getNextOnQueueFn != nil {
		return m.getNextOnQueueFn(ctx, id, stream, group)
	}
	return nil, nil
}

func (m *mockRedisClient) SetIdemKey(_ context.Context, _ string) error        { return nil }
func (m *mockRedisClient) GetIdemKey(_ context.Context, _ string) (int, error) { return 0, nil }
func (m *mockRedisClient) AddToInvoiceQueue(_ context.Context, _ uuid.UUID) error {
	return nil
}

// newTestStore creates a Store backed by the given mock, suitable for unit tests.
func newTestStore(mock redisClient) *Store {
	return &Store{r: mock}
}

// ---- Tests -------------------------------------------------------------------

// TestFetchNextTask_SuccessOnFirstCall verifies that a successful response is
// returned immediately without any retry.
func TestFetchNextTask_SuccessOnFirstCall(t *testing.T) {
	want := []goredis.XStream{
		{Stream: "bolt:queue:invoice:", Messages: []goredis.XMessage{{ID: "1-0", Values: map[string]any{"order_id": "abc"}}}},
	}

	mock := &mockRedisClient{
		getNextOnQueueFn: func(_ context.Context, _, _, _ string) ([]goredis.XStream, error) {
			return want, nil
		},
	}
	s := newTestStore(mock)

	got, err := s.FetchNextTask(context.Background(), "worker-1", "stream", "group", slog.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d streams, want %d", len(got), len(want))
	}
	if mock.callCount.Load() != 1 {
		t.Fatalf("GetNextOnQueue called %d times, want 1", mock.callCount.Load())
	}
}

// TestFetchNextTask_ContextAlreadyCancelledOnError verifies the PR change:
// when ctx is already cancelled at the time GetNextOnQueue returns an error,
// FetchNextTask must return ctx.Err() immediately without entering the backoff
// retry loop.
func TestFetchNextTask_ContextAlreadyCancelledOnError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mock := &mockRedisClient{
		getNextOnQueueFn: func(ctx context.Context, _, _, _ string) ([]goredis.XStream, error) {
			// Cancel the context before returning the error, simulating a
			// cancellation that races with a queue read failure.
			cancel()
			return nil, errors.New("redis: connection closed")
		},
	}
	s := newTestStore(mock)

	got, err := s.FetchNextTask(ctx, "worker-1", "stream", "group", slog.Default())
	if got != nil {
		t.Fatalf("expected nil data, got %v", got)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	// GetNextOnQueue should have been called exactly once; no retry after ctx cancel.
	if mock.callCount.Load() != 1 {
		t.Fatalf("GetNextOnQueue called %d times after ctx cancel, want 1", mock.callCount.Load())
	}
}

// TestFetchNextTask_ContextAlreadyCancelledBeforeCall verifies that if the
// context is already cancelled before the first call, FetchNextTask returns
// ctx.Err() after a single attempt (the inner select on ctx.Done will fire).
func TestFetchNextTask_ContextAlreadyCancelledBeforeCall(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	mock := &mockRedisClient{
		getNextOnQueueFn: func(_ context.Context, _, _, _ string) ([]goredis.XStream, error) {
			return nil, errors.New("redis: context deadline exceeded")
		},
	}
	s := newTestStore(mock)

	_, err := s.FetchNextTask(ctx, "worker-1", "stream", "group", slog.Default())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestFetchNextTask_ContextCancelledViaDeadline verifies the same early-return
// behaviour when the context expires with DeadlineExceeded rather than Canceled.
func TestFetchNextTask_ContextCancelledViaDeadline(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second)) // already expired
	defer cancel()

	mock := &mockRedisClient{
		getNextOnQueueFn: func(_ context.Context, _, _, _ string) ([]goredis.XStream, error) {
			return nil, errors.New("deadline exceeded")
		},
	}
	s := newTestStore(mock)

	_, err := s.FetchNextTask(ctx, "worker-1", "stream", "group", slog.Default())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

// TestFetchNextTask_RetriesOnTransientErrorThenSucceeds verifies that a
// transient error (with a non-cancelled context) causes a retry, and that the
// function returns successfully on the subsequent attempt.
func TestFetchNextTask_RetriesOnTransientErrorThenSucceeds(t *testing.T) {
	want := []goredis.XStream{
		{Stream: "bolt:queue:invoice:", Messages: []goredis.XMessage{{ID: "2-0"}}},
	}

	attempt := 0
	mock := &mockRedisClient{
		getNextOnQueueFn: func(_ context.Context, _, _, _ string) ([]goredis.XStream, error) {
			attempt++
			if attempt == 1 {
				return nil, errors.New("temporary network error")
			}
			return want, nil
		},
	}
	s := newTestStore(mock)

	// Use a short deadline so backoff doesn't make the test slow.
	// The initial backoff is 1 s; patch it by using a cancellable context that
	// we never cancel — the second attempt must succeed within a few seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	got, err := s.FetchNextTask(ctx, "worker-1", "stream", "group", slog.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected non-empty result on retry")
	}
	if mock.callCount.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", mock.callCount.Load())
	}
}

// TestFetchNextTask_CtxDoneSelectPathOnError verifies that when ctx is not yet
// cancelled when the error occurs, but is cancelled while waiting in the backoff
// select, FetchNextTask returns ctx.Err().
func TestFetchNextTask_CtxDoneSelectPathOnError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	mock := &mockRedisClient{
		getNextOnQueueFn: func(_ context.Context, _, _, _ string) ([]goredis.XStream, error) {
			callCount++
			// On the first call, return an error but do NOT cancel the context yet,
			// so we enter the backoff select. Cancel after a small delay so the
			// ctx.Done case fires before the 1-second timer.
			if callCount == 1 {
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}
			return nil, errors.New("transient error")
		},
	}
	s := newTestStore(mock)

	_, err := s.FetchNextTask(ctx, "worker-1", "stream", "group", slog.Default())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled from backoff select, got %v", err)
	}
}

// TestFetchNextTask_ReturnsNilErrorOnSuccess is a boundary check: an explicit
// nil error from the queue reader must propagate as nil to the caller.
func TestFetchNextTask_ReturnsNilErrorOnSuccess(t *testing.T) {
	mock := &mockRedisClient{
		getNextOnQueueFn: func(_ context.Context, _, _, _ string) ([]goredis.XStream, error) {
			return []goredis.XStream{}, nil
		},
	}
	s := newTestStore(mock)

	_, err := s.FetchNextTask(context.Background(), "w", "s", "g", slog.Default())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}