package relation

import (
	"errors"
	"gitlab-vywrajy.micoworld.net/yoho-go/ydb/ymongo"
)

type DaoConfig struct {
	RedisConf *RedisConfig
	MongoConf *MongoConfig
	//CacheConf *CacheConfig
}

type RedisConfig struct {
	RedisAddr     string
	RedisPassword string
	CacheCnt      int
}

type MongoConfig struct {
	Mongo ymongo.MongoConfig
}

type AlertConfig struct {
	AlertUrl  string
	NoticeUrl string
	TestAlert bool
}

type AppConfig struct {
	Name string
	Mode string
}

type Config struct {
	AppConf   *AppConfig
	AlertConf *AlertConfig
	DaoConf   *DaoConfig
}

var c *Config

func initConfig(conf *Config) error {
	if conf == nil {
		return errors.New("init relation core config is nil")
	}
	if conf.DaoConf == nil || conf.DaoConf.RedisConf == nil || conf.DaoConf.MongoConf == nil {
		return errors.New("init relation core dao config is wrong")
	}
	if conf.AppConf == nil || conf.AppConf.Mode == "" || conf.AppConf.Name == "" {
		return errors.New("init relation core app config is wrong")
	}

	c = conf
	if c.AlertConf == nil {
		c.AlertConf = &AlertConfig{
			AlertUrl:  "todo",
			NoticeUrl: "todo",
			TestAlert: true,
		}
	}
	if c.AlertConf.AlertUrl == "" {
		c.AlertConf.AlertUrl = "todo"
	}

	return nil
}
