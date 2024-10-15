package work_pool

import (
	"github.com/sirupsen/logrus"
	"relation/dao/db_mongo"
	"relation/dao/db_redis"
	"relation/types"
	"relation/util"
)

type delFriendHandler struct{}

func initDelFriendHandler() {
	l := delFriendHandler{}
	_pool.addCmdHandler(relation_types.GoHandleDelFriendCmd, l)
}

func (f delFriendHandler) Handle(d *GoWorkerData) {
	if d == nil {
		logrus.WithContext(d.Ctx).Error("[delFriendHandler.Handle] data is nil")
		return
	}
	logrus.WithContext(d.Ctx).Infof("[delFriendHandler.Handle] data:{%+v}", d)

	defer putGoWorkerData(d)

	friendUid, ok := d.Data["friend_uid"].(uint64)
	if !ok {
		logrus.WithContext(d.Ctx).Error("[delFriendHandler.Handle] friendUid is wrong")
		return
	}
	uid := uint64(d.ID)

	// 获取 mongo uid 好友数 并写入 redis
	friendCnt, err := db_mongo.GetFriendCnt(d.Ctx, uid)
	if err != nil {
		db_redis.DelRelateTotalCnt(d.Ctx, uid)
	} else {
		_ = db_redis.SetRelateTotalCnt(d.Ctx, uid, relation_types.RelateTypeName[relation_types.RelationFriend], friendCnt)
	}

	// 将 uid 添加到 followUid 的 mongo 粉丝列表
	if err = db_mongo.DelFriend(d.Ctx, friendUid, util.SeqID(friendUid, uid)); err != nil {
		abnormalID := util.AbnormalID(friendUid, uid, relation_types.RelationFriend, relation_types.StorageMongo, relation_types.OpDelRecord)
		db_mongo.AddAbnormal(d.Ctx, abnormalID, friendUid, "", err.Error(), 0)
	} else {
		// 获取 mongo followUid 好友数 并写入 redis
		friendCnt, err = db_mongo.GetFriendCnt(d.Ctx, friendUid)
		if err != nil {
			db_redis.DelRelateTotalCnt(d.Ctx, friendUid)
		} else {
			_ = db_redis.SetRelateTotalCnt(d.Ctx, friendUid, relation_types.RelateTypeName[relation_types.RelationFriend], friendCnt)
		}
	}
}
