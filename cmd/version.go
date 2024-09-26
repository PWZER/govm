package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "v0.0.1"
var GitCommit = ""

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print govm version",
	Long:  "print govm version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("govm version: %s, git commit: %s\n", Version, GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
