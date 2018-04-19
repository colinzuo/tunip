package cmd

import (
	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/modules/generator"
	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

var genConfig string

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate sample data for Endpoint and Server",
	Args:  cobra.NoArgs,
	Run:   Generate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	keyName := "genConfig"
	pflag.StringVar(&genConfig, keyName, "generate.json", "Generator configurations")
	generateCmd.Flags().AddFlag(pflag.CommandLine.Lookup(keyName))
}

// Generate main function for generate command
func Generate(cmd *cobra.Command, args []string) {
	logger := logp.NewLogger(ModuleName)
	logger.Infof("Enter with genConfig %s", genConfig)

	generator.Generate(genConfig)
}
