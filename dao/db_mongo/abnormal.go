package db_mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab-vywrajy.micoworld.net/yoho-go/yprometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"relation/types"
	"relation/util"
	"time"
)

var abnormalLog *logrus.Logger

func init() {
	abnormalLog = util.GetDiyLog("log/abnormal.log", true)
}

func collAbnormal() *mongo.Collection {
	return mgoCli.Collection(fmt.Sprintf("%s_relation_abnormal", app))
}

func AddAbnormal(ctx context.Context, id string, uid uint64, data, info string, t int64) {
	yprometheus.Inc("mongo_AddAbnormal", 0)
	m := &relation_types.AbnormalMgoModel{
		ID:         id,
		Uid:        uid,
		Data:       data,
		Info:       info,
		Status:     relation_types.AbnormalStReady,
		EventTime:  t,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	util.Alert(fmt.Sprintf("data:{%+v}", m), "Mongo.AddAbnormal")

	msg, _ := json.Marshal(m)

	abnormalLog.WithContext(ctx).Warnf("record%v", string(msg))

	//timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//defer cancel()
	//
	//coll := collAbnormal()
	//_, err := coll.InsertOne(timeCtx, m)
	//if err != nil {
	//	yprometheus.Inc("mongo_AddAbnormal_Err", 0)
	//	util.Alert(fmt.Sprintf("data:{%s} add error:{%v}", id, err.Error()), "Mongo.AddAbnormal")
	//	abnormalLog.WithContext(ctx).Errorf("[mongo.AddAbnormal] record:{%+v} error:{%v}", m, err.Error())
	//}
}

func GetAbnormal(ctx context.Context) []*relation_types.AbnormalMgoModel {
	yprometheus.Inc("mongo_GetAbnormal", 0)
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	findOptions := new(options.FindOptions)
	findOptions.SetLimit(100)
	findOptions.SetSkip(0)
	findOptions.SetSort(bson.M{"create_time": 1})

	coll := collAbnormal()
	cursor, err := coll.Find(timeCtx, bson.M{"status": relation_types.AbnormalStReady}, findOptions)
	if err != nil {
		yprometheus.Inc("mongo_GetAbnormal_Err", 0)
		util.Alert(fmt.Sprintf("get error:{%v}", err.Error()), "Mongo.GetAbnormal")
		abnormalLog.WithContext(ctx).Errorf("[mongo.GetAbnormal] get error:{%v}", err.Error())
		return nil
	}
	defer cursor.Close(timeCtx)

	var list []*relation_types.AbnormalMgoModel
	if err = cursor.All(timeCtx, &list); err != nil {
		util.Alert(fmt.Sprintf("cursor error:{%v}", err.Error()), "Mongo.GetAbnormal")
		abnormalLog.WithContext(ctx).Errorf("[mongo.GetAbnormal] cursor error:{%v}", err.Error())
		return nil
	}
	return list
}

func UpAbnormal(ctx context.Context, id string, st uint8) {
	abnormalLog.WithContext(ctx).Infof("[mongo.AddAbnormal] id:{%s} st:{%d}", id, st)
	yprometheus.Inc("mongo_UpAbnormal", 0)
	timeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	coll := collAbnormal()
	update := bson.D{
		{"$set", bson.D{
			{"status", st},
			{"update_time", time.Now()},
		}},
	}
	_, err := coll.UpdateOne(timeCtx, bson.M{"_id": id}, update)
	if err != nil {
		yprometheus.Inc("mongo_UpAbnormal_Err", 0)
		util.Alert(fmt.Sprintf("data:{%s} del error:{%v}", id, err.Error()), "Mongo.UpAbnormal")
		abnormalLog.WithContext(ctx).Errorf("[mongo.UpAbnormal] id:{%s} st:{%d} error:{%v}", id, st, err.Error())
	}
}
