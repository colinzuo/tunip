package main

import (
	"flag"
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/logp/configure"
)

func main() {
	appName := "log_demo"

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName(appName)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	configure.Logging(appName)
	logger := logp.NewLogger("main")

	logger.Info("Hello, world!")
}
