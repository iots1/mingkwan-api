// internal/shared/event/asynq_client.go
package event

import (
	"fmt"
	"log"

	"github.com/hibiken/asynq" // ตรวจสอบให้แน่ใจว่าได้ Import อันนี้แล้ว
)

// AsynqClientImpl implements the AsynqClient interface (defined in publisher.go)
type AsynqClientImpl struct {
	Client *asynq.Client
}

// NewAsynqClient creates a new AsynqClientImpl instance.
func NewAsynqClient(redisOpt asynq.RedisClientOpt) *AsynqClientImpl {
	client := asynq.NewClient(redisOpt)
	return &AsynqClientImpl{Client: client}
}

// EnqueueTask enqueues a new task with default options (e.g., critical queue).
func (a *AsynqClientImpl) EnqueueTask(taskType string, payload interface{}) error {
	task := asynq.NewTask(taskType, []byte(fmt.Sprintf("%v", payload)),
		asynq.Queue("critical"), asynq.MaxRetry(3))

	info, err := a.Client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("could not enqueue task %s: %w", taskType, err)
	}
	log.Printf("INFO: Enqueued task: id=%s, type=%s, queue=%s\n", info.ID, info.Type, info.Queue)
	return nil
}

// Close closes the underlying Asynq client connection.
func (a *AsynqClientImpl) Close() error {
	if a.Client == nil {
		return nil
	}
	log.Println("INFO: Closing Asynq client...")
	return a.Client.Close()
}

// --- เพิ่มฟังก์ชันใหม่นี้ ---
// GetRedisClientOpt creates an asynq.RedisClientOpt from connection details.
// ฟังก์ชันนี้ต้องถูก Export (ขึ้นต้นด้วยตัวพิมพ์ใหญ่ 'G')
// เพื่อให้สามารถเรียกใช้จาก main.go ได้
func GetRedisClientOpt(addr string, password string, db int) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
}

// --- สิ้นสุดฟังก์ชันใหม่ ---

// Ensure AsynqClientImpl implements the AsynqClient interface from publisher.go
var _ AsynqClient = (*AsynqClientImpl)(nil)
