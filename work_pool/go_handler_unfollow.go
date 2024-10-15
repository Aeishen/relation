package work_pool

import (
	"github.com/sirupsen/logrus"
	"relation/dao/db_mongo"
	"relation/dao/db_redis"
	"relation/types"
	"relation/util"
)

type unFollowHandler struct{}

func initUnFollowHandler() {
	l := unFollowHandler{}
	_pool.addCmdHandler(relation_types.GoHandleUnFollowCmd, l)
}

func (f unFollowHandler) Handle(d *GoWorkerData) {
	if d == nil {
		logrus.WithContext(d.Ctx).Error("[unFollowHandler.Handle] data is nil")
		return
	}
	logrus.WithContext(d.Ctx).Infof("[unFollowHandler.Handle] data:{%+v}", d)

	defer putGoWorkerData(d)

	followUid, ok := d.Data["follow_uid"].(uint64)
	if !ok {
		logrus.WithContext(d.Ctx).Error("[unFollowHandler.Handle] followUid is wrong")
		return
	}
	uid := uint64(d.ID)

	//// 减少 uid 关注数
	//if _, err := db_redis.UpRelateTotalCnt(d.Ctx, uid, -1, relation_types.RelationFollow); err != nil {
	//	abnormalID := util.AbnormalID(uid, followUid, relation_types.RelationFollow, relation_types.StorageRedis, relation_types.OpDescCnt)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "-1", err.Error(), 0)
	//}
	//
	//// 将 followUid 从 uid 的 redis 关注列表移除
	//if err := db_redis.DelFromList(d.Ctx, db_redis.KeyFollowList(uid), followUid); err != nil {
	//	abnormalID := util.AbnormalID(uid, followUid, relation_types.RelationFollow, relation_types.StorageRedis, relation_types.OpDelRecord)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "", err.Error(), 0)
	//}
	//
	// 将 uid 从 followUid 的 mongo 粉丝列表移除
	//if err := db_mongo.DelFans(d.Ctx, followUid, util.SeqID(followUid, uid)); err != nil {
	//	abnormalID := util.AbnormalID(followUid, uid, relation_types.RelationFans, relation_types.StorageMongo, relation_types.OpDelRecord)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "", err.Error(), 0)
	//}
	//
	//// 减少 followUid 粉丝数
	//if _, err := db_redis.UpRelateTotalCnt(d.Ctx, followUid, -1, relation_types.RelationFans); err != nil {
	//	abnormalID := util.AbnormalID(followUid, uid, relation_types.RelationFans, relation_types.StorageRedis, relation_types.OpDescCnt)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "-1", err.Error(), 0)
	//}
	//// 将 uid 从 followUid 的 redis 粉丝列表移除
	//if err := db_redis.DelFromList(d.Ctx, db_redis.KeyFansList(followUid), uid); err != nil {
	//	abnormalID := util.AbnormalID(followUid, uid, relation_types.RelationFans, relation_types.StorageRedis, relation_types.OpDelRecord)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "", err.Error(), 0)
	//}

	// 获取 mongo uid 关注数 并写入 redis
	followCnt, err := db_mongo.GetFollowCnt(d.Ctx, uid)
	if err != nil {
		db_redis.DelRelateTotalCnt(d.Ctx, uid)
	} else {
		_ = db_redis.SetRelateTotalCnt(d.Ctx, uid, relation_types.RelateTypeName[relation_types.RelationFollow], followCnt)
	}

	// 将 uid 从 followUid 的 mongo 粉丝列表移除
	if err = db_mongo.DelFans(d.Ctx, followUid, util.SeqID(followUid, uid)); err != nil {
		abnormalID := util.AbnormalID(followUid, uid, relation_types.RelationFans, relation_types.StorageMongo, relation_types.OpDelRecord)
		db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "", err.Error(), 0)
	} else {
		// 获取 mongo followUid 粉丝数 并写入 redis
		fansCnt, err := db_mongo.GetFansCnt(d.Ctx, followUid)
		if err != nil {
			db_redis.DelRelateTotalCnt(d.Ctx, followUid)
		} else {
			_ = db_redis.SetRelateTotalCnt(d.Ctx, followUid, relation_types.RelateTypeName[relation_types.RelationFans], fansCnt)
		}
	}

	// 移除新粉丝, 非核心业务，出错不处理
	_ = db_redis.DelFromList(d.Ctx, db_redis.KeyNewFansList(followUid), uid)
}
