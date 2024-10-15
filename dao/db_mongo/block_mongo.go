package db_mongo

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"relation/types"
	"relation/util"
	"time"
)

func AddBlock(ctx context.Context, uid, blockUid uint64, now int64) (bool, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	m := &relation_types.BlockMgoModel{
		SeqID:      util.SeqID(uid, blockUid),
		Uid:        uid,
		BlockUid:   blockUid,
		CreateTime: now,
	}

	_, err := collBlock().InsertOne(timeCtx, m)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		if mongo.IsDuplicateKeyError(err) {
			logrus.WithContext(ctx).Warnf("[db_mongo.AddBlock] data:{%+v} duplicateKey", m)
			return true, nil
		}
		logrus.WithContext(ctx).Errorf("[db_mongo.AddBlock] data:{%+v} error:{%v}", m, err)
		return false, err
	}

	return false, nil
}

func DelBlock(ctx context.Context, seqID string) error {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := collBlock().DeleteOne(timeCtx, bson.M{"_id": seqID})
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.DelBlock] seqID:{%s} error:{%v}", seqID, err)
	}
	return err
}

func GetIsBlock(ctx context.Context, seqID string) (bool, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result relation_types.BlockMgoModel
	err := collBlock().FindOne(timeCtx, bson.M{"_id": seqID}, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&result)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetIsBlock] seqID:{%+v} error:{%v}", seqID, err)
		return false, err
	}
	return result.SeqID == seqID, nil
}

func GetIsBlockBatch(ctx context.Context, seqIDs []string) ([]string, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var data []*relation_types.BlockMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetProjection(bson.M{"_id": 1})

	filter := bson.M{"_id": bson.M{"$in": seqIDs}}

	cursor, err := collBlock().Find(timeCtx, filter, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetIsBlockBatch] error:{%v}", err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &data); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetIsBlockBatch]  error:{%v}", err)
		return nil, err
	}

	result := make([]string, 0, len(data))
	for _, datum := range data {
		result = append(result, datum.SeqID)
	}

	logrus.WithContext(ctx).Debugf("[db_mongo.GetIsBlockBatch] seqIDs:{%+v} result:{%+v}", seqIDs, result)
	return result, nil
}

func GetBlockList(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]uint64, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var list []*relation_types.BlockMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetProjection(bson.M{"block_uid": 1})
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip(startIndex)
	findOptions.SetSort(bson.M{"create_time": -1})

	cursor, err := collBlock().Find(timeCtx, bson.M{"uid": uid}, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetBlockList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &list); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetBlockList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	uids := make([]uint64, 0, len(list))
	for _, block := range list {
		uids = append(uids, block.BlockUid)
	}
	return uids, nil
}

func GetBlockListWithTime(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]*relation_types.UidWithTime, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var list []*relation_types.BlockMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip(startIndex)
	findOptions.SetSort(bson.M{"create_time": -1})

	cursor, err := collBlock().Find(timeCtx, bson.M{"uid": uid}, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetBlockListWithTime] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &list); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetBlockListWithTime] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	uids := make([]*relation_types.UidWithTime, 0, len(list))
	for _, block := range list {
		uids = append(uids, &relation_types.UidWithTime{
			Uid:        block.BlockUid,
			RelateTime: block.CreateTime,
		})
	}

	return uids, nil
}

func GetBlockCnt(ctx context.Context, uid uint64) (uint32, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	count, err := collBlock().CountDocuments(timeCtx, bson.M{"uid": uid})
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetBlockCnt] uid:{%d} error:{%v}", uid, err)
		return 0, err
	}
	return uint32(count), nil
}
