package work_pool

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab-vywrajy.micoworld.net/yoho-go/ytools/cast"
	"relation/dao/db_mongo"
	"relation/types"
	"relation/util"
	"time"
)

type abnormalHandler struct{}

func initAbnormalHandler() {
	l := abnormalHandler{}
	_pool.addCmdHandler(relation_types.GoHandleAbnormalCmd, l)
}

func (r abnormalHandler) Handle(d *GoWorkerData) {
	if d == nil {
		logrus.WithContext(d.Ctx).Error("[abnormalHandler.Handle] data is nil")
		return
	}
	logrus.WithContext(d.Ctx).Infof("[abnormalHandler.Handle] data:{%+v}", d)

	defer putGoWorkerData(d)

	abnormalID, ok := d.Data["abnormal_id"].(string)
	if !ok {
		logrus.WithContext(d.Ctx).Error("[abnormalHandler.Handle] abnormalID is wrong")
		return
	}

	happenTime, ok := d.Data["event_time"].(int64)
	if !ok {
		happenTime = time.Now().Unix()
		logrus.WithContext(d.Ctx).Error("[abnormalHandler.Handle] abnormalID is wrong")
	}

	abnormalData := util.ParseAbnormalID(abnormalID)
	if len(abnormalData) < 5 {
		logrus.WithContext(d.Ctx).Error("[abnormalHandler.Handle] abnormalID is wrong")
		return
	}

	uid := cast.ToUint64(abnormalData[0])
	toUid := cast.ToUint64(abnormalData[1])
	relation := cast.ToUint8(abnormalData[2])
	storage := cast.ToUint8(abnormalData[3])
	op := cast.ToUint8(abnormalData[4])
	ctx := context.WithValue(d.Ctx, "t", fmt.Sprintf("recover_%s", abnormalID))

	// 先设置中间态
	db_mongo.UpAbnormal(ctx, abnormalID, relation_types.AbnormalStDuring)

	// 处理每一条异常
	var err error
	switch relation {
	case relation_types.RelationFollow:
		err = recoverFollow(ctx, uid, toUid, storage, op, happenTime)
	case relation_types.RelationFans:
		err = recoverFans(ctx, uid, toUid, storage, op, happenTime)
	case relation_types.RelationFriend:
		err = recoverFriends(ctx, uid, toUid, storage, op, happenTime)
	case relation_types.RelationBlock:
		err = recoverBlock(ctx, uid, toUid, storage, op, happenTime)
	}

	if err != nil {
		// 处理失败恢复待处理状态
		db_mongo.UpAbnormal(ctx, abnormalID, relation_types.AbnormalStReady)
	} else {
		// 处理成功修改为处理成功状态
		db_mongo.UpAbnormal(ctx, abnormalID, relation_types.AbnormalStDone)
	}
}

func recoverFollow(ctx context.Context, uid, toUid uint64, storage, op uint8, t int64) error {
	var err error
	switch storage {
	case relation_types.StorageMongo:
		switch op {
		case relation_types.OpAddRecord:
			_, err = db_mongo.AddFollow(ctx, uid, toUid, t)
		case relation_types.OpDelRecord:
			err = db_mongo.DelFollow(ctx, uid, util.SeqID(uid, toUid))
		}
	case relation_types.StorageRedis:
		//switch op {
		//case relation_types.OpIncrCnt:
		//	_, err = db_redis.UpRelateTotalCnt(ctx, uid, 1, relation_types.RelationFollow)
		//case relation_types.OpDescCnt:
		//	_, err = db_redis.UpRelateTotalCnt(ctx, uid, -1, relation_types.RelationFollow)
		//}
	}
	return err
}

func recoverFans(ctx context.Context, uid, toUid uint64, storage, op uint8, t int64) error {
	var err error
	switch storage {
	case relation_types.StorageMongo:
		switch op {
		case relation_types.OpAddRecord:
			_, err = db_mongo.AddFans(ctx, uid, toUid, t)
		case relation_types.OpDelRecord:
			err = db_mongo.DelFans(ctx, uid, util.SeqID(uid, toUid))
		}
	case relation_types.StorageRedis:
		//switch op {
		//case relation_types.OpIncrCnt:
		//	_, err = db_redis.UpRelateTotalCnt(ctx, uid, 1, relation_types.RelationFans)
		//case relation_types.OpDescCnt:
		//	_, err = db_redis.UpRelateTotalCnt(ctx, uid, -1, relation_types.RelationFans)
		//}
	}
	return err
}

func recoverFriends(ctx context.Context, uid, toUid uint64, storage, op uint8, t int64) error {
	var err error
	switch storage {
	case relation_types.StorageMongo:
		switch op {
		case relation_types.OpAddRecord:
			_, err = db_mongo.AddFriend(ctx, uid, toUid, t)
		case relation_types.OpDelRecord:
			err = db_mongo.DelFriend(ctx, uid, util.SeqID(uid, toUid))
		}
	case relation_types.StorageRedis:
		//switch op {
		//case relation_types.OpIncrCnt:
		//	_, err = db_redis.UpRelateTotalCnt(ctx, uid, 1, relation_types.RelationFriend)
		//case relation_types.OpDescCnt:
		//	_, err = db_redis.UpRelateTotalCnt(ctx, uid, -1, relation_types.RelationFriend)
		//}
	}
	return err
}

func recoverBlock(ctx context.Context, uid, toUid uint64, storage, op uint8, t int64) error {
	var err error
	switch storage {
	case relation_types.StorageMongo:
		switch op {
		case relation_types.OpAddRecord:
			_, err = db_mongo.AddBlock(ctx, uid, toUid, t)
		case relation_types.OpDelRecord:
			err = db_mongo.DelBlock(ctx, util.SeqID(uid, toUid))
		}
	}
	return err
}
