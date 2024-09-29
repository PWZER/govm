/*
Copyright © 2024 PWZER <pwzergo@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PWZER/govm/internal"
	"github.com/spf13/cobra"
)

type govmInfos []struct {
	Key   string
	Value string
}

func (infos govmInfos) show() {
	maxKeyLen := 0
	for _, info := range infos {
		if len(info.Key) > maxKeyLen {
			maxKeyLen = len(info.Key)
		}
	}

	for _, info := range infos {
		fmt.Printf("%s: %s\n", info.Key+strings.Repeat(" ", maxKeyLen-len(info.Key)), info.Value)
	}
}

var rootCmd = &cobra.Command{
	Use:          "govm",
	Short:        "A simple version manager for golang.",
	Long:         "A simple version manager for golang.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binaryFile, err := internal.GetCurrentUseVersionBinaryFile()
		if err != nil {
			return err
		}

		version := "Unknown"
		names := strings.Split(filepath.ToSlash(binaryFile), "/")
		for i := len(names) - 1; i >= 0; i-- {
			if matched, _ := regexp.MatchString(`^go\d+(\.\d+(\.\d+)?)?$`, names[i]); matched {
				version = names[i]
				break
			}
		}

		govmInfos{
			{"GoVM version", Version},
			{"GoVM git commit", GitCommit},
			{"Working directory", internal.Config.WorkingDir},
			{"Current use go version", version},
		}.show()
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
