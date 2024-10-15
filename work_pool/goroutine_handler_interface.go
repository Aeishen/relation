package work_pool

type DataHandler interface {
	Handle(param *GoWorkerData)
}
