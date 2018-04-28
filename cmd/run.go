package cmd

import (
	"github.com/colinzuo/tunip/logp"
	"github.com/spf13/cobra"

	"github.com/colinzuo/tunip/modules/auditmanager"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tunip",
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

//Run default cmd
func Run() error {
	logger := logp.NewLogger(ModuleName)
	logger.Info("Enter Run")

	auditmanager.Run()

	return nil
}
