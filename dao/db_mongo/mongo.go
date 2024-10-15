package db_mongo

import (
	"context"
	"fmt"
	"gitlab-vywrajy.micoworld.net/yoho-go/ydb/ymongo"
	"go.mongodb.org/mongo-driver/mongo"
	relation_types "relation/types"
)

var mgoCli *mongo.Database
var app string

func InitMongo(conf ymongo.MongoConfig, _app string) {
	mgoCli = ymongo.InitMongo(conf).Database("db_video_record")
	app = _app
}

func collFollow(uid uint64) *mongo.Collection {
	index := uid % 10
	return mgoCli.Collection(fmt.Sprintf("%s_follow_%d", app, index))
}

func collFans(uid uint64) *mongo.Collection {
	index := uid % 10
	return mgoCli.Collection(fmt.Sprintf("%s_fans_%d", app, index))
}

func collBlock() *mongo.Collection {
	return mgoCli.Collection(fmt.Sprintf("%s_block", app))
}

func collFriend(uid uint64) *mongo.Collection {
	index := uid % 5
	return mgoCli.Collection(fmt.Sprintf("%s_friend_%d", app, index))
}

func RelationCnt(ctx context.Context, uid uint64, relateTypes []uint8) map[uint8]int32 {
	data := make(map[uint8]int32)
	for _, relateTy := range relateTypes {
		data[relateTy] = RelationCntByType(ctx, uid, relateTy)
	}
	return data
}

func RelationCntByType(ctx context.Context, uid uint64, relateTy uint8) int32 {
	var cnt int32
	switch relateTy {
	case relation_types.RelationFollow:
		cnt, _ = GetFollowCnt(ctx, uid)
	case relation_types.RelationFans:
		cnt, _ = GetFansCnt(ctx, uid)
	case relation_types.RelationFriend:
		cnt, _ = GetFriendCnt(ctx, uid)
	}
	return cnt
}
