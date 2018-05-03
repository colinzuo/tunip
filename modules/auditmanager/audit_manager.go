package auditmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/utils"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
)

type baseMessage struct {
	Timestamp *time.Time `json:"timstamp"`
	GUID      string     `json:"guid"`
}

func (msg baseMessage) isValid() bool {
	if msg.Timestamp == nil {
		return false
	}

	if len(msg.GUID) == 0 {
		return false
	}

	return true
}

type batchAction struct {
	Index *baseMessage `json:"index"`
}

func (action batchAction) isValid() bool {
	if action.Index != nil {
		if action.Index.isValid() {
			return true
		}
	}

	return false
}

type workerRequest struct {
	Base     baseMessage
	ReqBody  string
	RspChan  chan workerResponse
	DoneChan chan bool
}

type getLevelResponse struct {
	ErrCode int    `json:"err_code"`
	ErrInfo string `json:"err_info"`
	Level   string `json:"level"`
}

type setLevelRequest struct {
	Level string `json:"level"`
}

type baseResponse struct {
	ErrCode  int    `json:"err_code"`
	ErrInfo  string `json:"err_info"`
	MoreInfo string `json:"more_info"`
}

type pingResponse struct {
	ErrCode int    `json:"err_code"`
	ErrInfo string `json:"err_info"`
	Message string `json:"message"`
}

type indexResponse struct {
	ErrCode int            `json:"err_code"`
	ErrInfo string         `json:"err_info"`
	Detail  workerResponse `json:"detail"`
}

type bulkResponse struct {
	ErrCode int              `json:"err_code"`
	ErrInfo string           `json:"err_info"`
	Details []workerResponse `json:"details"`
}

type workerResponse struct {
	GUID   string `json:"guid"`
	Result string `json:"result"`
	Status int    `json:"status"`
}

type manager struct {
	dispatchChan   chan workerRequest
	freeWorkerChan chan chan []workerRequest
	doneChan       chan bool
	config         Config
	logger         *logp.Logger
}

// Run function to start web listener and serve it
func Run() {
	logger := logp.NewLogger(ModuleName)

	config, _ := initConfig()
	logger.Infof("Run with config %+v", config)

	manager := manager{config: config, logger: logger}
	manager.dispatchChan = make(chan workerRequest, 2000)
	manager.freeWorkerChan = make(chan chan []workerRequest, config.MaxWorker)
	manager.doneChan = make(chan bool)

	go manager.dispatch()

	for i := 0; i < config.MaxWorker; i++ {
		workerID := fmt.Sprintf("worker_%d", i)
		go manager.work(workerID)
	}

	manager.webListen()

	close(manager.doneChan)

	time.Sleep(time.Duration(5) * time.Second)
}

func (m manager) webListen() {
	ginLogger := logp.NewLogger("gin")

	router := gin.New()
	router.Use(utils.Ginzap(ginLogger))
	router.Use(gin.Recovery())

	tunip := router.Group("/tunip")
	{
		tunip.GET("/ping", m.onPing)

		tunip.GET("/level", m.onGetLevel)
		tunip.POST("/level", m.onSetLevel)

		tunip.POST("/_index", m.onIndex)

		tunip.POST("/_bulk", m.onBulk)
	}

	portSpec := fmt.Sprintf(":%d", m.config.WebPort)

	router.Run(portSpec)
}

func (m manager) dispatch() {
	logger := m.logger.Named("Dispatch")

	queue := make([]workerRequest, 0)
	timeSet := false
	timeout := false
	var thresTimeout <-chan time.Time

	for {
		select {
		case req := <-m.dispatchChan:
			logger.Debugf("recv req: %+v", req.Base)
			queue = append(queue, req)
		case <-m.doneChan:
			logger.Infof("recv done")
			return
		case <-thresTimeout:
			logger.Debugf("timeout")
			timeout = true
		}

		for len(queue) >= m.config.BatchSize || timeout {
			timeout = false
			timeSet = false
			thresTimeout = nil
			var sendQueue []workerRequest
			if len(queue) > m.config.BatchSize {
				sendQueue = queue[:m.config.BatchSize]
				queue = queue[m.config.BatchSize:]
			} else {
				sendQueue = queue
				queue = make([]workerRequest, 0)
			}
			select {
			case workerChan := <-m.freeWorkerChan:
				workerChan <- sendQueue
				logger.Debugf("Send out batch request")
			case <-m.doneChan:
				logger.Infof("recv done")
				return
			}
		}

		if len(queue) > 0 && !timeSet {
			timeSet = true
			thresTimeout = time.After(time.Duration(m.config.BatchTimeout) * time.Millisecond)
		}
	}
}

