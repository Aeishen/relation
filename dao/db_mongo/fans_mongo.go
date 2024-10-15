package db_mongo

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"relation/types"
	"relation/util"
	"time"
)

func AddFans(ctx context.Context, uid, fromUid uint64, now int64) (bool, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	m := &relation_types.FansMgoModel{
		SeqID:      util.SeqID(uid, fromUid),
		Uid:        uid,
		FansUid:    fromUid,
		CreateTime: now,
	}

	_, err := collFans(m.Uid).InsertOne(timeCtx, m)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			logrus.WithContext(ctx).Warnf("[db_mongo.AddFans] data:{%+v} duplicateKey", m)
			return true, nil
		}
		logrus.WithContext(ctx).Errorf("[db_mongo.AddFans] data:{%+v} error:{%v}", m, err)
		return false, err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.AddFans] data:{%+v} successful", m)
	return false, nil
}

func DelFans(ctx context.Context, uid uint64, seqID string) error {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := collFans(uid).DeleteOne(timeCtx, bson.M{"_id": seqID})
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.DelFans] seqID:{%s} error:{%v}", seqID, err)
		return err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.DelFans] seqID:{%v} successful", seqID)
	return nil
}

func GetIsFansBatch(ctx context.Context, uid uint64, seqIDs []string) ([]string, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var data []*relation_types.FansMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetProjection(bson.M{"_id": 1})

	filter := bson.M{"_id": bson.M{"$in": seqIDs}}

	cursor, err := collFans(uid).Find(timeCtx, filter, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetIsFansBatch] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &data); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetIsFansBatch] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	result := make([]string, 0, len(data))
	for _, datum := range data {
		if datum == nil {
			continue
		}
		result = append(result, datum.SeqID)
	}

	logrus.WithContext(ctx).Debugf("[db_mongo.GetIsFansBatch] seqIDs:{%+v} result:{%+v}", seqIDs, result)
	return result, nil
}

func GetFansList(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]uint64, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var list []*relation_types.FansMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetProjection(bson.M{"fans_uid": 1})
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip(startIndex)
	findOptions.SetSort(bson.M{"create_time": -1})

	cursor, err := collFans(uid).Find(timeCtx, bson.M{"uid": uid}, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFansList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &list); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFansList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	uids := make([]uint64, 0, len(list))
	for _, fans := range list {
		uids = append(uids, fans.FansUid)
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.GetFansList] uids:{%v}", uids)
	return uids, nil
}

func GetFansListWithTime(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]*relation_types.UidWithTime, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var list []*relation_types.FansMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip(startIndex)
	findOptions.SetSort(bson.M{"create_time": -1})

	cursor, err := collFans(uid).Find(timeCtx, bson.M{"uid": uid}, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFansList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &list); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFansList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	uids := make([]*relation_types.UidWithTime, 0, len(list))
	for _, fans := range list {
		uids = append(uids, &relation_types.UidWithTime{
			Uid:        fans.FansUid,
			RelateTime: fans.CreateTime,
		})
	}

	return uids, nil
}

func GetFansCnt(ctx context.Context, uid uint64) (int32, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	count, err := collFans(uid).CountDocuments(timeCtx, bson.M{"uid": uid})
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFansCnt] uid:{%d} error:{%v}", uid, err)
		return 0, err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.GetFansCnt] uid:{%v} count{%v}", uid, count)
	return int32(count), nil
}
