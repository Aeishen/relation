package relation_types

const (
	DefaultGoWorkerCnt = 100

	DefaultRedisCacheCnt = 1000
	DefaultLocalCacheCnt = 500

	RedisNewFansTTL = 60 * 86400
	RedisCntTTL = 60
)

const (
	GoHandleAbnormalCmd  = 1
	GoHandleFollowCmd    = 2
	GoHandleUnFollowCmd  = 3
	GoHandleFriendCmd    = 4
	GoHandleDelFriendCmd = 5
	GoHandleBlockCmd     = 6
	GoHandleUnBlockCmd   = 7
)

const (
	AbnormalStReady  = 0 // 待处理
	AbnormalStDuring = 1 // 处理中，也可能是处理成功或失败
	AbnormalStDone   = 2 // 处理成功

	RelationFollow = 1
	RelationFans   = 2
	RelationFriend = 3
	RelationBlock  = 4

	StorageMysql = 1
	StorageMongo = 2
	StorageRedis = 3

	OpAddRecord = 1
	OpDelRecord = 2
	OpIncrCnt   = 3
	OpDescCnt   = 4
)
