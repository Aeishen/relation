package relation_core

import (
	"context"
	"relation/dao/db_mongo"
	"relation/types"
	"relation/util"
	"relation/work_pool"
)

// 好友

func ApplyFriend(ctx context.Context, uid, toUid uint64) error {
	return nil
}

func AgreeFriend(ctx context.Context, uid, toUid uint64) error {
	return nil
}

func RejectFriend(ctx context.Context, uid, toUid uint64) error {
	return nil
}

func ApplyFriendList(ctx context.Context, uid uint64, pageNum, pageSize, sortWay int) ([]uint64, error) {
	return nil, nil
}

func AddFriend(ctx context.Context, uid, toUid uint64, ctime int64, addToAbnormal bool) error {
	// 先写mongo
	isDuplicateKey, err := db_mongo.AddFriend(ctx, uid, toUid, ctime)
	if err != nil {
		if addToAbnormal {
			abnormalID := util.AbnormalID(uid, toUid, relation_types.RelationFriend, relation_types.StorageMongo, relation_types.OpAddRecord)
			db_mongo.AddAbnormal(ctx, abnormalID, uid, "", err.Error(), ctime)
		} else {
			return err
		}
	}
	if isDuplicateKey {
		return nil
	}

	// 异步处理其他数据写入
	data := map[string]any{
		"friend_uid":  toUid,
		"create_time": ctime,
	}
	work_pool.SendToWorkQueue(work_pool.GoCtx(ctx), uid, relation_types.GoHandleFriendCmd, data)
	return nil
}

func DelFriend(ctx context.Context, uid, toUid uint64, addToAbnormal bool) error {
	// 先写mongo
	err := db_mongo.DelFriend(ctx, uid, util.SeqID(uid, toUid))
	if err != nil {
		if addToAbnormal {
			abnormalID := util.AbnormalID(uid, toUid, relation_types.RelationFriend, relation_types.StorageMongo, relation_types.OpDelRecord)
			db_mongo.AddAbnormal(ctx, abnormalID, uid, "", err.Error(), 0)
		} else {
			return err
		}
	}

	// 异步处理其他数据写入
	data := map[string]any{
		"friend_uid": toUid,
	}
	work_pool.SendToWorkQueue(work_pool.GoCtx(ctx), uid, relation_types.GoHandleDelFriendCmd, data)
	return nil
}

func FriendList(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]uint64, uint64, error) {
	searchCnt := pageSize + 1
	// 读mongo
	uids, err := db_mongo.GetFriendList(ctx, uid, startIndex, searchCnt)
	if len(uids) == int(searchCnt) {
		return uids[:len(uids)-1], uint64(startIndex + pageSize), nil
	}

	return uids, 0, err
}
