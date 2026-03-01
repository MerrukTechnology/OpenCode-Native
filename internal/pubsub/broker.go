package pubsub

import (
	"context"
	"sync"
)

// Default constants
const defaultBufferSize = 64

// Broker manages the publisher/subscriber pattern in a thread-safe, generic way.
type Broker[T any] struct {
	subs       map[chan Event[T]]struct{}
	mu         sync.RWMutex
	done       chan struct{}
	bufferSize int
}

// NewBroker creates a new Broker with default settings.
func NewBroker[T any]() *Broker[T] {
	return NewBrokerWithOptions[T](defaultBufferSize)
}

// NewBrokerWithOptions creates a new Broker with custom buffer size.
func NewBrokerWithOptions[T any](bufferSize int) *Broker[T] {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}
	return &Broker[T]{
		subs:       make(map[chan Event[T]]struct{}),
		done:       make(chan struct{}),
		bufferSize: bufferSize,
	}
}

// Subscribe adds a new subscriber and returns the channel + an unsubscribe function.
// The caller is responsible for consuming the channel.
func (b *Broker[T]) Subscribe() (<-chan Event[T], func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If broker is already dead, return a closed channel
	select {
	case <-b.done:
		ch := make(chan Event[T])
		close(ch)
		return ch, func() {}
	default:
	}

	// Create the channel with the CORRECT configured buffer size
	sub := make(chan Event[T], b.bufferSize)
	b.subs[sub] = struct{}{}

	// Unsubscribe function to be called by the user (cleaner than context handling inside)
	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		// Check if broker is already shutdown (subs map would be nil or empty logic)
		select {
		case <-b.done:
			return // Already cleaned up by Shutdown
		default:
		}

		// Only close if it exists in the map (prevents double close)
		if _, ok := b.subs[sub]; ok {
			delete(b.subs, sub)
			close(sub)
		}
	}

	return sub, unsubscribe
}

// SubscribeWithContext is a helper that unsubscribes automatically when context dies.
func (b *Broker[T]) SubscribeWithContext(ctx context.Context) <-chan Event[T] {
	ch, unsub := b.Subscribe()
	go func() {
		<-ctx.Done()
		unsub()
	}()
	return ch
}

// Publish broadcasts an event to all subscribers.
// It is non-blocking: if a subscriber is slow/full, they miss the message.
func (b *Broker[T]) Publish(t EventType, payload T) {
	// OPTIMIZATION: We hold the RLock during the send loop.
	// Because we use a 'select default' (non-blocking), this is extremely fast.
	// Holding the lock prevents a race condition where Shutdown() closes the channel
	// while we are trying to write to it.
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Check if shutdown
	select {
	case <-b.done:
		return
	default:
	}

	event := Event[T]{Type: t, Payload: payload}

	for sub := range b.subs {
		select {
		case sub <- event:
			// Sent successfully
		default:
			// Channel full: Drop message (Standard pattern for high-perf PubSub)
			// Optional: Increment a metric counter here
		}
	}
}

// Shutdown closes the broker and all subscriber channels.
func (b *Broker[T]) Shutdown() {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-b.done:
		return // Already shutdown
	default:
		close(b.done)
	}

	for sub := range b.subs {
		close(sub)
	}
	// Clear the map so subsequent Publish calls do nothing
	b.subs = nil
}

// GetSubscriberCount returns the current number of active subscribers.
func (b *Broker[T]) GetSubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs) // Use len(), no need for manual counter
}
