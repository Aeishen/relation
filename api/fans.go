package relation_core

import (
	"context"
	"relation/dao/db_mongo"
	"relation/dao/db_redis"
	"relation/util"
)

// 粉丝

func FansList(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]uint64, uint64, error) {
	defer func() {
		db_redis.DelList(ctx, db_redis.KeyNewFansList(uid))
	}()

	searchCnt := pageSize + 1

	uids, err := db_mongo.GetFansList(ctx, uid, startIndex, searchCnt)
	if len(uids) == int(searchCnt) {
		return uids[:len(uids)-1], uint64(startIndex + pageSize), nil
	}

	return uids, 0, err
}

func IsFansBatch(ctx context.Context, uid uint64, toUids []uint64) (map[uint64]struct{}, error) {
	result := make(map[uint64]struct{})
	seqIDs := make([]string, 0, len(toUids))
	for _, toUid := range toUids {
		seqIDs = append(seqIDs, util.SeqID(uid, toUid))
	}
	list, err := db_mongo.GetIsFansBatch(ctx, uid, seqIDs)
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
