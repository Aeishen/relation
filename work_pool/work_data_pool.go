package work_pool

import (
	"context"
	"sync"
)

type GoWorkerData struct {
	Ctx  context.Context
	Cmd  uint32
	ID   int64
	Data map[string]any
}

func (p *GoWorkerData) Reset() {
	p.Ctx = nil
	p.Cmd = 0
	p.ID = 0
	p.Data = nil // todo 优化 map 生成
}

// GoWorkerData 对象缓冲池
var workerDataPool = sync.Pool{
	New: func() interface{} {
		return &GoWorkerData{}
	},
}

func getGoWorkerData() *GoWorkerData {
	v, ok := workerDataPool.Get().(*GoWorkerData)
	if ok {
		return v
	}
	return &GoWorkerData{}
}

func putGoWorkerData(v *GoWorkerData) {
	v.Reset()
	workerDataPool.Put(v)
}
