package util

import (
	"fmt"
	"gitlab-vywrajy.micoworld.net/yoho-go/ytools/cast"
	"strings"
)

func AbnormalID(uid, toUid uint64, relation, storage, op uint8) string {
	return fmt.Sprintf("%d|%d|%d|%d|%d", uid, toUid, relation, storage, op)
}

func ParseAbnormalID(abnormalID string) []string {
	return strings.Split(abnormalID, "|")
}

func SeqID(uid, toUid uint64) string {
	return fmt.Sprintf("%d|%d", uid, toUid)
}

func UnLashSeqID(seqID string) (uint64, uint64) {
	data := strings.Split(seqID, "|")
	if len(data) < 2 {
		return 0, 0
	}
	return cast.ToUint64(data[0]), cast.ToUint64(data[1])
}
