package relation

import (
	"context"
	"github.com/sirupsen/logrus"
	"relation/dao/db_mongo"
	"relation/dao/db_redis"
	"relation/util"
	"relation/work_pool"
)

func Init(ctx context.Context, conf *Config) error {
	logrus.Infof("relation core init config:{%+v}", conf)

	err := initConfig(conf)
	if err != nil {
		return err
	}

	db_redis.InitRedis(conf.DaoConf.RedisConf.RedisAddr, conf.DaoConf.RedisConf.RedisPassword, conf.AppConf.Name)
	db_mongo.InitMongo(conf.DaoConf.MongoConf.Mongo, conf.AppConf.Name)
	work_pool.Init()
	util.InitAlert(conf.AlertConf.AlertUrl, conf.AlertConf.NoticeUrl, conf.AppConf.Mode, conf.AlertConf.TestAlert)

	//go loopHandleAbnormal(ctx)

	logrus.Info("relation core init successful")
	return nil
}

func loopHandleAbnormal(ctx context.Context) {
	//t := time.NewTicker(2 * time.Second)
	//for {
	//	select {
	//	case <-t.C:
	//		abnormalList := db_mongo.GetAbnormal(context.Background())
	//		if len(abnormalList) <= 0 {
	//			continue
	//		}
	//		for _, cur := range abnormalList {
	//			// 异步处理其他数据写入
	//			data := map[string]any{
	//				"event_time":  cur.EventTime,
	//				"abnormal_id": cur.ID,
	//			}
	//			work_pool.SendToWorkQueue(work_pool.GoCtx(ctx), cur.Uid, relation_types.GoHandleAbnormalCmd, data)
	//		}
	//
	//	case <-ctx.Done():
	//		t.Stop()
	//		logrus.Info("loop context done")
	//		return
	//	}
	//}
}

// 数据异常时，可以通过nacos 刷数据，加锁即可，用户操作时返回异常
