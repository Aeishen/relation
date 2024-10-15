package relation_core

import (
	"context"
	"relation/dao/db_mongo"
	"relation/dao/db_redis"
	"relation/types"
	"relation/util"
	"relation/work_pool"
)

// 关注

func Follow(ctx context.Context, uid, toUid uint64, ctime int64) error {
	// 先写关注mongo
	isDuplicateKey, err := db_mongo.AddFollow(ctx, uid, toUid, ctime)
	if err != nil {
		return err
	}
	if isDuplicateKey {
		return nil
	}

	// 异步处理其他数据写入
	data := map[string]any{
		"follow_uid":  toUid,
		"create_time": ctime,
	}
	work_pool.SendToWorkQueue(work_pool.GoCtx(ctx), uid, relation_types.GoHandleFollowCmd, data)
	return nil
}

func UnFollow(ctx context.Context, uid, toUid uint64, addToAbnormal bool) error {
	// 先写关注mongo
	err := db_mongo.DelFollow(ctx, uid, util.SeqID(uid, toUid))
	if err != nil {
		if addToAbnormal {
			abnormalID := util.AbnormalID(uid, toUid, relation_types.RelationFollow, relation_types.StorageMongo, relation_types.OpDelRecord)
			db_mongo.AddAbnormal(ctx, abnormalID, uid, "", err.Error(), 0)
		} else {
			return err
		}
	}

	// 异步处理其他数据写入
	data := map[string]any{
		"follow_uid": toUid,
	}
	work_pool.SendToWorkQueue(work_pool.GoCtx(ctx), uid, relation_types.GoHandleUnFollowCmd, data)
	return nil
}

func FollowList(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]uint64, uint64, error) {
	searchCnt := pageSize + 1

	uids, err := db_mongo.GetFollowList(ctx, uid, startIndex, searchCnt)
	if len(uids) == int(searchCnt) {
		return uids[:len(uids)-1], uint64(startIndex + pageSize), nil
	}

	return uids, 0, err
}

func FollowCnt(ctx context.Context, uid uint64) (int32, error) {
	cnt := db_redis.GetRelateTotalCnt(ctx, uid, relation_types.RelationFollow)
	if cnt < 0 {
		var err error
		cnt, err = db_mongo.GetFollowCnt(ctx, uid)
		if err == nil {
			_ = db_redis.SetRelateTotalCnt(ctx, uid, relation_types.RelateTypeName[relation_types.RelationFollow], cnt)
		}
		return cnt, err
	}
	return cnt, nil
}

func IsFollow(ctx context.Context, uid, toUid uint64) (bool, error) {
	return db_mongo.GetIsFollow(ctx, uid, util.SeqID(uid, toUid))
}

func IsFollowBatch(ctx context.Context, uid uint64, toUids []uint64) (map[uint64]struct{}, error) {
	result := make(map[uint64]struct{})
	seqIDs := make([]string, 0, len(toUids))
	for _, toUid := range toUids {
		seqIDs = append(seqIDs, util.SeqID(uid, toUid))
	}

	list, err := db_mongo.GetIsFollowBatch(ctx, uid, seqIDs)
	if err != nil {
		return result, err
	}

	for _, seqID := range list {
		if seqID == "" {
			continue
		}
		_uid, _toUid := util.UnLashSeqID(seqID)
		if _uid == 0 || _toUid == 0 {
			continue
		}

		result[_toUid] = struct{}{}
	}

	return result, nil
}
