package delivery

import (
	"context"
	"errors" // Added for errors.Is when checking context.Done
	// Added for fmt.Errorf for better error messages
	"time"

	"go.uber.org/zap" // Make sure zap is imported for utils.Logger

	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/iots1/mingkwan-api/internal/shared/utils" // Import utils for utils.Logger
)

type UserInmemoryEventSubscribers struct {
	inMemoryBus *event.InMemPubSub
	// Add other dependencies needed to process events, e.g., some internal reporting service
}

func NewUserInmemoryEventSubscribers(bus *event.InMemPubSub /*, other dependencies */) *UserInmemoryEventSubscribers {
	return &UserInmemoryEventSubscribers{
		inMemoryBus: bus,
	}
}

func (s *UserInmemoryEventSubscribers) StartAllSubscribers(ctx context.Context) {
	go s.listenToUserCreatedEvents(ctx)
	go s.listenToUserUpdatedEvents(ctx)
	go s.listenToUserDeletedEvents(ctx)
	utils.Logger.Info("UserFeature/In-Memory Subscribers: All listeners started.")
}

func (s *UserInmemoryEventSubscribers) listenToUserCreatedEvents(ctx context.Context) {
	ch := s.inMemoryBus.SubscribeEvent(event.UserCreatedInMemoryEvent)
	utils.Logger.Info("UserFeature/In-Memory Subscriber: Listening for 'user.created.inmemory' events.")

	for {
		select {
		case eventData := <-ch:
			payload, ok := eventData.(event.UserCreatedPayload)
			if !ok {
				utils.Logger.Warn(
					"UserFeature/In-Memory Subscriber: Received unexpected payload type for 'user.created.inmemory' event.",
					zap.Any("event_data", eventData),
				)
				continue
			}
			utils.Logger.Info(
				"UserFeature/In-Memory Subscriber: UserCreatedInMemoryEvent received.",
				zap.String("user_id", payload.UserID.Hex()),
				zap.String("user_name", payload.Name),
				zap.String("action", "Performing internal user-specific action..."),
			)
			time.Sleep(50 * time.Millisecond) // Simulate work
			utils.Logger.Info(
				"UserFeature/In-Memory Subscriber: Internal action complete.",
				zap.String("user_name", payload.Name),
			)
		case <-ctx.Done():
			// It's good practice to log why the context was cancelled.
			err := ctx.Err()
			if errors.Is(err, context.Canceled) {
				utils.Logger.Info("UserFeature/In-Memory Subscriber: 'user.created.inmemory' event listener stopped due to context cancellation.")
			} else if errors.Is(err, context.DeadlineExceeded) {
				utils.Logger.Info("UserFeature/In-Memory Subscriber: 'user.created.inmemory' event listener stopped due to context deadline exceeded.")
			} else {
				utils.Logger.Error("UserFeature/In-Memory Subscriber: 'user.created.inmemory' event listener stopped unexpectedly.", zap.Error(err))
			}
			return
		}
	}
}

func (s *UserInmemoryEventSubscribers) listenToUserUpdatedEvents(ctx context.Context) {
	ch := s.inMemoryBus.SubscribeEvent(event.UserUpdatedInMemoryEvent)
	utils.Logger.Info("UserFeature/In-Memory Subscriber: Listening for 'user.updated.inmemory' events.")

	for {
		select {
		case eventData := <-ch:
			payload, ok := eventData.(event.UserUpdatedPayload)
			if !ok {
				utils.Logger.Warn(
					"UserFeature/In-Memory Subscriber: Received unexpected payload type for 'user.updated.inmemory' event.",
					zap.Any("event_data", eventData),
				)
				continue
			}
			utils.Logger.Info(
				"UserFeature/In-Memory Subscriber: UserUpdatedInMemoryEvent received.",
				zap.String("user_id", payload.UserID.Hex()),
				zap.String("user_name", payload.Name),
				zap.String("action", "Updating internal user cache..."),
			)
			time.Sleep(50 * time.Millisecond) // Simulate work
		case <-ctx.Done():
			err := ctx.Err()
			if errors.Is(err, context.Canceled) {
				utils.Logger.Info("UserFeature/In-Memory Subscriber: 'user.updated.inmemory' event listener stopped due to context cancellation.")
			} else if errors.Is(err, context.DeadlineExceeded) {
				utils.Logger.Info("UserFeature/In-Memory Subscriber: 'user.updated.inmemory' event listener stopped due to context deadline exceeded.")
			} else {
				utils.Logger.Error("UserFeature/In-Memory Subscriber: 'user.updated.inmemory' event listener stopped unexpectedly.", zap.Error(err))
			}
			return
		}
	}
}

func (s *UserInmemoryEventSubscribers) listenToUserDeletedEvents(ctx context.Context) {
	ch := s.inMemoryBus.SubscribeEvent(event.UserDeletedInMemoryEvent)
	utils.Logger.Info("UserFeature/In-Memory Subscriber: Listening for 'user.deleted.inmemory' events.")

	for {
		select {
		case eventData := <-ch:
			payload, ok := eventData.(event.UserDeletedPayload)
			if !ok {
				utils.Logger.Warn(
					"UserFeature/In-Memory Subscriber: Received unexpected payload type for 'user.deleted.inmemory' event.",
					zap.Any("event_data", eventData),
				)
				continue
			}
			utils.Logger.Info(
				"UserFeature/In-Memory Subscriber: UserDeletedInMemoryEvent received.",
				zap.String("user_id", payload.UserID.Hex()),
				zap.String("action", "Cleaning up related data..."),
			)
			time.Sleep(50 * time.Millisecond) // Simulate work
		case <-ctx.Done():
			err := ctx.Err()
			if errors.Is(err, context.Canceled) {
				utils.Logger.Info("UserFeature/In-Memory Subscriber: 'user.deleted.inmemory' event listener stopped due to context cancellation.")
			} else if errors.Is(err, context.DeadlineExceeded) {
				utils.Logger.Info("UserFeature/In-Memory Subscriber: 'user.deleted.inmemory' event listener stopped due to context deadline exceeded.")
			} else {
				utils.Logger.Error("UserFeature/In-Memory Subscriber: 'user.deleted.inmemory' event listener stopped unexpectedly.", zap.Error(err))
			}
			return
		}
	}
}
