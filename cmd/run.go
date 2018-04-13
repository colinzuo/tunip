package cmd

import (
	"github.com/colinzuo/tunip/cmd/impl"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tunip",
	Run: func(cmd *cobra.Command, args []string) {
		impl.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
