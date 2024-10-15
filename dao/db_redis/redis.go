package db_redis

import (
	"github.com/go-redis/redis/v8"
	"gitlab-vywrajy.micoworld.net/yoho-go/ydb/yredis"
)

var redisCli *redis.Client
var app string

func InitRedis(addr, password, _app string) {
	redisCli = yredis.NewRedisClient(
		yredis.WithAddr(addr),
		yredis.WithPassword(password),
	)
	app = _app
}
