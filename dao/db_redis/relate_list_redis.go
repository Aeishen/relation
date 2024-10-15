package db_redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gitlab-vywrajy.micoworld.net/yoho-go/yprometheus"
	"gitlab-vywrajy.micoworld.net/yoho-go/ytools/cast"
	"relation/types"
	"time"
)

// ----------------------------关系列表-------------------------------------

func KeyNewFansList(uid uint64) string {
	return fmt.Sprintf("%s_relate:new_fans_list:%d", app, uid)
}

func AddToList(ctx context.Context, key string, toUid uint64, t, ex int64) error {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := redisCli.ZAdd(timeCtx, key, &redis.Z{
		Score:  float64(t),
		Member: cast.ToString(toUid),
	}).Err()
	if err != nil {
		yprometheus.Inc("redis_AddList_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.AddList] key:{%v} toUid:{%v} error:{%v}", key, toUid, err)
	}

	go redisCli.Expire(context.Background(), key, time.Duration(ex)*time.Second)

	logrus.WithContext(ctx).Debugf("[db_redis.AddList] successful key:{%v} toUid:{%v}", key, toUid)
	return err
}

func DelFromList(ctx context.Context, key string, toUid uint64) error {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := redisCli.ZRem(timeCtx, key, cast.ToString(toUid)).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logrus.WithContext(ctx).Debugf("[db_redis.DelFromList] key:{%v} toUid:{%v} not found", key, toUid)
			return nil
		}
		yprometheus.Inc("redis_DelFromList_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.DelFromList] key:{%v} toUid:{%v} error:{%v}", key, toUid, err)
	}
	logrus.WithContext(ctx).Debugf("[db_redis.DelFromList] successful key:{%v} toUid:{%v}", key, toUid)
	return err
}

func DelList(ctx context.Context, key string) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := redisCli.Del(timeCtx, key).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logrus.WithContext(ctx).Debugf("[db_redis.DelList] key:{%v} not found", key)
			return
		}
		yprometheus.Inc("redis_DelList_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.DelList] key:{%v}  error:{%v}", key, err)
	}
	logrus.WithContext(ctx).Debugf("[db_redis.DelList] successful key:{%v}", key)
	return
}

func IsInList(ctx context.Context, key string, toUid uint64) int64 {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t, err := redisCli.ZScore(timeCtx, key, cast.ToString(toUid)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		yprometheus.Inc("redis_IsInList_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.IsInList] key:{%v} toUid:{%v} error:{%v}", key, toUid, err)
	}
	logrus.WithContext(ctx).Debugf("[db_redis.IsInList] key:{%v} t:{%v}", key, t)
	return int64(t)
}

func GetList(ctx context.Context, key string, startIndex, pageSize int64) []uint64 {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := redisCli.ZRevRange(timeCtx, key, startIndex, startIndex+pageSize-1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		yprometheus.Inc("redis_GetList_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.GetList] key:{%v} error:{%v}", key, err)
		return []uint64{}
	}

	uids := make([]uint64, 0, len(result))
	for _, u := range result {
		uids = append(uids, cast.ToUint64(u))
	}
	logrus.WithContext(ctx).Debugf("[db_redis.GetList] key:{%v} uids:{%v}", key, uids)
	return uids
}

func GetListWithTime(ctx context.Context, key string, startIndex, pageSize int64) []*relation_types.UidWithTime {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := redisCli.ZRevRangeWithScores(timeCtx, key, startIndex, startIndex+pageSize).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		yprometheus.Inc("redis_GetListWithTime_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.GetListWithTime] key:{%v} error:{%v}", key, err)
		return []*relation_types.UidWithTime{}
	}

	ret := make([]*relation_types.UidWithTime, 0, len(result))
	for _, member := range result {
		ret = append(ret, &relation_types.UidWithTime{
			Uid:        cast.ToUint64(member.Member),
			RelateTime: int64(member.Score),
		})
		logrus.WithContext(ctx).Debugf("[db_redis.GetListWithTime] key:{%v} member:{%v}", key, member)
	}

	return ret
}

func DelListRange(ctx context.Context, key string, st, en int64) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := redisCli.ZRemRangeByRank(timeCtx, key, st, en).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		yprometheus.Inc("redis_DelListRange_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.DelListRange] key:{%v} st:{%v}, en:{%v} error:{%v}", key, st, en, err)
	}
	logrus.WithContext(ctx).Debugf("[db_redis.DelListRange] key:{%v} successful", key)
	return
}

func GetListLen(ctx context.Context, key string) uint32 {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := redisCli.ZCard(timeCtx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		yprometheus.Inc("redis_GetListLen_Err", 0)
		logrus.WithContext(ctx).Errorf("[db_redis.GetListLen] key:{%v} error:{%v}", key, err)
		return 0
	}

	logrus.WithContext(ctx).Debugf("[db_redis.GetListLen] key:{%v} result:{%v}", key, result)
	return uint32(result)
}
