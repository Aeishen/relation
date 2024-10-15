package work_pool

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab-vywrajy.micoworld.net/yoho-go/ytools/graceful"
	"relation/util"
	"time"
)

var stopCtx context.Context
var stopCancel context.CancelFunc
var poolAbnormalLog *logrus.Logger

func initPoolAbnormalLog() {
	poolAbnormalLog = util.GetDiyLog("log/pool_abnormal.log", true)
}

type pool struct {
	handleRouter map[uint32]DataHandler
	workQueue    []chan *GoWorkerData
}

var _pool *pool

func initGoroutinePool() {
	stopCtx, stopCancel = context.WithCancel(context.Background())
	_pool = &pool{
		handleRouter: make(map[uint32]DataHandler),
		workQueue:    make([]chan *GoWorkerData, 500),
	}
	_pool.startHandle()

	graceful.Close(func() {
		stopCancel()
		time.Sleep(100 * time.Millisecond)
		_pool.close()
	})
}

func (p *pool) startHandle() {
	// 开启工作池，workerID 用数组索引，自增
	for workerID, n := 0, len(p.workQueue); workerID < n; workerID++ {
		p.workQueue[workerID] = make(chan *GoWorkerData, 100)
		worker := p.workQueue[workerID]

		// 每个worker单独工作
		go p.startOneWorker(worker)
	}
}

func (p *pool) addCmdHandler(cmd uint32, handler DataHandler) {
	logrus.Infof("[pool.AddCmdHandler] cmd:{%d} start", cmd)
	if _, ok := p.handleRouter[cmd]; ok {
		return
	}
	p.handleRouter[cmd] = handler
	logrus.Infof("[ReqHandler.AddCmdHandler] cmd:{%d} end, len(r.handleRouter):{%d}", cmd, len(p.handleRouter))
}

func (p *pool) close() {
	for _, worker := range p.workQueue {
		close(worker)
	}
	p.workQueue = nil
}

func (p *pool) startOneWorker(worker chan *GoWorkerData) {
	for {
		select {
		case req, ok := <-worker:
			if !ok {
				return
			}
			p.handle(req)
		}
	}
}

func (p *pool) handle(req *GoWorkerData) {
	//从绑定好的消息和对应的处理方法中执行对应的Handle方法
	handler, ok := p.handleRouter[req.Cmd]
	if !ok {
		logrus.Warnf("[pool.handle] handler not found with cmd:{%d} ", req.Cmd)
		return
	}
	util.SafeFunc(context.Background(), "", func() {
		handler.Handle(req)
	})
}

func SendToWorkQueue(ctx context.Context, opUid uint64, cmd uint32, data map[string]any) {
	goWorkData := getGoWorkerData()
	goWorkData.Cmd = cmd
	goWorkData.Ctx = ctx
	goWorkData.ID = time.Now().UnixMilli()
	goWorkData.Data = data

	if opUid != 0 {
		goWorkData.ID = int64(opUid)
		goWorkData.Data["opUid"] = opUid
	}

	lenWorkQueue := len(_pool.workQueue)
	if lenWorkQueue <= 0 {
		// 没有启动工作池，直接处理该请求
		go _pool.handle(goWorkData)
		return
	}

	//已经启动工作池，将消息交给worker处理, 哈希取模得到要处理连接的workerID
	workerID := int(goWorkData.ID) % lenWorkQueue

	timeCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	select {
	case <-stopCtx.Done():
		if poolAbnormalLog == nil {
			initPoolAbnormalLog()
		}
		info := fmt.Sprintf("server stop not handle data:{%+v, %+v}", goWorkData, data)
		poolAbnormalLog.Error(info)
		util.Alert(info, "SendToWorkQueue")
	case <-timeCtx.Done():
		if poolAbnormalLog == nil {
			initPoolAbnormalLog()
		}
		info := fmt.Sprintf("ctx time out not handle data:{%+v, %+v}", goWorkData, data)
		poolAbnormalLog.Error(info)
		util.Alert(info, "SendToWorkQueue")
	case _pool.workQueue[workerID] <- goWorkData:
	}
}
