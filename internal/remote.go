package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//[{
//  "version": "go1.23.1",
//  "stable": true,
//  "files": [{
//    "filename": "go1.23.1.src.tar.gz",
//    "os": "",
//    "arch": "",
//    "version": "go1.23.1",
//    "sha256": "6ee44e298379d146a5e5aa6b1c5b5d5f5d0a3365eabdd70741e6e21340ec3b0d",
//    "size": 28164249,
//    "kind": "source"
//  }]
//}]

const (
	GolangOfficialVersionsURL = "https://go.dev/dl/?mode=json&include=all"
)

type RemoteVersionFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     uint64 `json:"size"`
	Kind     string `json:"kind"`
}

type RemoteVersion struct {
	Version string              `json:"version"`
	Stable  bool                `json:"stable"`
	Files   []RemoteVersionFile `json:"files"`
}

func getRemoteVersionsCacheFile() string {
	return filepath.Join(getCacheDir(), "versions.json")
}

func getRemoteVersionsFromOfficial() (versions []*RemoteVersion, err error) {
	resp, err := http.Get(GolangOfficialVersionsURL)
	if err != nil {
		return versions, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return versions, err
	}

	err = json.Unmarshal(body, &versions)
	if err != nil {
		return versions, err
	}
	return versions, nil
}

func getRemoteVersionsFromCache() (versions []*RemoteVersion, err error) {
	cacheFile := getRemoteVersionsCacheFile()
	if stat, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return versions, fmt.Errorf("cache file not exists: %s", cacheFile)
	} else if stat.IsDir() {
		return versions, fmt.Errorf("cache file is a directory: %s", cacheFile)
	} else if time.Now().After(stat.ModTime().Add(24 * time.Hour)) {
		return versions, fmt.Errorf("cache file is expired: %s", cacheFile)
	}

	cacheData, err := os.ReadFile(cacheFile)
	if err != nil {
		return versions, err
	}

	err = json.Unmarshal(cacheData, &versions)
	if err != nil {
		return versions, err
	}
	return versions, nil
}

func getRemoteVersions() (versions []*RemoteVersion, err error) {
	versions, err = getRemoteVersionsFromCache()
	if err != nil {
		versions, err = getRemoteVersionsFromOfficial()
		if err != nil {
			return versions, err
		}

		cacheFile := getRemoteVersionsCacheFile()
		cacheData, err := json.MarshalIndent(versions, "", "  ")
		if err != nil {
			return versions, err
		}
		err = os.WriteFile(cacheFile, cacheData, 0644)
		if err != nil {
			return versions, err
		}
	}
	return versions, nil
}

func getRemoteVersion(version string) (v *RemoteVersion, err error) {
	remoteVersions, err := getRemoteVersions()
	if err != nil {
		return v, err
	}

	for _, remoteVersion := range remoteVersions {
		if remoteVersion.Version == version {
			return remoteVersion, nil
		}
	}
	return v, fmt.Errorf("remote version not found: %s", version)
}

func getRemoteVersionFile(version string) (*RemoteVersionFile, error) {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	arch := runtime.GOARCH
	if runtime.GOOS == "linux" && runtime.GOARCH == "arm" {
		arch = "armv6l"
	}

	remoteVersion, err := getRemoteVersion(version)
	if err != nil {
		return nil, err
	}

	for _, remoteVersionFile := range remoteVersion.Files {
		if remoteVersionFile.OS == runtime.GOOS && remoteVersionFile.Arch == arch &&
			strings.HasSuffix(remoteVersionFile.Filename, ext) {
			return &remoteVersionFile, nil
		}
	}
	return nil, fmt.Errorf("remote version file not found: %s", version)
}

func GetRemoteVersions(stableOnly bool) (versions []*Version, err error) {
	remoteVersions, err := getRemoteVersions()
	if err != nil {
		return versions, err
	}

	versions = make([]*Version, 0)
	for _, remoteVersion := range remoteVersions {
		if stableOnly && !remoteVersion.Stable {
			continue
		}

		version := &Version{
			Version: remoteVersion.Version,
			Stable:  remoteVersion.Stable,
			Remote:  remoteVersion,
		}

		localVersion, err := GetLocalVersion(remoteVersion.Version)
		if err == nil {
			version.Local = localVersion
		}

		versions = append(versions, version)
	}
	return versions, nil
}
