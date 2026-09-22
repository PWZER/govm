package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	repoOwner      = "PWZER"
	repoName       = "govm"
	latestPageUrl  = "https://github.com/PWZER/govm/releases/latest"
	latestApiUrl   = "https://api.github.com/repos/PWZER/govm/releases/latest"
	downloadUrlFmt = "https://github.com/PWZER/govm/releases/download/%s/%s"
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

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DisableCompression: true,
			DisableKeepAlives:  true,
			Proxy:              http.ProxyFromEnvironment,
		},
	}
}

// getLatestVersionViaRedirect resolves the latest release tag from the
// releases page redirect (no API rate limits apply to this endpoint).
func getLatestVersionViaRedirect() (*latestVersion, error) {
	c := newHTTPClient(15 * time.Second)
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := c.Get(latestPageUrl)
	if err != nil {
		return nil, fmt.Errorf("request failed, err: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusMultipleChoices || resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	location := resp.Header.Get("Location")
	if location == "" {
		return nil, errors.New("empty location header")
	}
	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("invalid location: %s, err: %s", location, err)
	}
	prefix := fmt.Sprintf("/%s/%s/releases/tag/", repoOwner, repoName)
	if !strings.HasPrefix(u.Path, prefix) {
		return nil, fmt.Errorf("unexpected location: %s", location)
	}
	tag, err := url.PathUnescape(strings.TrimPrefix(u.Path, prefix))
	if err != nil || tag == "" {
		return nil, fmt.Errorf("invalid tag in location: %s", location)
	}
	return &latestVersion{TagName: tag}, nil
}

func getLatestVersionViaAPI() (*latestVersion, error) {
	req, err := http.NewRequest(http.MethodGet, latestApiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request, err: %s", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "govm-upgrade")
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := newHTTPClient(15 * time.Second).Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed, err: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		errMsg := fmt.Sprintf("status: %s, body: %s", resp.Status, strings.TrimSpace(string(body)))
		if resetAt, resetErr := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); resetErr == nil {
			errMsg += fmt.Sprintf(", rate limit resets at %s", time.Unix(resetAt, 0).Format(time.RFC3339))
		}
		if token == "" {
			errMsg += " (hint: set GITHUB_TOKEN to raise the rate limit)"
		}
		return nil, errors.New(errMsg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read latest version, err: %s", err)
	}

	var latestVersion latestVersion
	if err := json.Unmarshal(body, &latestVersion); err != nil {
		return nil, fmt.Errorf("failed to unmarshal latest version, err: %s", err)
	}

	return &latestVersion, nil
}

func getLatestVersion() (*latestVersion, error) {
	latestVersion, err := getLatestVersionViaRedirect()
	if err == nil {
		return latestVersion, nil
	}
	latestVersion, apiErr := getLatestVersionViaAPI()
	if apiErr != nil {
		return nil, fmt.Errorf("failed to get latest version, redirect: %s, api: %s", err, apiErr)
	}
	return latestVersion, nil
}

func Upgrade(dummy bool, currentVersion string) (err error) {
	latestVersion, err := getLatestVersion()
	if err != nil {
		return err
	}

	if latestVersion.TagName == "" {
		return fmt.Errorf("got the latest version is empty")
	}

	if latestVersion.TagName <= currentVersion {
		fmt.Printf("already the latest version, or current version: %s is newer than latest version: %s\n",
			currentVersion, latestVersion.TagName)
		return nil
	}

	if dummy {
		fmt.Printf("current version: %s, latest version: %s\n", currentVersion, latestVersion.TagName)
		return nil
	}

	var useAsset *versionAsset = nil
	expectedAssetName := fmt.Sprintf("govm-%s-%s", runtime.GOOS, runtime.GOARCH)
	for _, asset := range latestVersion.Assets {
		if asset.Name == expectedAssetName {
			useAsset = &asset
			break
		}
	}

	if len(latestVersion.Assets) > 0 && useAsset == nil {
		return fmt.Errorf("latest version %s not found asset for %s", latestVersion.TagName, expectedAssetName)
	}

	downloadUrl := fmt.Sprintf(downloadUrlFmt, latestVersion.TagName, expectedAssetName)
	if useAsset != nil {
		downloadUrl = useAsset.BrowserDownloadUrl
	}
	fmt.Printf("Upgrade version %s => %s, downloading from %s\n",
		currentVersion, latestVersion.TagName, downloadUrl)

	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get binary path, err: %s", err)
	}

	tmpPath := filepath.Join(filepath.Dir(binPath), fmt.Sprintf(".%s.tmp", filepath.Base(binPath)))
	if err := downloadFileFromURL(tmpPath, downloadUrl); err != nil {
		return fmt.Errorf("failed to download file, err: %s", err)
	}

	// check file size
	if useAsset != nil {
		stat, err := os.Stat(tmpPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("downloaded file not found: %s", tmpPath)
		} else if stat.Size() != int64(useAsset.Size) {
			return fmt.Errorf("downloaded file size mismatch, expect %d but got %d", useAsset.Size, stat.Size())
		}
	}

	// make it executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to make file executable, err: %s", err)
	}

	// replace the binary
	if err := os.Rename(tmpPath, binPath); err != nil {
		return fmt.Errorf("failed to replace binary, err: %s", err)
	}
	fmt.Printf("upgrade to version %s successfully\n", latestVersion.TagName)
	return nil
}
