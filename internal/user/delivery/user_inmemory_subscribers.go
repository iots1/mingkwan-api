// internal/user/delivery/user_inmemory_subscribers.go
package delivery

import (
	"context"
	"log"
	"time"

	"github.com/iots1/mingkwan-api/internal/shared/event" // Ensure this is imported
)

// UserInmemoryEventSubscribers handles in-memory events relevant to the user feature.
type UserInmemoryEventSubscribers struct {
	inMemoryBus *event.InMemPubSub // <--- Add this field to hold the bus instance
	// Add other dependencies needed to process events, e.g., some internal reporting service
}

// NewUserInmemoryEventSubscribers creates a new UserInmemoryEventSubscribers.
// It now accepts the in-memory bus as a dependency.
func NewUserInmemoryEventSubscribers(bus *event.InMemPubSub /*, other dependencies */) *UserInmemoryEventSubscribers {
	return &UserInmemoryEventSubscribers{
		inMemoryBus: bus, // <--- Initialize the bus
	}
}

// StartAllSubscribers kicks off all in-memory event listeners for this feature.
func (s *UserInmemoryEventSubscribers) StartAllSubscribers(ctx context.Context) {
	go s.listenToUserCreatedEvents(ctx)
	go s.listenToUserUpdatedEvents(ctx)
	go s.listenToUserDeletedEvents(ctx)
}

func (s *UserInmemoryEventSubscribers) listenToUserCreatedEvents(ctx context.Context) {
	// Call the SubscribeEvent method on the injected inMemoryBus instance
	ch := s.inMemoryBus.SubscribeEvent(event.UserCreatedInMemoryEvent)
	log.Println("UserFeature/In-Memory Subscriber: Listening for 'user.created.inmemory' events.")

	for {
		select {
		case eventData := <-ch:
			payload, ok := eventData.(event.UserCreatedPayload)
			if !ok {
				log.Printf("UserFeature/In-Memory Subscriber: Received unexpected payload type for 'user.created.inmemory' event. Data: %+v\n", eventData)
				continue
			}
			log.Printf("UserFeature/In-Memory Subscriber: UserCreatedInMemoryEvent for User ID %s (%s). Performing internal user-specific action...\n", payload.UserID, payload.Name)
			time.Sleep(50 * time.Millisecond) // Simulate work
			log.Printf("UserFeature/In-Memory Subscriber: Internal action complete for user %s.\n", payload.Name)
		case <-ctx.Done():
			// OPTIONAL: Call UnsubscribeEvent if the channel needs explicit closing before context is done.
			// s.inMemoryBus.UnsubscribeEvent(event.UserCreatedInMemoryEvent, ch)
			log.Println("UserFeature/In-Memory Subscriber: 'user.created.inmemory' event listener stopped.")
			return
		}
	}
}

func (s *UserInmemoryEventSubscribers) listenToUserUpdatedEvents(ctx context.Context) {
	ch := s.inMemoryBus.SubscribeEvent(event.UserUpdatedInMemoryEvent) // <--- Changed
	log.Println("UserFeature/In-Memory Subscriber: Listening for 'user.updated.inmemory' events.")

	for {
		select {
		case eventData := <-ch:
			payload, ok := eventData.(event.UserUpdatedPayload)
			if !ok {
				log.Printf("UserFeature/In-Memory Subscriber: Received unexpected payload type for 'user.updated.inmemory' event. Data: %+v\n", eventData)
				continue
			}
			log.Printf("UserFeature/In-Memory Subscriber: UserUpdatedInMemoryEvent for User ID %s (%s). Updating internal user cache...\n", payload.UserID, payload.Name)
			time.Sleep(50 * time.Millisecond) // Simulate work
		case <-ctx.Done():
			// s.inMemoryBus.UnsubscribeEvent(event.UserUpdatedInMemoryEvent, ch)
			log.Println("UserFeature/In-Memory Subscriber: 'user.updated.inmemory' event listener stopped.")
			return
		}
	}
}

func (s *UserInmemoryEventSubscribers) listenToUserDeletedEvents(ctx context.Context) {
	ch := s.inMemoryBus.SubscribeEvent(event.UserDeletedInMemoryEvent) // <--- Changed
	log.Println("UserFeature/In-Memory Subscriber: Listening for 'user.deleted.inmemory' events.")

	for {
		select {
		case eventData := <-ch:
			payload, ok := eventData.(event.UserDeletedPayload)
			if !ok {
				log.Printf("UserFeature/In-Memory Subscriber: Received unexpected payload type for 'user.deleted.inmemory' event. Data: %+v\n", eventData)
				continue
			}
			log.Printf("UserFeature/In-Memory Subscriber: UserDeletedInMemoryEvent for User ID %s. Cleaning up related data...\n", payload.UserID)
			time.Sleep(50 * time.Millisecond) // Simulate work
		case <-ctx.Done():
			// s.inMemoryBus.UnsubscribeEvent(event.UserDeletedInMemoryEvent, ch)
			log.Println("UserFeature/In-Memory Subscriber: 'user.deleted.inmemory' event listener stopped.")
			return
		}
	}
}
