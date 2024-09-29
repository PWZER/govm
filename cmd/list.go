/*
Copyright © 2024 PWZER <pwzergo@gmail.com>
*/
package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/PWZER/govm/internal"
	"github.com/spf13/cobra"
)

var (
	listRemoteVersions bool
	includeAllVersions bool
)

type goVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
}

func parseVersion(version string) goVersion {
	v := goVersion{Major: -1, Minor: -1, Patch: -1, PreRelease: ""}
	re := regexp.MustCompile(`^go(\d+)(\.(\d+))?(\.(\d+))?(rc\d+|beta\d+)?$`)
	matches := re.FindStringSubmatch(version)
	if len(matches) == 7 {
		if len(matches[1]) > 0 {
			v.Major, _ = strconv.Atoi(matches[1])
		}
		if len(matches[3]) > 0 {
			v.Minor, _ = strconv.Atoi(matches[3])
		}
		if len(matches[5]) > 0 {
			v.Patch, _ = strconv.Atoi(matches[5])
		}
		v.PreRelease = matches[6]
	}
	return v
}

func (v goVersion) Less(v2 goVersion) bool {
	if v.Major != v2.Major {
		return v.Major < v2.Major
	}
	if v.Minor != v2.Minor {
		return v.Minor < v2.Minor
	}
	if v.Patch != v2.Patch {
		return v.Patch < v2.Patch
	}
	return v.PreRelease < v2.PreRelease
}

var listCmd = &cobra.Command{
	Use:          "ls",
	Short:        "list versions for golang.",
	Long:         "list versions for golang.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		versions, err := internal.GetVersions(listRemoteVersions, !includeAllVersions)
		if err != nil {
			return err
		}

		sort.Slice(versions, func(i, j int) bool {
			v1 := parseVersion(versions[i].Version)
			v2 := parseVersion(versions[j].Version)
			return v1.Less(v2)
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
			parseVersion(v.Version)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listRemoteVersions, "remote", false, "list remote versions")
	listCmd.Flags().BoolVar(&includeAllVersions, "all", false, "list all versions")
}