func (m manager) work(workID string) {
	workerChan := make(chan []workerRequest)

	serverAddr := m.config.EsServerAddr
	logger := m.logger.Named(workID)
	ctx := context.Background()

	lastErrorID := -1
	var nextReportTime time.Time
	reportDuration := time.Duration(60) * time.Second

	var client *elastic.Client
	var err error

	defer logger.Info("Leave")

	for {
		if client == nil {
			client, err = elastic.NewClient(elastic.SetURL(serverAddr),
				elastic.SetErrorLog(logger),
				elastic.SetSniff(false))

			if err != nil {
				errorID := 100
				if errorID != lastErrorID || time.Now().After(nextReportTime) {
					lastErrorID = errorID
					nextReportTime = time.Now().Add(reportDuration)

					logger.Warnf("NewClient failed with: %s", err)
				}
			} else {
				defer client.Stop()
			}
		}

		if client != nil {
			info, code, err := client.Ping(serverAddr).Do(ctx)
			if err == nil {
				logger.Infof("Elasticsearch returned with code %d and version %s\n",
					code, info.Version.Number)
				break
			} else {
				errorID := 101
				if errorID != lastErrorID || time.Now().After(nextReportTime) {
					lastErrorID = errorID
					nextReportTime = time.Now().Add(reportDuration)

					logger.Warnf("Ping failed with: %s", err)
				}
			}
		}

		select {
		case <-m.doneChan:
			return
		case <-time.After(time.Duration(10) * time.Second):
			break
		}
	}

	for {
		m.freeWorkerChan <- workerChan

		select {
		case <-m.doneChan:
			return
		case batchRequest := <-workerChan:
			if len(batchRequest) == 0 {
				break
			}
			logger.Debugf("to process %d requests", len(batchRequest))
			reqMap := make(map[string]workerRequest)
			bulkRequest := client.Bulk()
			for _, req := range batchRequest {
				indexDate := req.Base.Timestamp.UTC().Format("2006.01.02")
				indexName := fmt.Sprintf("logstash-%s", indexDate)
				bulkIndexRequest := elastic.NewBulkIndexRequest()
				bulkIndexRequest.Index(indexName).Type("doc").Id(req.Base.GUID).Doc(req.ReqBody)
				bulkRequest.Add(bulkIndexRequest)
				reqMap[req.Base.GUID] = req
			}
			bulkResponse, err := bulkRequest.Do(ctx)
			if err != nil {
				logger.Errorf("Failed to index: %s", bulkResponse.Failed())
			} else {
				items := bulkResponse.Indexed()
				if items != nil {
					for _, item := range items {
						req, ok := reqMap[item.Id]
						if ok {
							rsp := workerResponse{GUID: req.Base.GUID, Result: item.Result, Status: item.Status}
							logger.Debugf("to send rsp for %+v", rsp)
							select {
							case <-m.doneChan:
								return
							case <-req.DoneChan:
								logger.Warn("requester doesn't want rsp")
								break
							case req.RspChan <- rsp:
								break
							}
						}
					}
				}
			}
		}
	}
}

func (m manager) onPing(c *gin.Context) {
	c.JSON(200, pingResponse{ErrCode: ErrCodeOk, ErrInfo: ErrInfoOk, Message: "pong"})
}

func (m manager) onIndex(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, baseResponse{ErrCode: ErrCodeFailedToReadBody,
			ErrInfo: ErrInfoFailedToReadBody})
	}

	var baseMessage baseMessage
	err = json.Unmarshal(body, &baseMessage)
	if err != nil || !baseMessage.isValid() {
		c.JSON(http.StatusBadRequest, baseResponse{ErrCode: ErrCodeFailedToParseBody,
			ErrInfo: ErrInfoFailedToParseBody})
		return
	}

	rspChan := make(chan workerResponse)
	doneChan := make(chan bool)
	timeoutChan := time.After(time.Duration(m.config.ReqTimeout) * time.Millisecond)

	workerReq := workerRequest{Base: baseMessage, ReqBody: string(body),
		RspChan: rspChan, DoneChan: doneChan}

	select {
	case m.dispatchChan <- workerReq:
		break
	case <-timeoutChan:
		m.logger.Errorf("Failed to send request %+v to dispatcher, timeout",
			baseMessage)
		c.JSON(http.StatusInternalServerError, baseResponse{ErrCode: ErrCodeTimeout,
			ErrInfo: ErrInfoTimeout})
		return
	}

	var rsp workerResponse

	select {
	case rsp = <-rspChan:
		break
	case <-timeoutChan:
		m.logger.Errorf("Failed to recv response %+v from worker, timeout",
			baseMessage)
		doneChan <- true
		c.JSON(http.StatusInternalServerError, baseResponse{ErrCode: ErrCodeTimeout,
			ErrInfo: ErrInfoTimeout})
		return
	}

	if rsp.Status >= 200 && rsp.Status < 300 {
		idxRsp := indexResponse{ErrCode: ErrCodeOk, ErrInfo: ErrInfoOk,
			Detail: rsp}
		jsonRsp, _ := json.Marshal(idxRsp)
		m.logger.Debugf("To send rsp: %s", string(jsonRsp))
		c.Data(http.StatusOK, ContentTypeJSON, jsonRsp)
	} else {
		idxRsp := indexResponse{ErrCode: ErrCodeIndex, ErrInfo: ErrInfoIndex,
			Detail: rsp}
		jsonRsp, _ := json.Marshal(idxRsp)
		m.logger.Errorf("To send rsp: %s", string(jsonRsp))
		c.Data(http.StatusInternalServerError, ContentTypeJSON, jsonRsp)
	}
}

