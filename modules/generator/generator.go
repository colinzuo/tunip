package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/utils"
	"github.com/olivere/elastic"
)

// Generator generator struct
type Generator struct {
	Logger *logp.Logger
	Config *Config

	client        *elastic.Client
	clients       []*elastic.Client
	channelNum    int
	confStartTime time.Time
	confEndTime   time.Time
	confPeriod    time.Duration
	timeLongForm  string
}

// ParseGenConfig parse generator config
func ParseGenConfig(genConfig string) (*Config, error) {
	logger := logp.NewLogger(ModuleName)
	content, err := ioutil.ReadFile(genConfig)
	if err != nil {
		logger.Errorf("read genConfig %s failed, %s", genConfig, err)
		return nil, err
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		logger.Errorf("parse genConfig %s failed, %s", genConfig, err)
		return nil, err
	}
	return &config, nil
}

// Generate generate fake data
func Generate(genConfig string) error {
	logger := logp.NewLogger(ModuleName)
	var config *Config
	config, err := ParseGenConfig(genConfig)
	if err != nil {
		logger.Panicf("ParseGenConfig: failed with %s", err)
	}
	logger.Infof("genConfig %s, content: %+v", genConfig, config)

	ctx := context.Background()
	serverAddr := config.ServerAddr

	client, err := elastic.NewClient(elastic.SetURL(serverAddr),
		elastic.SetErrorLog(logger),
		elastic.SetInfoLog(logger),
		elastic.SetSniff(false))
	if err != nil {
		logger.Panic(err)
	}
	defer client.Stop()

	info, code, err := client.Ping(serverAddr).Do(ctx)
	if err != nil {
		logger.Panic(err)
	}
	logger.Infof("Elasticsearch returned with code %d and version %s\n",
		code, info.Version.Number)

	channelNum := 10
	clients := make([]*elastic.Client, channelNum)
	for i := 0; i < 10; i++ {
		clients[i], err = elastic.NewClient(elastic.SetURL(serverAddr),
			elastic.SetErrorLog(logger),
			elastic.SetInfoLog(logger),
			elastic.SetSniff(false))
		if err != nil {
			logger.Panic(err)
		}
		defer clients[i].Stop()
	}

	generator := Generator{Logger: logger, Config: config,
		client:       client,
		clients:      clients,
		channelNum:   channelNum,
		timeLongForm: "2006-01-02T15:04:05.000-0700"}
	generator.Generate()

	return nil
}

func (g *Generator) generateSampleMcuConf() {
	logger := g.Logger
	client := g.client

	ctx := context.Background()

	guid, err := utils.NewUUID()
	sampleMcuConf := McuConf{
		Type:       "SERVER_MCU_CONF",
		GUID:       guid,
		ConfDetail: McuConfDetail{Number: "8001001"},
		StartTime:  "2017-04-17T15:30:34.742+0800",
		EndTime:    "2017-04-17T17:35:34.742+0800",
		Duration:   7500,
		ErrorCode:  0,
		ErrorInfo:  "OK",
	}
	fbWrapper := FilebeatWrapper{
		Timestamp: sampleMcuConf.EndTime,
		Hostname:  "generator",
		Fields: FilebeatFields{
			ContainerType: "mru",
		},
		JSONExtract: sampleMcuConf,
	}

	startTime, _ := time.Parse(g.timeLongForm, sampleMcuConf.StartTime)
	logger.Infof("StartTime %s", startTime)

	indexDate := startTime.UTC().Format("2006.01.02")
	logger.Infof("Formatted StartTime %s", indexDate)

	indexName := fmt.Sprintf("logstash-%s", indexDate)
	put1, err := client.Index().
		Index(indexName).
		Type("doc").
		BodyJson(fbWrapper).
		Do(ctx)
	if err != nil {
		logger.Panic(err)
	}
	logger.Infof("Index mcuConf %s to index %s, type %s\n", put1.Id,
		put1.Index, put1.Type)
}

