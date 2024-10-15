package relation_core

import (
	"context"
	"gitlab-vywrajy.micoworld.net/yoho-go/ytools/cast"
	"relation/dao/db_mongo"
	"relation/dao/db_redis"
	"relation/types"
)

func RelationCnt(ctx context.Context, uid uint64, relateTypes []uint8) map[uint8]*relation_types.TypeCnt {
	result := make(map[uint8]*relation_types.TypeCnt)
	tyCntMap := make(map[uint8]int32)
	tyNameCntMap := make(map[string]int32)

	redisData := db_redis.MGetRelateTotalCnt(ctx, uid, relateTypes)
	if redisData == nil {
		// 读取mongo
		tyCntMap = db_mongo.RelationCnt(ctx, uid, relateTypes)
	} else {
		for i, cnt := range redisData {
			if len(relateTypes) <= i {
				break
			}

			ty := relateTypes[i]

			if cnt == nil {
				tyCntMap[ty] = db_mongo.RelationCntByType(ctx, uid, ty)

			} else {
				tyCntMap[ty] = cast.ToInt32(cnt)
			}
		}
	}

	for ty, cnt := range tyCntMap {
		result[ty] = &relation_types.TypeCnt{
			TotalCnt: uint32(cnt),
		}

		if tyName, ok := relation_types.RelateTypeName[ty]; ok {
			tyNameCntMap[tyName] = cnt
		}
	}

	_ = db_redis.MSetRelateTotalCnt(ctx, uid, tyNameCntMap)

	if _, ok := result[relation_types.RelationFans]; ok {
		result[relation_types.RelationFans].NewCnt = db_redis.GetListLen(ctx, db_redis.KeyNewFansList(uid))
	}

	return result
}
