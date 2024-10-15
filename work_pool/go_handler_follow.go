package work_pool

import (
	"github.com/sirupsen/logrus"
	"relation/dao/db_mongo"
	"relation/dao/db_redis"
	"relation/types"
	"relation/util"
)

type followHandler struct{}

func initFollowHandler() {
	l := followHandler{}
	_pool.addCmdHandler(relation_types.GoHandleFollowCmd, l)
}

func (f followHandler) Handle(d *GoWorkerData) {
	if d == nil {
		logrus.WithContext(d.Ctx).Error("[followHandler.Handle] data is nil")
		return
	}
	logrus.WithContext(d.Ctx).Infof("[followHandler.Handle] data:{%+v}", d)

	defer putGoWorkerData(d)

	followUid, ok := d.Data["follow_uid"].(uint64)
	if !ok {
		logrus.WithContext(d.Ctx).Error("[followHandler.Handle] followUid is wrong")
		return
	}
	cTime, ok := d.Data["create_time"].(int64)
	if !ok {
		logrus.WithContext(d.Ctx).Error("[followHandler.Handle] cTime is wrong")
		return
	}
	uid := uint64(d.ID)

	// 增加 uid 关注数
	//if _, err := db_redis.UpRelateTotalCnt(d.Ctx, uid, 1, relation_types.RelationFollow); err != nil {
	//	abnormalID := util.AbnormalID(uid, followUid, relation_types.RelationFollow, relation_types.StorageRedis, relation_types.OpIncrCnt)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "1", err.Error(), cTime)
	//} else if followCnt > relation_types.DefaultRedisCacheCnt {
	//	// uid 关注数超过redis设定限制，移除一半
	//	db_redis.DelListRange(d.Ctx, db_redis.KeyFollowList(uid), relation_types.DefaultRedisCacheCnt/2, -1)
	//}

	//// 将 followUid 添加到 uid 的 redis 关注列表
	//if err := db_redis.AddToList(d.Ctx, db_redis.KeyFollowList(uid), followUid, cTime); err != nil {
	//	abnormalID := util.AbnormalID(uid, followUid, relation_types.RelationFollow, relation_types.StorageRedis, relation_types.OpAddRecord)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "", err.Error(), cTime)
	//}

	// 增加 followUid 粉丝数
	//if _, err := db_redis.UpRelateTotalCnt(d.Ctx, followUid, 1, relation_types.RelationFans); err != nil {
	//	abnormalID := util.AbnormalID(followUid, uid, relation_types.RelationFans, relation_types.StorageRedis, relation_types.OpIncrCnt)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "1", err.Error(), cTime)
	//} else if fansCnt > relation_types.DefaultRedisCacheCnt {
	//	// followUid 粉丝数超过redis设定限制，移除一半
	//	db_redis.DelListRange(d.Ctx, db_redis.KeyFansList(followUid), relation_types.DefaultRedisCacheCnt/2, -1)
	//}

	//// 将 uid 添加到 followUid 的 redis 粉丝列表
	//if err := db_redis.AddToList(d.Ctx, db_redis.KeyFansList(followUid), uid, cTime); err != nil {
	//	abnormalID := util.AbnormalID(followUid, uid, relation_types.RelationFans, relation_types.StorageRedis, relation_types.OpAddRecord)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "", err.Error(), cTime)
	//}

	// 获取 mongo uid 关注数 并写入 redis
	followCnt, err := db_mongo.GetFollowCnt(d.Ctx, uid)
	if err != nil {
		db_redis.DelRelateTotalCnt(d.Ctx, uid)
	} else {
		_ = db_redis.SetRelateTotalCnt(d.Ctx, uid, relation_types.RelateTypeName[relation_types.RelationFollow], followCnt)
	}

	// 将 uid 添加到 followUid 的 mongo 粉丝列表
	if _, err = db_mongo.AddFans(d.Ctx, followUid, uid, cTime); err != nil {
		abnormalID := util.AbnormalID(followUid, uid, relation_types.RelationFans, relation_types.StorageMongo, relation_types.OpAddRecord)
		db_mongo.AddAbnormal(d.Ctx, abnormalID, followUid, "", err.Error(), cTime)
	} else {
		// 获取 mongo followUid 粉丝数 并写入 redis
		fansCnt, err := db_mongo.GetFansCnt(d.Ctx, followUid)
		if err != nil {
			db_redis.DelRelateTotalCnt(d.Ctx, followUid)
		} else {
			_ = db_redis.SetRelateTotalCnt(d.Ctx, followUid, relation_types.RelateTypeName[relation_types.RelationFans], fansCnt)
		}
	}

	// 增加新粉丝, 非核心业务，出错不处理
	_ = db_redis.AddToList(d.Ctx, db_redis.KeyNewFansList(followUid), uid, cTime, relation_types.RedisNewFansTTL)
}