func (g *Generator) generateMcuConf() {
	logger := g.Logger
	config := g.Config
	client := g.client

	ctx := context.Background()

	logger.Infof("Enter generateMcuConf")

	for i := 0; i < 100; i++ {
		// select time range
		offset := time.Duration(rand.Float64()*(g.confPeriod.Seconds()-(float64)(config.McuConfConfig.DurationMin))) * time.Second
		startTime := g.confStartTime.Add(offset)
		duration := time.Duration(config.McuConfConfig.DurationMin+
			rand.Intn(config.McuConfConfig.DurationMax-config.McuConfConfig.DurationMin)) * time.Second
		endTime := startTime.Add(duration)

		if endTime.After(g.confEndTime) {
			logger.Infof("#%d: endTime %s is after configed endTime %s", i, endTime, g.confEndTime)
			continue
		}

		// select conf number
		confNumber := config.McuConfConfig.NumberMin +
			rand.Intn(config.McuConfConfig.NumberMax-config.McuConfConfig.NumberMin)

		// check if conf number is free during this period
		indexDate := startTime.UTC().Format("2006.01.02")
		indexName := fmt.Sprintf("logstash-%s", indexDate)

		logger.Infof("indexDate %s, confNumber %d, startTime %s, endTime %s", indexDate, confNumber, startTime, endTime)

		termQuery := elastic.NewTermQuery("json_extract.conf.number", strconv.Itoa(confNumber))
		rangeQuery1 := elastic.NewRangeQuery("json_extract.start_time")
		rangeQuery1.Gte(startTime.Format(g.timeLongForm))
		rangeQuery1.Lte(endTime.Format(g.timeLongForm))
		rangeQuery2 := elastic.NewRangeQuery("json_extract.end_time")
		rangeQuery2.Gte(startTime.Format(g.timeLongForm))
		rangeQuery2.Lte(endTime.Format(g.timeLongForm))
		rangeQuery3 := elastic.NewRangeQuery("json_extract.start_time")
		rangeQuery3.Lte(startTime.Format(g.timeLongForm))
		rangeQuery4 := elastic.NewRangeQuery("json_extract.end_time")
		rangeQuery4.Gte(endTime.Format(g.timeLongForm))
		boolQuery2 := elastic.NewBoolQuery()
		boolQuery2.Must(rangeQuery3, rangeQuery4)
		boolQuery3 := elastic.NewBoolQuery()
		boolQuery3.Should(rangeQuery1, rangeQuery2, boolQuery2).MinimumNumberShouldMatch(1)
		boolQuery := elastic.NewBoolQuery()
		boolQuery.Must(termQuery)
		boolQuery.Filter(boolQuery3)
		querySource, _ := boolQuery.Source()
		logger.Infof("query source: %s", querySource)
		searchResult, err := client.Search(indexName).
			Query(boolQuery).
			From(0).Size(10).
			Pretty(true).
			Do(ctx)
		if err == nil {
			logger.Infof("Found a total of %d confs\n", searchResult.TotalHits())
			if searchResult.TotalHits() > 0 {
				continue
			}
		} else {
			elasticErr, ok := err.(*elastic.Error)
			if !(ok && elasticErr.Status == 404) {
				logger.Infof("search failed with: %s", err)
			}
		}

		guid, err := utils.NewUUID()
		mcuConf := McuConf{
			Type:       "SERVER_MCU_CONF",
			GUID:       guid,
			ConfDetail: McuConfDetail{Number: strconv.Itoa(confNumber)},
			StartTime:  startTime.Format(g.timeLongForm),
			EndTime:    endTime.Format(g.timeLongForm),
			Duration:   (int)(duration.Seconds()),
			ErrorCode:  0,
			ErrorInfo:  "OK",
		}
		fbWrapper := FilebeatWrapper{
			Timestamp: mcuConf.EndTime,
			Hostname:  "generator",
			Fields: FilebeatFields{
				ContainerType: "mru",
			},
			JSONExtract: mcuConf,
		}

		put1, err := client.Index().
			Index(indexName).
			Type("doc").
			BodyJson(fbWrapper).
			Do(ctx)
		if err != nil {
			logger.Errorf("index fail with error: %f", err)
			continue
		}
		logger.Infof("Index mcuConf %s to index %s, type %s\n", put1.Id,
			put1.Index, put1.Type)

		if config.GenMcuCall {
			interval := rand.Intn(4)
			min := 0
			max := 0
			switch interval {
			case 0:
				min = config.McuCallConfig.NumMin
				max = config.McuCallConfig.Num25
			case 1:
				min = config.McuCallConfig.Num25
				max = config.McuCallConfig.Num50
			case 2:
				min = config.McuCallConfig.Num50
				max = config.McuCallConfig.Num75
			case 3:
				min = config.McuCallConfig.Num75
				max = config.McuCallConfig.NumMax
			}
			callNum := min + rand.Intn(max-min)
			syncChannel := make(chan int, g.channelNum)
			defer close(syncChannel)
			for k := 0; k < g.channelNum; k++ {
				syncChannel <- k
			}

			var wg sync.WaitGroup
			wg.Add(callNum)
			for j := 0; j < callNum; j++ {
				go func(callSeq int) {
					channelNum := <-syncChannel
					logger.Infof("To generate call %d / %d using channel %d", callSeq, callNum, channelNum)
					g.generateMcuCall(mcuConf, g.clients[channelNum])
					syncChannel <- channelNum
					wg.Done()
				}(j)
			}
			wg.Wait()
		}

		return
	}

	logger.Panic("Failed after try many times")
}

