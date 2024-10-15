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

func AddFriend(ctx context.Context, uid, toUid uint64, now int64) (bool, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	m := &relation_types.FriendMgoModel{
		SeqID:      util.SeqID(uid, toUid),
		Uid:        uid,
		FriendUid:  toUid,
		CreateTime: now,
	}

	coll := collFriend(m.Uid)
	_, err := coll.InsertOne(timeCtx, m)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			logrus.WithContext(ctx).Warnf("[db_mongo.AddFriend] data:{%+v} duplicateKey", m)
			return true, nil
		}
		logrus.WithContext(ctx).Errorf("[db_mongo.AddFriend] data:{%+v} error:{%v}", m, err)
		return false, err
	}

	return false, nil
}

func DelFriend(ctx context.Context, uid uint64, seqID string) error {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	coll := collFriend(uid)
	_, err := coll.DeleteOne(timeCtx, bson.M{"_id": seqID})
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.DelFriend] seqID:{%s} error:{%v}", seqID, err)
		return err
	}
	return nil
}

func GetFriendList(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]uint64, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	coll := collFriend(uid)
	var list []*relation_types.FriendMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetProjection(bson.M{"friend_uid": 1})
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip(startIndex)
	findOptions.SetSort(bson.M{"create_time": -1})

	cursor, err := coll.Find(timeCtx, bson.M{"uid": uid}, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFriendList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &list); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFriendList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	uids := make([]uint64, 0, len(list))
	for _, friend := range list {
		uids = append(uids, friend.FriendUid)
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.GetFriendList] uid:{%d} uids:{%v}", uid, uids)
	return uids, nil
}

func GetFriendListWithTime(ctx context.Context, uid uint64, startIndex, pageSize int64) ([]*relation_types.UidWithTime, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	coll := collFriend(uid)
	var list []*relation_types.FriendMgoModel

	findOptions := new(options.FindOptions)
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip(startIndex)
	findOptions.SetSort(bson.M{"create_time": -1})

	cursor, err := coll.Find(timeCtx, bson.M{"uid": uid}, findOptions)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFriendList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}
	defer cursor.Close(timeCtx)
	if err = cursor.All(timeCtx, &list); err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFriendList] uid:{%d} error:{%v}", uid, err)
		return nil, err
	}

	uids := make([]*relation_types.UidWithTime, 0, len(list))
	for _, Friend := range list {
		uids = append(uids, &relation_types.UidWithTime{
			Uid:        Friend.FriendUid,
			RelateTime: Friend.CreateTime,
		})
	}

	return uids, nil
}

func GetFriendCnt(ctx context.Context, uid uint64) (int32, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	coll := collFriend(uid)

	count, err := coll.CountDocuments(timeCtx, bson.M{"uid": uid})
	if err != nil {
		logrus.WithContext(ctx).Errorf("[db_mongo.GetFriendCnt] uid:{%d} error:{%v}", uid, err)
		return 0, err
	}
	logrus.WithContext(ctx).Debugf("[db_mongo.GetFriendCnt] uid:{%v} count{%v}", uid, count)
	return int32(count), nil
}
