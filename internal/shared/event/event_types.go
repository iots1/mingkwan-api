// internal/shared/event/event_types.go
package event

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Topic represents an event topic/name
type Topic string

// Define your in-memory event topics
const (
	UserCreatedInMemoryEvent = Topic("user.created.inmemory")
	UserUpdatedInMemoryEvent = Topic("user.updated.inmemory")
	UserDeletedInMemoryEvent = Topic("user.deleted.inmemory")
)

// --- NEW --- Define Asynq Task Names
const (
	SendWelcomeEmailTaskName         = "user:send_welcome_email" // Define this task name
	UserDeletedHighImportance string = "user:deleted_high_importance"
)

// --- END NEW ---

// Define payloads for your in-memory events
type UserCreatedPayload struct {
	UserID primitive.ObjectID `json:"userId"`
	Name   string             `json:"name"`
	Email  string             `json:"email"`
}

type UserUpdatedPayload struct {
	UserID primitive.ObjectID `json:"userId"`
	Name   string             `json:"name,omitempty"`
	Email  string             `json:"email,omitempty"`
}

type UserDeletedPayload struct {
	UserID primitive.ObjectID `json:"userId"`
}

// --- NEW --- Define Payload for SendWelcomeEmailTaskName
type SendWelcomeEmailPayload struct {
	UserID string `json:"user_id"` // Assuming you convert ObjectID to string for Asynq
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// --- END NEW ---

// Unified Publisher interface: All publishers (in-memory, Asynq) will implement this.
type Publisher interface {
	Publish(ctx context.Context, topicOrTaskName string, payload interface{}) error
}

// EventHandler (still needed for InMemPubSub's subscription mechanism)
type EventHandler func(ctx context.Context, payload interface{}) error
