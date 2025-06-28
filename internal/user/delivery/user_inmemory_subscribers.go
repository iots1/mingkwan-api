package delivery

import (
	"context"
	"errors" // Added for errors.Is when checking context.Done

	// Added for fmt.Errorf for better error messages
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
