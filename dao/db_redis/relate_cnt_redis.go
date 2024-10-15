package db_redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gitlab-vywrajy.micoworld.net/yoho-go/yprometheus"
	"gitlab-vywrajy.micoworld.net/yoho-go/ytools/cast"
	relation_types "relation/types"
	"time"
)

// ---------------------------- 每个用户对应关系的用户数（粉丝数，关注数，好友数）-------------------------------------
func keyRelateTotalCnt(uid uint64) string {
	return fmt.Sprintf("%s_relate:total_cnt:%d", app, uid)
}

func DelRelateTotalCnt(ctx context.Context, uid uint64) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := keyRelateTotalCnt(uid)
	err := redisCli.Del(timeCtx, key).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		yprometheus.Inc("redis_DelRelateTotalCnt_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.DelRelateTotalCnt] key:{%v}, error:{%v}", key, err.Error())
		return
	}
	logrus.WithContext(ctx).Debugf("[db_redis.DelRelateTotalCnt] key:{%v}", key)
	return
}

func SetRelateTotalCnt(ctx context.Context, uid uint64, relateTyName string, cnt int32) error {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := keyRelateTotalCnt(uid)

	err := redisCli.HSet(timeCtx, key, map[string]interface{}{relateTyName: cnt}).Err()
	if err != nil {
		yprometheus.Inc("redis_SetRelateTotalCnt_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.SetRelateTotalCnt] key:{%v}, relateTyName:{%v}, error:{%v}", key, relateTyName, err.Error())
		return err
	}
	logrus.WithContext(ctx).Debugf("[db_redis.SetRelateTotalCnt] key:{%v} relateTyName:{%v} cnt:{%v}", key, relateTyName, cnt)

	go redisCli.Expire(context.Background(), key, time.Duration(relation_types.RedisCntTTL)*time.Second)
	return nil
}

func MSetRelateTotalCnt(ctx context.Context, uid uint64, data map[string]int32) error {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := keyRelateTotalCnt(uid)
	var values []interface{}
	for s, i := range data {
		values = append(values, s, i)
	}

	err := redisCli.HMSet(timeCtx, key, values).Err()
	if err != nil {
		yprometheus.Inc("redis_MSetRelateTotalCnt_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.MSetRelateTotalCnt] key:{%v}, data{%v}, error:{%v}", key, data, err.Error())
		return err
	}
	logrus.WithContext(ctx).Debugf("[db_redis.MSetRelateTotalCnt] key{%v} data{%v}", key, data)

	go redisCli.Expire(context.Background(), key, time.Duration(relation_types.RedisCntTTL)*time.Second)
	return nil
}

func GetRelateTotalCnt(ctx context.Context, uid uint64, relateType uint8) int32 {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := keyRelateTotalCnt(uid)
	cnt, err := redisCli.HGet(timeCtx, key, relation_types.RelateTypeName[relateType]).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logrus.WithContext(ctx).Debugf("[db_redis.GetRelateTotalCnt] key:{%v} relateType:{%v} not found", key, relateType)
			return -1
		}
		yprometheus.Inc("redis_GetRelateTotalCnt_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.GetRelateTotalCnt] key:{%v}, relateType:{%v}, error:{%v}", key, relateType, err.Error())
		return -1
	}
	logrus.WithContext(ctx).Debugf("[db_redis.GetRelateTotalCnt] key:{%v} relateType:{%v} cnt:{%v}", key, relateType, cnt)

	if cnt == "" {
		return -1
	}

	return cast.ToInt32(cnt)
}

func MGetRelateTotalCnt(ctx context.Context, uid uint64, relateTypes []uint8) []interface{} {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := keyRelateTotalCnt(uid)

	fields := make([]string, 0, len(relateTypes))
	for _, u := range relateTypes {
		fields = append(fields, relation_types.RelateTypeName[u])
	}

	result, err := redisCli.HMGet(timeCtx, key, fields...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logrus.WithContext(ctx).Debugf("[db_redis.MGetRelateTotalCnt] key:{%v} relateTypes{%v} not found result:{%v}", key, relateTypes, result)
			return nil
		}
		yprometheus.Inc("redis_MGetRelateTotalCnt_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.MGetRelateTotalCnt] key:{%v}, relateTypes{%v}, error:{%v}", key, relateTypes, err.Error())
		return nil
	}

	logrus.WithContext(ctx).Debugf("[db_redis.MGetRelateTotalCnt] key{%v} relateTypes{%v}, result:{%v}", key, relateTypes, result)

	return result
}
