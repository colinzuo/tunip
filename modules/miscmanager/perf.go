package miscmanager

import (
	"fmt"
	"time"

	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/utils"
)

// WorkerRequest request wrapper
type WorkerRequest struct {
	Type     string
	GUID     string
	Body     interface{}
	RspChan  chan interface{}
	DoneChan chan bool
}

// WorkerContext def
type WorkerContext struct {
	logger *logp.Logger
	wr     WorkerRequest
}

// SampleWorkerReq def
type SampleWorkerReq struct {
	GUID string `json:"guid"`
}

// BaseResponse definition
type BaseResponse struct {
	ErrCode  int    `json:"err_code"`
	ErrInfo  string `json:"err_info"`
	MoreInfo string `json:"more_info"`
}

// SampleWorkerRsp def
type SampleWorkerRsp struct {
	BaseResponse
	GUID string `json:"guid"`
}

func (m *Manager) perfTest() {
	logger := m.logger
	config := m.config.PerfTest

	logger.Info("Enter")
	defer logger.Info("Leave")

	m.dispatchChan = make(chan WorkerRequest, 10000)
	m.freeWorkerChan = make(chan chan interface{}, config.MaxWorker)
	m.doneChan = make(chan bool)

	go m.dispatch()

	for i := 0; i < config.MaxWorker; i++ {
		workerID := fmt.Sprintf("worker_%d", i)
		go m.work(workerID)
	}

	m.sendRequestWaitRsp()

	close(m.doneChan)

	time.Sleep(time.Duration(5) * time.Second)
}

func (m Manager) dispatch() {
	logger := m.logger.Named("Dispatch")

	var req WorkerRequest

	for {
		select {
		case req = <-m.dispatchChan:
			logger.Debugf("recv req: %+v", req)
			select {
			case workerChan := <-m.freeWorkerChan:
				workerChan <- req
				logger.Debugf("Send out request %s  %s", req.Type, req.GUID)
			case <-m.doneChan:
				logger.Infof("recv done")
				return
			}
		case <-m.doneChan:
			logger.Infof("recv done")
			return
		}
	}
}

func (m Manager) work(workID string) {
	workerChan := make(chan interface{})
	logger := m.logger.Named(workID)
	logger.Info("Enter")
	defer logger.Info("Leave")

	var inRequest interface{}

	workerCtx := &WorkerContext{logger: logger}

	for {
		m.freeWorkerChan <- workerChan

		select {
		case <-m.doneChan:
			return
		case inRequest = <-workerChan:
			break
		}

		switch v := inRequest.(type) {
		case WorkerRequest:
			workerCtx.wr = v
			switch v.Type {
			case RequestSampleWorkerReq:
				m.workerOnSampleWorkerReq(workerCtx)
			default:
				logger.Errorf("Unexpected type %s", v.Type)
			}
		default:
			logger.Error("Unexpected request type")
		}
	}
}

func (m Manager) workerSendRsp(workerCtx *WorkerContext, rsp interface{}) {
	workerCtx.logger.Debugf("to send rsp for %+v", rsp)
	select {
	case <-m.doneChan:
		return
	case <-workerCtx.wr.DoneChan:
		workerCtx.logger.Debug("requester doesn't want rsp for %s  %s", workerCtx.wr.Type, workerCtx.wr.GUID)
		break
	case workerCtx.wr.RspChan <- rsp:
		break
	}
}

func (m Manager) workerOnSampleWorkerReq(workerCtx *WorkerContext) {
	logger := workerCtx.logger

	logger.Infof("recv req: %+v", workerCtx.wr)

	sampleWorkerReq := workerCtx.wr.Body.(SampleWorkerReq)

	logger.Infof("process req: %+v", sampleWorkerReq)

	var rsp SampleWorkerRsp

	logger.Infof("send out rsp: %+v", rsp)

	m.workerSendRsp(workerCtx, rsp)
}

func (m Manager) sendRequestWaitRsp() {
	logger := m.logger.Named("sendRequestWaitRsp")
	config := m.config.PerfTest

	rspChan := make(chan interface{}, config.Number)
	doneChan := m.doneChan

	var i int
	for i = 0; i < config.Number; i++ {
		guid := utils.NewUUID()

		workerReq := WorkerRequest{Type: RequestSampleWorkerReq,
			GUID: guid,
			Body: SampleWorkerReq{
				GUID: guid,
			},
			RspChan:  rspChan,
			DoneChan: doneChan}

		logger.Infof("Send out req %d: %+v", i+1, workerReq)
		m.dispatchChan <- workerReq
	}

	var genericRsp interface{}
	rspNum := 0

	for rspNum < config.Number {
		genericRsp = <-rspChan
		rspNum++

		lrsp := genericRsp.(SampleWorkerRsp)

		logger.Infof("Recv rsp %d: %+v", rspNum, lrsp)
	}
}
