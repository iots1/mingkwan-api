package event

import (
	"sync"

	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/internal/shared/utils"
)

type subscriber chan interface{}

type InMemPubSub struct {
	mu          sync.RWMutex
	subscribers map[Topic][]subscriber
}

var globalPubSub *InMemPubSub

func NewInMemoryBus() *InMemPubSub {
	if globalPubSub == nil {
		globalPubSub = &InMemPubSub{
			subscribers: make(map[Topic][]subscriber),
		}
		utils.Logger.Info("Initialized custom in-memory event bus.")
	}
	return globalPubSub
}

func (p *InMemPubSub) PublishEvent(topic Topic, payload interface{}) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	utils.Logger.Debug("In-Memory Bus: Publishing event", zap.String("topic", string(topic)))

	for _, sub := range p.subscribers[topic] {
		select {
		case sub <- payload:
		default:
			utils.Logger.Warn("In-Memory Bus: Subscriber channel is full, dropping event",
				zap.String("topic", string(topic)),
			)
		}
	}
}

func (p *InMemPubSub) SubscribeEvent(topic Topic) chan interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan interface{}, 10)
	p.subscribers[topic] = append(p.subscribers[topic], ch)
	utils.Logger.Info("In-Memory Bus: Subscribed to topic",
		zap.String("topic", string(topic)),
		zap.Int("total_subscribers", len(p.subscribers[topic])),
	)
	return ch
}

func (p *InMemPubSub) UnsubscribeEvent(topic Topic, ch chan interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	subs := p.subscribers[topic]
	for i, sub := range subs {
		if sub == ch {
			p.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			close(ch)
			utils.Logger.Info("In-Memory Bus: Unsubscribed from topic",
				zap.String("topic", string(topic)),
				zap.Int("remaining_subscribers", len(p.subscribers[topic])),
			)
			return
		}
	}
}
