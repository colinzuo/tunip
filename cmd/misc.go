package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/colinzuo/tunip/internal/miscmanager"
	"github.com/colinzuo/tunip/pkg/logp"
)

var miscConfig string

// miscCmd represents the misc command
var miscCmd = &cobra.Command{
	Use:   "misc",
	Short: "Run misc function according to specified configurations",
	Args:  cobra.NoArgs,
	Run:   misc,
}

func init() {
	rootCmd.AddCommand(miscCmd)

	keyName := "miscConfig"
	pflag.StringVar(&miscConfig, keyName, "./configs/misc.json", "Misc configurations")
	miscCmd.Flags().AddFlag(pflag.CommandLine.Lookup(keyName))
}

// misc main function for misc command
func misc(cmd *cobra.Command, args []string) {
	logger := logp.NewLogger(ModuleName)
	logger.Infof("Enter with miscConfig %s", miscConfig)

	miscmanager.Run(miscConfig)
}
