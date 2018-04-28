package auditmanager

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/utils"
	"github.com/gin-gonic/gin"
)

type bulkItem struct {
	Action string
	Source string
}

// Run function to start web listener and serve it
func Run() {
	logger := logp.NewLogger(ModuleName)
	ginLogger := logp.NewLogger("gin")

	config, _ := initConfig()
	logger.Infof("Run with config %+v", config)

	router := gin.New()
	router.Use(utils.Ginzap(ginLogger))
	router.Use(gin.Recovery())

	tunip := router.Group("/tunip")
	{
		tunip.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				KeyErrCode: ErrCodeOk,
				KeyErrInfo: ErrInfoOk,
				"message":  "pong",
			})
		})

		tunip.POST("/_index", func(c *gin.Context) {
			body, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{KeyErrCode: ErrCodeFailedToReadBody,
					KeyErrInfo: ErrInfoFailedToReadBody})
			}
			c.JSON(http.StatusOK, gin.H{KeyErrCode: ErrCodeOk,
				KeyErrInfo: ErrInfoOk,
				"body":     string(body)})
		})

		tunip.POST("/_bulk", func(c *gin.Context) {
			body, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{KeyErrCode: ErrCodeFailedToReadBody,
					KeyErrInfo: ErrInfoFailedToReadBody})
			}
			bulkItems := make([]bulkItem, 0)
			var curAction string
			needAction := true
			for _, reqLine := range strings.Split(string(body), "\n") {
				if needAction {
					curAction = reqLine
					needAction = false
					logger.Infof("action line: %s", curAction)
				} else {
					logger.Infof("source line: %s", reqLine)
					bulkItems = append(bulkItems, bulkItem{Action: curAction, Source: reqLine})
					needAction = true
				}
			}
			c.JSON(http.StatusOK, gin.H{KeyErrCode: ErrCodeOk,
				KeyErrInfo: ErrInfoOk,
				"body":     bulkItems})
		})
	}

	portSpec := fmt.Sprintf(":%d", config.WebPort)

	router.Run(portSpec)
}
