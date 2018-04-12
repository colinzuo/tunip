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
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName("tunip")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	configure.Logging("tunip")
	logger := logp.NewLogger("main")

	logger.Info("Hello, world!")
}
