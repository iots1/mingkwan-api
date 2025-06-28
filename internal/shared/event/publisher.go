package event

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/internal/shared/utils"
)

// LowImportancePublisher implements the Publisher interface for in-memory events.
type LowImportancePublisher struct {
	inMemoryBus *InMemPubSub
}

// NewLowImportancePublisher creates a new LowImportancePublisher.
func NewLowImportancePublisher(bus *InMemPubSub) *LowImportancePublisher {
	return &LowImportancePublisher{inMemoryBus: bus}
}

// Publish sends a low-importance event to the in-memory bus.
func (p *LowImportancePublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	switch topic {
	case string(UserCreatedInMemoryEvent):
		if _, ok := payload.(UserCreatedPayload); !ok {
			return fmt.Errorf("invalid payload type for %s: %T", topic, payload)
		}
	default:
		return fmt.Errorf("unsupported in-memory event topic: %s", topic)
	}

	p.inMemoryBus.PublishEvent(Topic(topic), payload)
	utils.Logger.Info("Published In-Memory event",
		zap.String("topic", topic),
		zap.Any("payload", payload),
	)
	return nil
}

// AsynqClient defines the interface for enqueuing tasks.
type AsynqClient interface {
	EnqueueTask(taskType string, payload interface{}) error
}

// HighImportancePublisher implements the Publisher interface for high-importance tasks using Asynq.
type HighImportancePublisher struct {
	asynqClient AsynqClient
}

// NewHighImportancePublisher creates a new HighImportancePublisher.
func NewHighImportancePublisher(client AsynqClient) *HighImportancePublisher {
	return &HighImportancePublisher{asynqClient: client}
}

// Publish enqueues a high-importance task using Asynq.
func (p *HighImportancePublisher) Publish(ctx context.Context, taskType string, payload interface{}) error {
	if err := p.asynqClient.EnqueueTask(taskType, payload); err != nil {
		return fmt.Errorf("failed to enqueue Asynq task %s: %w", taskType, err)
	}
	utils.Logger.Info("Enqueued Asynq task",
		zap.String("type", taskType),
		zap.Any("payload", payload),
	)
	return nil
}
