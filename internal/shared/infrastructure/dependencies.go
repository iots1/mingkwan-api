package infrastructure

import (
	"context"

	"github.com/iots1/mingkwan-api/config"
	"github.com/iots1/mingkwan-api/internal/shared/adapters"
	"github.com/iots1/mingkwan-api/internal/shared/event"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type AppDependencies struct {
	AppCtx         context.Context
	DB             *mongo.Database
	RedisClient    *redis.Client
	LowPub         event.Publisher
	HighPub        event.Publisher
	InMemPubSub    *event.InMemPubSub
	AppConfig      config.AppConfig
	PasswordHasher adapters.PasswordHasher
}

func NewAppDependencies(
	ctx context.Context,
	db *mongo.Database,
	rdb *redis.Client,
	lowPub event.Publisher,
	highPub event.Publisher,
	inMemPubSub *event.InMemPubSub,
	appConfig config.AppConfig,
	passwordHasher adapters.PasswordHasher,
) AppDependencies {
	return AppDependencies{
		AppCtx:         ctx,
		DB:             db,
		RedisClient:    rdb,
		LowPub:         lowPub,
		HighPub:        highPub,
		InMemPubSub:    inMemPubSub,
		AppConfig:      appConfig,
		PasswordHasher: passwordHasher,
	}
}
