package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

type Version struct {
	Version string         `json:"version"`
	Stable  bool           `json:"stable"`
	Local   *LocalVersion  `json:"local"`
	Remote  *RemoteVersion `json:"remote"`
}

func getGoBinaryLinkFile() string {
	return filepath.Join(getUserHomeDir(), ".local", "bin", "go"+binarySuffix())
}

func UseGoVersion(version string) (err error) {
	// link go binary
	goBin := getGoBinaryLinkFile()
	if _, err = os.Stat(goBin); err == nil {
		os.Remove(goBin)
	}

	if !localVersionAlreadyInstalled(version) {
		fmt.Printf("Local version %s not found, installing ...\n", version)
		if err = Install(version); err != nil {
			return err
		}
	}

	versionBin := getLocalVersionBinaryPath(version)
	if err = os.Symlink(versionBin, goBin); err != nil {
		return err
	}

	// link gopath
	goPath := filepath.Join(getUserHomeDir(), ".govm", "go")
	if _, err = os.Stat(goPath); err == nil {
		os.Remove(goPath)
	}

	versionGoPath := getLocalVersionGoPathDir(version)
	if _, err = os.Stat(versionGoPath); os.IsNotExist(err) {
		if err = os.MkdirAll(versionGoPath, 0755); err != nil {
			return err
		}
	}

	if err = os.Symlink(versionGoPath, goPath); err != nil {
		return err
	}
	fmt.Printf("Using go version %s\n", version)
	return nil
}

func GetVersions(remote bool, all bool) (versions []*Version, err error) {
	if remote {
		return GetRemoteVersions(all)
	} else {
		return GetLocalVersions(all)
	}
}
