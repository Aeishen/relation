package relation_core

import (
	"context"
	"relation/dao/db_mongo"
	"relation/types"
	"relation/util"
	"time"
)

// 拉黑

func Block(ctx context.Context, uid, toUid uint64, addToAbnormal bool) error {
	now := time.Now().Unix()
	_, err := db_mongo.AddBlock(ctx, uid, toUid, now)
	if err != nil {
		if addToAbnormal {
			abnormalID := util.AbnormalID(uid, toUid, relation_types.RelationBlock, relation_types.StorageMongo, relation_types.OpAddRecord)
			db_mongo.AddAbnormal(ctx, abnormalID, uid, "", err.Error(), now)
		} else {
			return err
		}
	}
	return nil
}

func UnBlock(ctx context.Context, uid, toUid uint64, addToAbnormal bool) error {
	err := db_mongo.DelBlock(ctx, util.SeqID(uid, toUid))
	if err != nil {
		if addToAbnormal {
			abnormalID := util.AbnormalID(uid, toUid, relation_types.RelationBlock, relation_types.StorageMongo, relation_types.OpDelRecord)
			db_mongo.AddAbnormal(ctx, abnormalID, uid, "", err.Error(), 0)
		} else {
			return err
		}
	}
	return nil
}

func IsBlock(ctx context.Context, uid, toUid uint64) (bool, error) {
	return db_mongo.GetIsBlock(ctx, util.SeqID(uid, toUid))
}

func IsBlockBatch(ctx context.Context, uid uint64, toUids []uint64) (map[uint64]struct{}, error) {
	result := make(map[uint64]struct{})
	seqIDs := make([]string, 0, len(toUids))
	for _, toUid := range toUids {
		seqIDs = append(seqIDs, util.SeqID(uid, toUid))
	}

	list, err := db_mongo.GetIsBlockBatch(ctx, seqIDs)
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

func IsBeingBlockBatch(ctx context.Context, uid uint64, toUids []uint64) (map[uint64]struct{}, error) {
	result := make(map[uint64]struct{})
	seqIDs := make([]string, 0, len(toUids))
	for _, toUid := range toUids {
		seqIDs = append(seqIDs, util.SeqID(toUid, uid))
	}
	list, err := db_mongo.GetIsBlockBatch(ctx, seqIDs)
	if err != nil {
		return result, err
	}

	for _, seqID := range list {
		if seqID == "" {
			continue
		}
		_toUid, _uid := util.UnLashSeqID(seqID)
		if _uid == 0 || _toUid == 0 {
			continue
		}

		result[_toUid] = struct{}{}
	}

	return result, nil
}

func BlockList(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]uint64, uint64, error) {
	searchCnt := pageSize + 1

	uids, err := db_mongo.GetBlockList(ctx, uid, startIndex, searchCnt)
	if len(uids) == int(searchCnt) {
		return uids[:len(uids)-1], uint64(startIndex + pageSize), nil
	}

	return uids, 0, err
}
