package util

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func SafeFuncDiy(f func(), recFunc func()) {
	if recFunc != nil {
		defer recFunc()
	} else {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("[super.go] error:%v", err)
				fName := ""
				pc, _, _, ok := runtime.Caller(2)
				if ok {
					f_ := runtime.FuncForPC(pc)
					if f_ != nil {
						fName = f_.Name()
					}
				}
				info := fmt.Sprintf("[super.go] error:%v \n stack:%v", err, string(debug.Stack()))
				Alert(info, fName)
			}
		}()
	}
	f()
}

func SafeFunc(ctx context.Context, fName string, f func()) {
	defer func() {
		if err := recover(); err != nil {
			logrus.WithContext(ctx).Errorf("[super.go] error:%v", err)
			Alert(fmt.Sprintf("[super.SafeFunc] error:%v \n traceID:{%v} \n stack:%v", err, ctx.Value("t"), string(debug.Stack())), fName)
		}
	}()
	f()
}

func SafeFuncErrAlert(ctx context.Context, fName string, f func() error) {
	defer func() {
		if err := recover(); err != nil {
			logrus.WithContext(ctx).Errorf("[super.SafeFuncErrAlert] error:%v", err)
			Alert(fmt.Sprintf("[super.SafeFuncErrAlert] error:%v \n traceID:{%v} \n stack:%v", err, ctx.Value("t"), string(debug.Stack())), fName)
		}
	}()
	e := f()
	if e != nil {
		Alert(fmt.Sprintf("[super.go] traceID:{%v} error:%v", ctx.Value("t"), e), fName)
	}
}

func SafeFuncDiyAfterRecover(ctx context.Context, fName string, f func(), recFunc func()) {
	defer func() {
		if err := recover(); err != nil {
			logrus.WithContext(ctx).Errorf("[super.SafeFuncDiyAfterRecover] error:%v", err)
			Alert(fmt.Sprintf("[super.SafeFuncDiyAfterRecover] error:%v \n traceID:{%v} \n stack:%v", err, ctx.Value("t"), string(debug.Stack())), fName)
			if recFunc != nil {
				recFunc()
			}
		}
	}()
	f()
}
