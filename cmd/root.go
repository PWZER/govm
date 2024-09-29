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

		infos := map[string]string{
			"GoVM version":           Version,
			"GoVM git commit":        GitCommit,
			"Working directory":      internal.Config.WorkingDir,
			"Current use go version": version,
		}

		maxKeyLen := 0
		for key := range infos {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}
		}

		for key, value := range infos {
			fmt.Printf("%s: %s\n", key+strings.Repeat(" ", maxKeyLen-len(key)), value)
		}
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
