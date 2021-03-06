package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/colinzuo/tunip/pkg/logp"
	"github.com/colinzuo/tunip/pkg/logp/configure"
)

var cfgFile string
var appName = "tunip"

// Module Name
const (
	ModuleName string = "Cmd"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Tunip with Go",
	Run:   runCmd.Run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./configs/tunip.json or $HOME/tunip.json)")

	// for log
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	rootCmd.PersistentFlags().AddFlag(pflag.CommandLine.Lookup("verbose"))
	rootCmd.PersistentFlags().AddFlag(pflag.CommandLine.Lookup("toStderr"))
	rootCmd.PersistentFlags().AddFlag(pflag.CommandLine.Lookup("debug"))
	rootCmd.PersistentFlags().AddFlag(pflag.CommandLine.Lookup("logConfig"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name $appName (without extension).
		viper.AddConfigPath(filepath.Join(".", "configs"))
		viper.AddConfigPath(home)
		viper.SetConfigName(appName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())

		logConfig := viper.GetString("logConfig")

		if _, err := os.Stat(logConfig); os.IsNotExist(err) {
			if filepath.IsAbs(logConfig) {
				log.Panicf("logConfig %s doesn't exist", logConfig)
			}
			dir := filepath.Dir(viper.ConfigFileUsed())
			logConfig = filepath.Join(dir, logConfig)
			viper.Set("logConfig", logConfig)
			fmt.Println("Update logConfig file:", logConfig)
		}

		if _, err := os.Stat(logConfig); os.IsNotExist(err) {
			log.Panicf("Updated logConfig %s also doesn't exist", logConfig)
		}
	}

	configure.Logging(appName)
	logger := logp.NewLogger(ModuleName)

	logger.Info("Cobra init done")
}
