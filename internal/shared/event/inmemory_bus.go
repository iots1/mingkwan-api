// internal/shared/event/inmemory_bus.go
package event

import (
	"log"
	"sync"
)

// subscriber defines a channel that listens for events on a specific topic.
type subscriber chan interface{}

// InMemPubSub is a simple in-memory publish-subscribe bus.
type InMemPubSub struct {
	mu          sync.RWMutex
	subscribers map[Topic][]subscriber // Map topic to a list of subscriber channels
}

var globalPubSub *InMemPubSub // Singleton instance

// NewInMemoryBus initializes and returns the singleton in-memory bus.
func NewInMemoryBus() *InMemPubSub {
	if globalPubSub == nil {
		globalPubSub = &InMemPubSub{
			subscribers: make(map[Topic][]subscriber),
		}
		log.Println("Initialized custom in-memory event bus.")
	}
	return globalPubSub
}

// PublishEvent publishes a payload to the specified topic.
func (p *InMemPubSub) PublishEvent(topic Topic, payload interface{}) {
	p.mu.RLock() // Use RLock as we are only reading the subscriber map
	defer p.mu.RUnlock()

	log.Printf("In-Memory Bus: Publishing event to topic '%s'", topic)

	// Iterate over all subscribers for this topic and send the payload
	// Use a goroutine for each send to prevent blocking if a channel is full or slow
	for _, sub := range p.subscribers[topic] {
		select {
		case sub <- payload: // Attempt non-blocking send
		default:
			log.Printf("In-Memory Bus: Subscriber channel for topic '%s' is full, dropping event for one listener.", topic)
		}
	}
}

// SubscribeEvent subscribes to a topic and returns a channel for events.
func (p *InMemPubSub) SubscribeEvent(topic Topic) chan interface{} {
	p.mu.Lock() // Use Lock as we are modifying the subscriber map
	defer p.mu.Unlock()

	ch := make(chan interface{}, 10) // Buffered channel to prevent blocking publisher
	p.subscribers[topic] = append(p.subscribers[topic], ch)
	log.Printf("In-Memory Bus: Subscribed to topic '%s'. Total subscribers for this topic: %d", topic, len(p.subscribers[topic]))
	return ch
}

// UnsubscribeEvent unsubscribes a channel from a topic.
// This is an optional method. For simplicity, we might not always need to unsubscribe explicitly
// if subscribers are short-lived or the app simply shuts down.
func (p *InMemPubSub) UnsubscribeEvent(topic Topic, ch chan interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	subs := p.subscribers[topic]
	for i, sub := range subs {
		if sub == ch {
			p.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			close(ch) // Close the channel when unsubscribed
			log.Printf("In-Memory Bus: Unsubscribed from topic '%s'. Remaining subscribers for this topic: %d", topic, len(p.subscribers[topic]))
			return
		}
	}
}
