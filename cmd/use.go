/*
Copyright © 2024 PWZER <pwzergo@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/PWZER/govm/internal"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Specify the version of golang to use",
	Long:  "Specify the version of golang to use",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Please specify the version of golang to use, just like 'govm use 1.16.3'")
		}
		version := args[0]
		if version == "" {
			return fmt.Errorf("Please specify the version of golang to use")
		}
		return internal.UseGoVersion(version)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
