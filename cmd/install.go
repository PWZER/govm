/*
Copyright © 2024 PWZER <pwzergo@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/PWZER/govm/internal"
	"github.com/spf13/cobra"
)

var (
	installMirror string
)

var installCmd = &cobra.Command{
	Use:          "install",
	Short:        "Install specified version of golang",
	Long:         "Install specified version of golang",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if installMirror != "" {
			internal.Config.InstallMirror = installMirror
		}

		if len(args) > 0 {
			for _, version := range args {
				if err := internal.Install(version); err != nil {
					return err
				}
			}
			return nil
		}
		return fmt.Errorf("please specify the version of golang you want to install")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVarP(&installMirror, "mirror", "m", internal.Config.InstallMirror,
		"Specify the mirror you want to download the golang package.")
}
