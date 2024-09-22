/*
Copyright © 2024 PWZER <pwzergo@gmail.com>
*/
package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/PWZER/govm/internal"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

var (
	listRemoteVersions bool
	includeAllVersions bool
)

var listCmd = &cobra.Command{
	Use:          "ls",
	Short:        "list versions for golang.",
	Long:         "list versions for golang.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var getVersionsFunc func(bool) ([]*internal.Version, error)
		if listRemoteVersions {
			getVersionsFunc = internal.GetRemoteVersions
		} else {
			getVersionsFunc = internal.GetLocalVersions
		}

		versions, err := getVersionsFunc(!includeAllVersions)
		if err != nil {
			return err
		}

		sort.Slice(versions, func(i, j int) bool {
			return semver.Compare(strings.ReplaceAll(versions[i].Version, "go", "v"),
				strings.ReplaceAll(versions[j].Version, "go", "v")) < 0
		})

		curBinaryFile, err := internal.GetCurrentUseVersionBinaryFile()
		if err != nil {
			return err
		}

		for _, v := range versions {
			if (listRemoteVersions && v.Local != nil) ||
				(!listRemoteVersions && v.Local.BinaryFile == curBinaryFile) {
				fmt.Printf("[*] %s\n", v.Version)
			} else {
				fmt.Printf("[ ] %s\n", v.Version)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listRemoteVersions, "remote", false, "list remote versions")
	listCmd.Flags().BoolVar(&includeAllVersions, "all", false, "list all versions")
}
