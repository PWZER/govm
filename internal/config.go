package internal

import (
	"os"
	"path/filepath"
)

type ConfigType struct {
	WorkingDir    string `json:"working_dir"`
	InstallMirror string `json:"install_mirror"`
}

var Config ConfigType

func init() {
	defaultWorkingDir := filepath.Join(getUserHomeDir(), ".govm")

	Config = ConfigType{
		WorkingDir:    defaultWorkingDir,
		InstallMirror: "https://golang.google.cn/dl/",
	}
}

func getUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDir
}

func getWorkingDir() string {
	if _, err := os.Stat(Config.WorkingDir); os.IsNotExist(err) {
		os.MkdirAll(Config.WorkingDir, 0755)
	}
	return Config.WorkingDir
}

func getCacheDir() string {
	cacheDir := filepath.Join(getWorkingDir(), "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.MkdirAll(cacheDir, 0755)
	}
	return cacheDir
}
