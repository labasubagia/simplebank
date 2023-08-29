package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

//go:generate go run go.uber.org/mock/mockgen -source=distributor.go -destination=./mock/distributor.go
type TaskDistributor interface {
	DistributeTaskVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
