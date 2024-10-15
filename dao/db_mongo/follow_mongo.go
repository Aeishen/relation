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

func AddFollow(ctx context.Context, uid, followUid uint64, now int64) (bool, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	m := &relation_types.FollowMgoModel{
		SeqID:      util.SeqID(uid, followUid),
		Uid:        uid,
		FollowUid:  followUid,
		CreateTime: now,
	}

	_, err := collFollow(m.Uid).InsertOne(timeCtx, m)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			logrus.WithContext(ctx).Warnf("[db_mongo.AddFollow] data:{%+v} duplicateKey", m)
			return true, nil
		}
		logrus.WithContext(ctx).Errorf("[db_mongo.AddFollow] data:{%+v} error:{%v}", m, err)
		return false, err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.AddFollow] data:{%+v} successful", m)
	return false, nil
}

func DelFollow(ctx context.Context, uid uint64, seqID string) error {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := collFollow(uid).DeleteOne(timeCtx, bson.M{"_id": seqID})
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.DelFollow] seqID:{%s} error:{%v}", seqID, err)
		return err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.DelFollow] seqID:{%v} successful", seqID)
	return nil
}

func GetFollowTime(ctx context.Context, uid uint64, seqID string) (int64, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result relation_types.FollowMgoModel
	err := collFollow(uid).FindOne(timeCtx, bson.M{"_id": seqID}, options.FindOne().SetProjection(bson.M{"create_time": 1})).Decode(&result)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFollowTime] seqID:{%+v} error:{%v}", seqID, err)
		return 0, err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.GetFollowTime] seqID:{%+v} result:{%+v}", seqID, result)
	return result.CreateTime, nil
}

func GetIsFollowBatch(ctx context.Context, uid uint64, seqIDs []string) ([]string, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var data []*relation_types.FollowMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetProjection(bson.M{"_id": 1})

	filter := bson.M{"_id": bson.M{"$in": seqIDs}}

	cursor, err := collFollow(uid).Find(timeCtx, filter, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetIsFollowBatch] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &data); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetIsFollowBatch] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	result := make([]string, 0, len(data))
	for _, datum := range data {
		if datum == nil {
			continue
		}
		result = append(result, datum.SeqID)
	}

	logrus.WithContext(ctx).Debugf("[db_mongo.GetIsFollowBatch] seqIDs:{%+v} result:{%+v}", seqIDs, data)
	return result, nil
}

func GetIsFollow(ctx context.Context, uid uint64, seqID string) (bool, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result relation_types.FollowMgoModel
	err := collFollow(uid).FindOne(timeCtx, bson.M{"_id": seqID}, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&result)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetIsFollow] seqID:{%+v} error:{%v}", seqID, err)
		return false, err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.GetIsFollow] seqID:{%+v} result:{%+v}", seqID, result)
	return result.SeqID == seqID, nil
}

func GetFollowList(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]uint64, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var list []*relation_types.FollowMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetProjection(bson.M{"follow_uid": 1})
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip(startIndex)
	findOptions.SetSort(bson.M{"create_time": -1})

	cursor, err := collFollow(uid).Find(timeCtx, bson.M{"uid": uid}, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFollowList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &list); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFollowList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	uids := make([]uint64, 0, len(list))
	for _, follow := range list {
		uids = append(uids, follow.FollowUid)
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.GetFollowList] uids:{%v}", uids)
	return uids, nil
}

func GetFollowListWithTime(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]*relation_types.UidWithTime, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var list []*relation_types.FollowMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip(startIndex)
	findOptions.SetSort(bson.M{"create_time": -1})

	cursor, err := collFollow(uid).Find(timeCtx, bson.M{"uid": uid}, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFollowList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &list); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFollowList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	uids := make([]*relation_types.UidWithTime, 0, len(list))
	for _, follow := range list {
		uids = append(uids, &relation_types.UidWithTime{
			Uid:        follow.FollowUid,
			RelateTime: follow.CreateTime,
		})
	}

	return uids, nil
}

func GetFollowCnt(ctx context.Context, uid uint64) (int32, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	count, err := collFollow(uid).CountDocuments(timeCtx, bson.M{"uid": uid})
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFollowCnt] uid:{%d} error:{%v}", uid, err)
		return 0, err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.GetFollowCnt] uid:{%v} count{%v}", uid, count)
	return int32(count), nil
}