func (g *Generator) generateMcuCall(mcuConf McuConf, client *elastic.Client) {
	logger := g.Logger
	config := g.Config

	ctx := context.Background()

	logger.Infof("Enter generateMcuConf")

	confStartTime, _ := time.Parse(g.timeLongForm, mcuConf.StartTime)
	confEndTime, _ := time.Parse(g.timeLongForm, mcuConf.EndTime)

	for i := 0; i < 100; i++ {
		// select time range
		offset := time.Duration(rand.Intn(mcuConf.Duration-config.McuCallConfig.DurationMin)) * time.Second
		startTime := confStartTime.Add(offset)
		duration := time.Duration(config.McuCallConfig.DurationMin+
			rand.Intn(config.McuCallConfig.DurationMax-config.McuCallConfig.DurationMin)) * time.Second
		endTime := startTime.Add(duration)

		if endTime.After(confEndTime) {
			logger.Infof("#%d: endTime %s is after confEndTime %s", i, endTime, confEndTime)
			continue
		}

		// check if conf number is free during this period
		indexDate := startTime.UTC().Format("2006.01.02")
		indexName := fmt.Sprintf("logstash-%s", indexDate)
		logger.Infof("indexDate %s", indexDate)

		guid, err := utils.NewUUID()
		mcuCall := McuCall{
			Type:       "SERVER_MCU_CALL",
			GUID:       guid,
			confGUID:   mcuConf.GUID,
			CallDetail: McuCallDetail{Number: mcuConf.ConfDetail.Number},
			StartTime:  startTime.Format(g.timeLongForm),
			EndTime:    endTime.Format(g.timeLongForm),
			Duration:   (int)(duration.Seconds()),
			ErrorCode:  0,
			ErrorInfo:  "OK",
		}
		fbWrapper := FilebeatWrapper{
			Timestamp: mcuCall.EndTime,
			Hostname:  "generator",
			Fields: FilebeatFields{
				ContainerType: "mru",
			},
			JSONExtract: mcuCall,
		}

		put1, err := client.Index().
			Index(indexName).
			Type("doc").
			BodyJson(fbWrapper).
			Do(ctx)
		if err != nil {
			logger.Errorf("index fail with error: %f", err)
			continue
		}
		logger.Infof("Index mcuCall %s to index %s, type %s\n", put1.Id,
			put1.Index, put1.Type)

		return
	}

	logger.Errorf("Failed after try many times")
}

// Generate generate according to config
func (g *Generator) Generate() {
	logger := g.Logger
	config := g.Config

	logger.Infof("Enter with config: %+v", config)

	rand.Seed((int64)(time.Now().Second()))

	if config.GenMcuConf {
		g.confStartTime, _ = time.Parse(g.timeLongForm, config.McuConfConfig.StartTime)
		g.confEndTime, _ = time.Parse(g.timeLongForm, config.McuConfConfig.EndTime)
		g.confPeriod = g.confEndTime.Sub(g.confStartTime)
		logger.Infof("confStartTime %s, confEndTime %s", g.confStartTime, g.confEndTime)

		for i := 1; i < config.McuConfConfig.Num; i++ {
			g.generateMcuConf()
		}
	}
}
