package relation_types

import "time"

type AbnormalMgoModel struct {
	ID         string    `json:"id" bson:"_id"`
	Uid        uint64    `json:"uid" bson:"uid"`
	Data       string    `json:"data" bson:"data"`
	Info       string    `json:"info" bson:"info"`
	Status     uint8     `json:"status" bson:"status"`
	EventTime  int64     `json:"event_time" bson:"event_time"`
	CreateTime time.Time `json:"create_time" bson:"create_time"`
	UpdateTime time.Time `json:"update_time" bson:"update_time"`
}

type FollowMgoModel struct {
	SeqID     string `bson:"_id"` // 序列号，唯一ID "uid|follow_uid"
	Uid       uint64 `bson:"uid"`
	FollowUid uint64 `bson:"follow_uid"`
	//Source     string `bson:"source"`
	CreateTime int64 `bson:"create_time"`
}

type FansMgoModel struct {
	SeqID   string `bson:"_id"` // 序列号，唯一ID "uid|fans_uid"
	Uid     uint64 `bson:"uid"`
	FansUid uint64 `bson:"fans_uid"`
	//Source     string `bson:"source"`
	CreateTime int64 `bson:"create_time"`
}

type FriendMgoModel struct {
	SeqID     string `bson:"_id"` // 序列号，唯一ID "uid|friend_uid"
	Uid       uint64 `bson:"uid"`
	FriendUid uint64 `bson:"friend_uid"`
	//Source     string `bson:"source"`
	CreateTime int64 `bson:"create_time"`
}

type ApplyFriendMgoModel struct {
	SeqID     string `bson:"_id"` // 序列号，唯一ID "uid|target_uid"
	Uid       uint64 `bson:"uid"`
	TargetUid uint64 `bson:"target_uid"`
	Status    string `bson:"status"` // 0申请中 1被拒绝
	//Source     string `bson:"source"`
	CreateTime int64 `bson:"create_time"`
}

type BlockMgoModel struct {
	SeqID    string `bson:"_id"` // 序列号，唯一ID "uid|block_uid"
	Uid      uint64 `bson:"uid"`
	BlockUid uint64 `bson:"block_uid"`
	//Source     string `bson:"source"`
	CreateTime int64 `bson:"create_time"`
}
