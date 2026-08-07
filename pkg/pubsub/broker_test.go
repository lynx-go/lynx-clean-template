package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	lxpubsub "github.com/lynx-go/lynx/contrib/pubsub"
)

type helloData struct {
	Message string `json:"message"`
}

func TestTypedSubscribe(t *testing.T) {
	broker := NewPubSub(lxpubsub.NewBroker(lxpubsub.Options{
		DefaultTransport: lxpubsub.NewMemoryTransport(),
	}))
	if err := broker.Init(nil); err != nil {
		t.Fatal(err)
	}

	var (
		mu     sync.Mutex
		got    *cloudevents.Event
		gotErr error
		done   = make(chan struct{})
	)
	if err := broker.Subscribe("hello", "hello", "TestTypedHandler",
		func(ctx context.Context, e *cloudevents.Event) error {
			mu.Lock()
			defer mu.Unlock()
			got = e
			close(done)
			return nil
		}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := broker.Start(ctx); err != nil {
			mu.Lock()
			gotErr = err
			mu.Unlock()
			t.Logf("broker start error: %v", err)
		}
	}()

	// 等待订阅注册完成（memory transport 的订阅在 Start 时建立）
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		received := got != nil
		mu.Unlock()
		if received {
			break
		}
		if err := broker.Publish(context.Background(), "hello", "hello", &helloData{Message: "hi"}); err != nil {
			t.Fatalf("publish: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("handler not called within timeout")
	}

	mu.Lock()
	defer mu.Unlock()
	if gotErr != nil {
		t.Fatalf("broker start error: %v", gotErr)
	}
	if got == nil {
		t.Fatal("no event received")
	}
	if got.Type() != "hello" {
		t.Errorf("event type = %q, want hello", got.Type())
	}
	if got.Source() == "" {
		t.Error("event source is empty")
	}
	var data helloData
	if err := got.DataAs(&data); err != nil {
		t.Fatalf("data as: %v", err)
	}
	if data.Message != "hi" {
		t.Errorf("message = %q, want hi", data.Message)
	}
}

func TestTypedSubscribeFiltersEventType(t *testing.T) {
	broker := NewPubSub(lxpubsub.NewBroker(lxpubsub.Options{
		DefaultTransport: lxpubsub.NewMemoryTransport(),
	}))
	if err := broker.Init(nil); err != nil {
		t.Fatal(err)
	}

	called := make(chan struct{}, 1)
	if err := broker.Subscribe("hello", "other-event", "TestFilterHandler",
		func(ctx context.Context, e *cloudevents.Event) error {
			called <- struct{}{}
			return nil
		}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = broker.Start(ctx) }()
	time.Sleep(100 * time.Millisecond)

	if err := broker.Publish(context.Background(), "hello", "hello", &helloData{Message: "hi"}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case <-called:
		t.Fatal("handler should not be called for different event type")
	case <-time.After(500 * time.Millisecond):
	}
}
