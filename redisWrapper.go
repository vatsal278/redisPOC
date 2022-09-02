package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type config struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	MaxRetries   int
	DialTimeout  time.Duration
	PoolSize     int
	MinIdleConns int
}
type RedisSdk interface {
	redisGet(string) ([]byte, error)
	redisSet(string, interface{}) error
}

func NewHandler(c config) RedisSdk {
	return &config{
		Addr:         c.Addr,
		Username:     c.Username,
		Password:     c.Password, // no password set
		DB:           c.DB,       // use default DB
		MaxRetries:   c.MaxRetries,
		DialTimeout:  c.DialTimeout,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
	}
}
func (c config) redisGet(key string) ([]byte, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Username:     c.Username,
		Password:     c.Password, // no password set
		DB:           c.DB,       // use default DB
		MaxRetries:   c.MaxRetries,
		DialTimeout:  c.DialTimeout,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
	})
	data, err := rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		log.Print(err)
	}
	return data, err
}
func (c config) redisSet(key string, value interface{}) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Username:     c.Username,
		Password:     c.Password, // no password set
		DB:           c.DB,       // use default DB
		MaxRetries:   c.MaxRetries,
		DialTimeout:  c.DialTimeout,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
	})

	err := rdb.Set(context.Background(), key, value, 5*time.Minute).Err()
	if err != nil {
		log.Print(err)
	}
	return err
}
