package ws

import (
	"context"
	"testing"
	"time"
)

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("condition not met before timeout")
}

func TestSubscribeTickSizeChangeDispatchAndCleanup(t *testing.T) {
	client := NewClient()
	connCtx, connCancel := context.WithCancel(context.Background())
	conn := &connection{
		ctx:    connCtx,
		cancel: connCancel,
	}

	client.mu.Lock()
	client.marketConn = conn
	client.mu.Unlock()

	subCtx, subCancel := context.WithCancel(context.Background())
	out := client.SubscribeTickSizeChange(subCtx, "1")

	waitFor(t, time.Second, func() bool {
		conn.listMu.Lock()
		defer conn.listMu.Unlock()
		return len(conn.listeners) == 1
	})

	conn.dispatchSingle([]byte(`{"event_type":"tick_size_change","asset_id":"1","market":"m","old_tick_size":"0.01","new_tick_size":"0.001","timestamp":"t"}`))

	select {
	case msg := <-out:
		if msg.AssetID != "1" || msg.NewTickSize != "0.001" {
			t.Fatalf("unexpected message: %+v", msg)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for tick-size message")
	}

	subCancel()

	waitFor(t, time.Second, func() bool {
		conn.listMu.Lock()
		defer conn.listMu.Unlock()
		return len(conn.listeners) == 0
	})

	select {
	case _, ok := <-out:
		if ok {
			t.Fatalf("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for closed output channel")
	}
}

func TestWithConnectionContextControlsLifecycle(t *testing.T) {
	parentCtx, parentCancel := context.WithCancel(context.Background())
	client := NewClient(WithConnectionContext(parentCtx), WithEndpoint("ws://127.0.0.1:1"))

	conn := client.getMarketConn(context.Background())
	if conn == nil {
		t.Fatalf("expected market connection")
	}

	parentCancel()

	waitFor(t, time.Second, func() bool {
		return conn.ctx.Err() != nil
	})

	client.Close()
}
