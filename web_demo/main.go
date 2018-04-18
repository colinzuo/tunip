package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/logp/configure"
	"github.com/colinzuo/tunip/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// Login Binding from JSON
type Login struct {
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func main() {
	appName := "web_demo"

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName(appName)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	configure.Logging(appName)
	logger := logp.NewLogger("gin")

	router := gin.New()
	router.Use(utils.Ginzap(logger))
	router.Use(gin.Recovery())

	tunip := router.Group("/tunip")
	{
		tunip.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		tunip.POST("/login", func(c *gin.Context) {
			var json Login
			if err := c.ShouldBindWith(&json, binding.JSON); err == nil {
				if json.User == "colinzuo" && json.Password == "123456" {
					c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
		})

		tunip.GET("/user/:name", func(c *gin.Context) {
			name := c.Param("name")
			c.String(http.StatusOK, "Hello %s", name)
		})

		// However, this one will match /user/john/ and also /user/john/send
		// If no other routers match /user/john, it will redirect to /user/john/
		tunip.GET("/user/:name/*action", func(c *gin.Context) {
			name := c.Param("name")
			action := c.Param("action")
			message := name + " is " + action
			c.String(http.StatusOK, message)
		})
	}

	router.Run() // listen and serve on 0.0.0.0:8080
}
