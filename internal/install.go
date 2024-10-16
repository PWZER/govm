package internal

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

func verifyFileSHA256(file, sha256sum string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return err
	}

	fileSum := fmt.Sprintf("%x", hash.Sum(nil))
	if fileSum != sha256sum {
		return fmt.Errorf("file %s sha256sum mismatch, expect %s but got %s", file, sha256sum, fileSum)
	}
	return nil
}

func downloadFileFromURL(saveFile, srcURL string) (err error) {
	fd, err := os.Create(saveFile)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			fd.Close()
			os.Remove(saveFile)
		}
	}()
	c := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			DisableKeepAlives:  true,
			Proxy:              http.ProxyFromEnvironment,
		},
	}
	res, err := c.Get(srcURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}

	bar := progressbar.DefaultBytes(
		res.ContentLength,
		"Downloading",
	)
	n, err := io.Copy(io.MultiWriter(fd, bar), res.Body)
	if err != nil {
		return err
	}
	if res.ContentLength != -1 && res.ContentLength != n {
		return fmt.Errorf("downloaded size mismatch, expect %d but got %d", res.ContentLength, n)
	}
	return fd.Close()
}

func downloadArchiveFile(saveFile, url, sha256sum string) error {
	if _, err := os.Stat(saveFile); err == nil {
		if err := verifyFileSHA256(saveFile, sha256sum); err == nil {
			return nil
		}
		os.Remove(saveFile)
	}

	// Check if the url archive file exists
	res, err := http.Head(url)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("archive file not found: %s", url)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("request archive file info failed, Status: %s, url: %s", http.StatusText(res.StatusCode), url)
	}

	// Download the archive file
	if err := downloadFileFromURL(saveFile, url); err != nil {
		return err
	}

	// Check file size
	if stat, err := os.Stat(saveFile); os.IsNotExist(err) {
		return fmt.Errorf("downloaded file not found: %s", saveFile)
	} else if stat.Size() != res.ContentLength {
		return fmt.Errorf("downloaded file size mismatch, expect %d but got %d", res.ContentLength, stat.Size())
	}

	// Verify the downloaded archive file
	if err := verifyFileSHA256(saveFile, sha256sum); err != nil {
		os.Remove(saveFile)
		return err
	}
	return nil
}

func Install(version string) (err error) {
	if localVersionAlreadyInstalled(version) {
		return nil
	}

	remoteVersionFile, err := getRemoteVersionFile(version)
	if err != nil {
		return err
	}

	saveFile := filepath.Join(getCacheDir(), remoteVersionFile.Filename)
	downloadUrl := fmt.Sprintf("%s/%s", Config.InstallMirror, remoteVersionFile.Filename)
	fmt.Println("Downloading from", downloadUrl)
	if err := downloadArchiveFile(saveFile, downloadUrl, remoteVersionFile.SHA256); err != nil {
		return err
	}

	installDir := getLocalVersionInstallDir(version)
	if err := unpackArchiveFile(saveFile, installDir); err != nil {
		return err
	}

	return setLocalVersionInstalled(version)
}
