package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	latestVersionUrl string = "https://api.github.com/repos/PWZER/govm/releases/latest"
)

type versionAsset struct {
	Url                string `json:"url"`
	Name               string `json:"name"`
	Size               int    `json:"size"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type latestVersion struct {
	TagName string         `json:"tag_name"`
	Assets  []versionAsset `json:"assets"`
}

func Upgrade(dummy bool, currentVersion string) (err error) {
	req, err := http.NewRequest(http.MethodGet, latestVersionUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get latest version: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var latestVersion latestVersion
	if err := json.Unmarshal(body, &latestVersion); err != nil {
		return err
	}

	if latestVersion.TagName == "" {
		return fmt.Errorf("failed to get latest version")
	}

	if latestVersion.TagName <= currentVersion {
		fmt.Println("already the latest version")
		return nil
	}

	if dummy {
		fmt.Printf("Current version: %s, Latest version: %s\n", currentVersion, latestVersion.TagName)
		return err
	}

	binPath, err := os.Executable()
	if err != nil {
		return err
	}

	var useAsset *versionAsset = nil
	expectedAssetName := fmt.Sprintf("govm-%s-%s", runtime.GOOS, runtime.GOARCH)
	for _, asset := range latestVersion.Assets {
		if asset.Name == expectedAssetName {
			useAsset = &asset
			break
		}
	}

	if useAsset == nil {
		return fmt.Errorf("no asset found for %s", expectedAssetName)
	}
	fmt.Printf("Upgrade version %s => %s, downloading from %s\n",
		currentVersion, latestVersion.TagName, useAsset.BrowserDownloadUrl)

	tmpPath := filepath.Join(filepath.Dir(binPath), fmt.Sprintf(".%s.tmp", filepath.Base(binPath)))
	if err := downloadArchiveFileFromURL(tmpPath, useAsset.BrowserDownloadUrl); err != nil {
		return err
	}

	// check file size
	stat, err := os.Stat(tmpPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("downloaded file not found: %s", tmpPath)
	} else if stat.Size() != int64(useAsset.Size) {
		return fmt.Errorf("downloaded file size mismatch, expect %d but got %d", useAsset.Size, stat.Size())
	}

	// make it executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	// replace the binary
	if err := os.Rename(tmpPath, binPath); err != nil {
		return err
	}
	fmt.Printf("upgrade to version %s successfully\n", latestVersion.TagName)
	return nil
}