func (m manager) onBulk(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, baseResponse{ErrCode: ErrCodeFailedToReadBody,
			ErrInfo: ErrInfoFailedToReadBody})
	}
	batchRequests := make([]workerRequest, 0)
	needAction := true
	needSource := false
	var batchAction batchAction
	var curRequest workerRequest

	rspChan := make(chan workerResponse)
	doneChan := make(chan bool)
	timeoutChan := time.After(time.Duration(m.config.ReqTimeout) * time.Millisecond)

	for _, reqLine := range strings.Split(string(body), "\n") {
		if len(reqLine) == 0 {
			continue
		}
		if needAction {
			err = json.Unmarshal([]byte(reqLine), &batchAction)

			if err != nil || !batchAction.isValid() {
				c.JSON(http.StatusBadRequest, baseResponse{ErrCode: ErrCodeFailedToParseBody,
					ErrInfo:  ErrInfoFailedToParseBody,
					MoreInfo: fmt.Sprintf("failed to parse action line: %s", reqLine)})
				return
			}

			needAction = false
			needSource = true

			curRequest = workerRequest{Base: *batchAction.Index, RspChan: rspChan, DoneChan: doneChan}
		} else {
			curRequest.ReqBody = reqLine
			batchRequests = append(batchRequests, curRequest)
			needAction = true
			needSource = false
		}
	}
	if needSource {
		c.JSON(http.StatusBadRequest, gin.H{KeyErrCode: ErrCodeFailedToParseBody,
			KeyErrInfo: ErrInfoFailedToParseBody})
		return
	}

	details := make([]workerResponse, 0)
	guidToIdxMap := make(map[string]int)

	for idx, workerReq := range batchRequests {
		select {
		case m.dispatchChan <- workerReq:
			details = append(details, workerResponse{GUID: workerReq.Base.GUID, Result: "Miss", Status: -1})
			guidToIdxMap[workerReq.Base.GUID] = idx
		case <-timeoutChan:
			m.logger.Errorf("Failed to send request %+v to dispatcher, timeout",
				workerReq.Base)
			c.JSON(http.StatusInternalServerError, baseResponse{ErrCode: ErrCodeTimeout,
				ErrInfo: ErrInfoTimeout})
			return
		}
	}

	var rsp workerResponse

	i := 0
	for i < len(batchRequests) {
		select {
		case rsp = <-rspChan:
			idx, ok := guidToIdxMap[rsp.GUID]
			if !ok || details[idx].Status >= 0 {
				moreInfo := ""
				if !ok {
					moreInfo = fmt.Sprintf("unexpected worker rsp with GUID: %s", rsp.GUID)
				} else {
					moreInfo = fmt.Sprintf("GUID: %s, idx %d, old status %d, new status %d",
						rsp.GUID, idx, details[idx].Status, rsp.Status)
				}
				close(doneChan)
				c.JSON(http.StatusInternalServerError, baseResponse{ErrCode: ErrCodeUnexpected,
					ErrInfo:  ErrInfoUnexpected,
					MoreInfo: moreInfo})
				return
			}
			i++
			details[idx] = rsp
		case <-timeoutChan:
			moreInfo := ""
			for _, subRsp := range details {
				if subRsp.Status <= 0 {
					moreInfo = fmt.Sprintf("Failed to recv response for %s from worker, timeout",
						subRsp.GUID)
					m.logger.Error(moreInfo)
					break
				}
			}
			close(doneChan)
			c.JSON(http.StatusInternalServerError, baseResponse{ErrCode: ErrCodeTimeout,
				ErrInfo: ErrInfoTimeout, MoreInfo: moreInfo})
			return
		}
	}

	bulkRsp := bulkResponse{ErrCode: ErrCodeOk, ErrInfo: ErrInfoOk,
		Details: details}
	jsonRsp, _ := json.Marshal(bulkRsp)
	m.logger.Debugf("To send rsp: %s", string(jsonRsp))
	c.Data(http.StatusOK, ContentTypeJSON, jsonRsp)
}

func (m manager) onGetLevel(c *gin.Context) {
	level := logp.GetLevel()
	c.JSON(200, getLevelResponse{ErrCode: ErrCodeOk, ErrInfo: ErrInfoOk, Level: level})
}

func (m manager) onSetLevel(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, baseResponse{ErrCode: ErrCodeFailedToReadBody,
			ErrInfo: ErrInfoFailedToReadBody})
	}

	var setLevelRequest setLevelRequest
	err = json.Unmarshal(body, &setLevelRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseResponse{ErrCode: ErrCodeFailedToParseBody,
			ErrInfo: ErrInfoFailedToParseBody})
		return
	}

	err = logp.SetLevel(setLevelRequest.Level)

	if err != nil {
		c.JSON(http.StatusBadRequest, baseResponse{ErrCode: ErrCodeGeneral,
			ErrInfo: err.Error()})
	} else {
		c.JSON(200, baseResponse{ErrCode: ErrCodeOk, ErrInfo: ErrInfoOk})
	}
}
