// internal/shared/event/publisher.go
package event

import (
	"context"
	"fmt"
	"log"
	// No need to import "github.com/hibiken/asynq" here unless you explicitly use asynq.Option etc.
	// The AsynqClient interface handles the dependency.
)

// Publisher interface (This will be moved to event_types.go)
// REMOVE THIS SECTION FROM THIS FILE, IT WILL BE IN event_types.go
/*
type Publisher interface {
    Publish(ctx context.Context, topicOrTaskName string, payload interface{}) error
}
*/

// --- Low Importance Publisher (Now uses custom In-Memory Pub/Sub) ---

// LowImportancePublisher implements the Publisher interface for in-memory events.
type LowImportancePublisher struct {
	inMemoryBus *InMemPubSub // Correctly references InMemPubSub from inmemory_bus.go
}

// NewLowImportancePublisher creates a new LowImportancePublisher.
// It accepts an instance of our custom InMemPubSub.
func NewLowImportancePublisher(bus *InMemPubSub) *LowImportancePublisher {
	return &LowImportancePublisher{inMemoryBus: bus}
}

// Publish sends a low-importance event to the in-memory bus.
func (p *LowImportancePublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	// For simplicity, this publisher only handles in-memory events.
	// You could add a switch statement here if one LowImportancePublisher
	// instance needed to handle different internal routing.

	// Ensure the payload matches one of our defined event payloads
	switch topic {
	case string(UserCreatedInMemoryEvent):
		if _, ok := payload.(UserCreatedPayload); !ok {
			return fmt.Errorf("invalid payload type for %s: %T", topic, payload)
		}
	case string(UserUpdatedInMemoryEvent):
		if _, ok := payload.(UserUpdatedPayload); !ok {
			return fmt.Errorf("invalid payload type for %s: %T", topic, payload)
		}
	case string(UserDeletedInMemoryEvent):
		if _, ok := payload.(UserDeletedPayload); !ok {
			return fmt.Errorf("invalid payload type for %s: %T", topic, payload)
		}
	// Add cases for other in-memory event topics if you introduce them
	default:
		return fmt.Errorf("unsupported in-memory event topic: %s", topic)
	}

	p.inMemoryBus.PublishEvent(Topic(topic), payload) // Call our custom bus's method
	log.Printf("INFO: Published In-Memory event: Topic='%s', Payload='%+v'\n", topic, payload)
	return nil
}

// --- High Importance Publisher (Asynq Client) ---

// AsynqClient defines the interface for enqueuing tasks.
// This matches the methods in internal/shared/event/asynq_client.go
type AsynqClient interface {
	EnqueueTask(taskType string, payload interface{}) error
	// You might add methods for other Asynq functionalities if needed (e.g., Close() if HighImportancePublisher manages lifecycle)
}

// HighImportancePublisher implements the Publisher interface for high-importance tasks using Asynq.
type HighImportancePublisher struct {
	asynqClient AsynqClient // Use the interface for flexibility
}

// NewHighImportancePublisher creates a new HighImportancePublisher.
func NewHighImportancePublisher(client AsynqClient) *HighImportancePublisher {
	return &HighImportancePublisher{asynqClient: client}
}

// Publish enqueues a high-importance task using Asynq.
func (p *HighImportancePublisher) Publish(ctx context.Context, taskType string, payload interface{}) error {
	// Asynq's EnqueueTask typically handles various payload types, but it's good practice
	// for the client to pass something easily marshaled (e.g., a struct, map, or []byte)
	if err := p.asynqClient.EnqueueTask(taskType, payload); err != nil {
		return fmt.Errorf("failed to enqueue Asynq task %s: %w", taskType, err)
	}
	log.Printf("INFO: Enqueued Asynq task: Type='%s', Payload='%+v'\n", taskType, payload)
	return nil
}
