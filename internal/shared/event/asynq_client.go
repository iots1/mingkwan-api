package event

import (
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/internal/shared/utils"
)

type AsynqClientImpl struct {
	Client *asynq.Client
}

func NewAsynqClient(redisOpt asynq.RedisClientOpt) *AsynqClientImpl {
	client := asynq.NewClient(redisOpt)
	return &AsynqClientImpl{Client: client}
}

func (a *AsynqClientImpl) EnqueueTask(taskType string, payload interface{}) error {
	task := asynq.NewTask(taskType, []byte(fmt.Sprintf("%v", payload)),
		asynq.Queue("critical"), asynq.MaxRetry(3))

	info, err := a.Client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("could not enqueue task %s: %w", taskType, err)
	}
	utils.Logger.Info("Enqueued task",
		zap.String("id", info.ID),
		zap.String("type", info.Type),
		zap.String("queue", info.Queue),
	)
	return nil
}

func (a *AsynqClientImpl) Close() error {
	if a.Client == nil {
		return nil
	}
	utils.Logger.Info("Closing Asynq client...")
	return a.Client.Close()
}

func GetRedisClientOpt(addr string, password string, db int) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
}

var _ AsynqClient = (*AsynqClientImpl)(nil)
