package redis

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"shortbin/pkg/config"
	"shortbin/pkg/logger"
)

const (
	ContextTimeout     = 3
	InitContextTimeout = 5
)

// IRedis interface
type IRedis interface {
	Get(key string, value interface{}) error
	GetByRefreshingExpiry(key string, value interface{}) error
	Set(key string, value interface{}, expiryTime time.Duration) error
	SetExpiry(key string, expiryTime time.Duration) error
}

// Config redis
type Config struct {
	Address  string
	Password string
	Database int
}

const NilReturn = goredis.Nil

type redis struct {
	cmd goredis.Cmdable
}

// New Redis interface with config
func New(config Config) IRedis {
	ctx, cancel := context.WithTimeout(context.Background(), InitContextTimeout*time.Second)
	defer cancel()

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.Database,
	})

	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Fatal(pong, err)
		return nil
	}

	return &redis{
		cmd: redisClient,
	}
}

func (r *redis) Get(key string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout*time.Second)
	defer cancel()

	strValue, err := r.cmd.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(strValue), value)
	if err != nil {
		return err
	}

	return nil
}

func (r *redis) GetByRefreshingExpiry(key string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout*time.Second)
	defer cancel()

	strValue, err := r.cmd.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	go func() {
		err := r.SetExpiry(key, config.GetConfig().Redis.TTL*time.Minute)
		if err != nil {
			logger.Error("error refreshing ttl: ", err)
		}
	}()

	err = json.Unmarshal([]byte(strValue), value)
	if err != nil {
		return err
	}

	return nil
}

func (r *redis) Set(key string, value interface{}, expiryTime time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout*time.Second)
	defer cancel()

	bData, _ := json.Marshal(value)
	err := r.cmd.Set(ctx, key, bData, expiryTime).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *redis) SetExpiry(key string, expiryTime time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout*time.Second)
	defer cancel()

	_, err := r.cmd.Expire(ctx, key, expiryTime).Result()
	if err != nil {
		return err
	}

	return nil
}
