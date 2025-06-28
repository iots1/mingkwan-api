// internal/shared/event/task_handlers.go
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// SendWelcomeEmailHandler handles the 'email.send.welcome' task.
func SendWelcomeEmailHandler(ctx context.Context, t *asynq.Task) error {
	var payload SendWelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Printf("ERROR: Failed to unmarshal SendWelcomeEmailPayload: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %w", asynq.SkipRetry) // Skip retry if payload is malformed
	}

	log.Printf("Asynq Worker: Sending welcome email to %s (%s) for User ID: %s\n",
		payload.Name, payload.Email, payload.UserID)

	// Simulate email sending delay
	time.Sleep(3 * time.Second)

	log.Printf("Asynq Worker: Welcome email sent successfully to %s.\n", payload.Email)
	return nil
}

// You can add more Asynq task handlers here.
// func ProcessPaymentHandler(ctx context.Context, t *asynq.Task) error { ... }
