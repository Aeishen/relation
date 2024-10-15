package util

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"gitlab-vywrajy.micoworld.net/yoho-go/ylogs/ylogrus"
	"path/filepath"
	"runtime"
)

func shortPathCallerFormatter(frame *runtime.Frame) string {
	return fmt.Sprintf(" [%v:%v %v]", filepath.Base(frame.File), frame.Line, filepath.Base(frame.Function))
}

func GetDiyLog(logPath string, isShort bool) *logrus.Logger {
	l := ylogrus.NewLogger(ylogrus.WithFileName(logPath))
	if isShort {
		l.SetFormatter(&nested.Formatter{
			TimestampFormat:       "2006-01-02 15:04:05.000",
			ShowFullLevel:         true,
			CallerFirst:           true,
			CustomCallerFormatter: shortPathCallerFormatter,
		})
	}
	return l
}
