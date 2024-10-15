package util

import (
	"fmt"
	"gitlab-vywrajy.micoworld.net/yoho-go/ytools/helper"
	"gitlab-vywrajy.micoworld.net/yoho-go/ytools/yalert"
)

var noticeUrl string
var alertUrl string
var mode string
var testAlert bool

func InitAlert(_alertUrl, _noticeUrl, _mode string, _testAlert bool) {
	alertUrl = _alertUrl
	noticeUrl = _noticeUrl
	mode = _mode
	testAlert = _testAlert
}

func Alert(info, fun string) {
	if !testAlert && mode == "test" {
		return
	}
	a := yalert.AleterGetter(alertUrl, helper.Runtime.Exec())
	msg := fmt.Sprintf("Service:%s \nFunction:%s \ninfo:%s", mode, fun, info)
	_ = a.Alert(msg)
}

func Notice(info, fun string) {
	if !testAlert && mode == "test" {
		return
	}
	a := yalert.AleterGetter(noticeUrl, helper.Runtime.Exec())
	msg := fmt.Sprintf("Service:%s \nFunction:%s \ninfo:%s", mode, fun, info)
	_ = a.Alert(msg)
}
