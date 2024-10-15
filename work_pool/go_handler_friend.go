package work_pool

import (
	"github.com/sirupsen/logrus"
	"relation/dao/db_mongo"
	"relation/dao/db_redis"
	"relation/types"
	"relation/util"
)

type friendHandler struct{}

func initFriendHandler() {
	l := friendHandler{}
	_pool.addCmdHandler(relation_types.GoHandleFriendCmd, l)
}

func (f friendHandler) Handle(d *GoWorkerData) {
	if d == nil {
		logrus.WithContext(d.Ctx).Error("[friendHandler.Handle] data is nil")
		return
	}
	logrus.WithContext(d.Ctx).Infof("[friendHandler.Handle] data:{%+v}", d)

	defer putGoWorkerData(d)

	friendUid, ok := d.Data["friend_uid"].(uint64)
	if !ok {
		logrus.WithContext(d.Ctx).Error("[friendHandler.Handle] friendUid is wrong")
		return
	}
	cTime, ok := d.Data["create_time"].(int64)
	if !ok {
		logrus.WithContext(d.Ctx).Error("[friendHandler.Handle] cTime is wrong")
		return
	}
	uid := uint64(d.ID)

	//// 增加 uid 好友数
	//if _, err := db_redis.UpRelateTotalCnt(d.Ctx, uid, 1, relation_types.RelationFriend); err != nil {
	//	abnormalID := util.AbnormalID(uid, friendUid, relation_types.RelationFriend, relation_types.StorageRedis, relation_types.OpIncrCnt)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "1", err.Error(), cTime)
	//}
	//
	//// 将 uid 添加到 friendUid 的 mongo 好友列表
	//if _, err := db_mongo.AddFriend(d.Ctx, friendUid, uid, cTime); err != nil {
	//	abnormalID := util.AbnormalID(friendUid, uid, relation_types.RelationFriend, relation_types.StorageMongo, relation_types.OpAddRecord)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "", err.Error(), cTime)
	//}
	//
	//// 增加 friendUid 好友数
	//if _, err := db_redis.UpRelateTotalCnt(d.Ctx, friendUid, 1, relation_types.RelationFriend); err != nil {
	//	abnormalID := util.AbnormalID(friendUid, uid, relation_types.RelationFriend, relation_types.StorageRedis, relation_types.OpIncrCnt)
	//	db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "1", err.Error(), cTime)
	//}

	// 获取 mongo uid 好友数 并写入 redis
	friendCnt, err := db_mongo.GetFriendCnt(d.Ctx, uid)
	if err != nil {
		db_redis.DelRelateTotalCnt(d.Ctx, uid)
	} else {
		_ = db_redis.SetRelateTotalCnt(d.Ctx, uid, relation_types.RelateTypeName[relation_types.RelationFriend], friendCnt)
	}

	// 将 uid 添加到 friendUid 的 mongo 好友列表
	if _, err := db_mongo.AddFriend(d.Ctx, friendUid, uid, cTime); err != nil {
		abnormalID := util.AbnormalID(friendUid, uid, relation_types.RelationFriend, relation_types.StorageMongo, relation_types.OpAddRecord)
		db_mongo.AddAbnormal(d.Ctx, abnormalID, uid, "", err.Error(), cTime)
	} else {
		// 获取 mongo friendUid 好友数 并写入 redis
		friendCnt, err = db_mongo.GetFriendCnt(d.Ctx, friendUid)
		if err != nil {
			db_redis.DelRelateTotalCnt(d.Ctx, friendUid)
		} else {
			_ = db_redis.SetRelateTotalCnt(d.Ctx, friendUid, relation_types.RelateTypeName[relation_types.RelationFriend], friendCnt)
		}
	}
}
