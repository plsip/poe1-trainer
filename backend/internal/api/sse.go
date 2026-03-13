package api

import (
	"fmt"
	"net/http"
	"sync"
)

// lineBroadcaster fans out string messages to all active SSE subscribers.
type lineBroadcaster struct {
	mu   sync.Mutex
	subs map[chan string]struct{}
}

func (b *lineBroadcaster) subscribe() chan string {
	ch := make(chan string, 64)
	b.mu.Lock()
	if b.subs == nil {
		b.subs = make(map[chan string]struct{})
	}
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *lineBroadcaster) unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
	// drain to unblock any pending send
	for len(ch) > 0 {
		<-ch
	}
}

func (b *lineBroadcaster) emit(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- msg:
		default: // drop if subscriber is slow
		}
	}
}

// signalBroadcaster fans out empty signals to per-ID subscriber sets.
type signalBroadcaster struct {
	mu   sync.Mutex
	subs map[int64]map[chan struct{}]struct{}
}

func (b *signalBroadcaster) subscribe(id int64) chan struct{} {
	ch := make(chan struct{}, 4)
	b.mu.Lock()
	if b.subs == nil {
		b.subs = make(map[int64]map[chan struct{}]struct{})
	}
	if b.subs[id] == nil {
		b.subs[id] = make(map[chan struct{}]struct{})
	}
	b.subs[id][ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *signalBroadcaster) unsubscribe(id int64, ch chan struct{}) {
	b.mu.Lock()
	if m := b.subs[id]; m != nil {
		delete(m, ch)
	}
	b.mu.Unlock()
	for len(ch) > 0 {
		<-ch
	}
}

func (b *signalBroadcaster) emit(id int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs[id] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// sseHeaders sets the required headers for a Server-Sent Events response.
func sseHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Prevents nginx from buffering SSE responses.
	w.Header().Set("X-Accel-Buffering", "no")
}

// sseEvent writes a named SSE event with data and flushes immediately.
func sseEvent(w http.ResponseWriter, flusher http.Flusher, eventType, data string) {
	if eventType != "" {
		fmt.Fprintf(w, "event: %s\n", eventType)
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}
