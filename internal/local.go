package internal

import (
	"os"
	"path/filepath"
	"runtime"
)

type LocalVersion struct {
	Version    string `json:"version"`
	RootDir    string `json:"RootDir"`
	BinaryFile string `json:"BinaryFile"`
}

func getLocalVersionsRootDir() string {
	rootDir := filepath.Join(getWorkingDir(), "versions")
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		os.MkdirAll(rootDir, 0755)
	}
	return rootDir
}

func getLocalVersionRootDir(version string) string {
	return filepath.Join(getLocalVersionsRootDir(), version)
}

func getLocalVersionInstallDir(version string) string {
	return filepath.Join(getLocalVersionRootDir(version), "install")
}

func getLocalVersionGoPathDir(version string) string {
	return filepath.Join(getLocalVersionRootDir(version), "gopath")
}

func binarySuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func getLocalVersionBinaryPath(version string) string {
	return filepath.Join(getLocalVersionInstallDir(version), "bin", "go"+binarySuffix())
}

func getLocalVersionGoPath(version string) string {
	return filepath.Join(getWorkingDir(), "gopath", version)
}

func getLocalVersionInstallSuccessFile(version string) string {
	return filepath.Join(getLocalVersionInstallDir(version), ".install-success")
}

func localVersionAlreadyInstalled(version string) bool {
	versionDir := getLocalVersionInstallDir(version)
	if _, err := os.Stat(versionDir); err != nil {
		return false
	}

	if _, err := os.Stat(getLocalVersionInstallSuccessFile(version)); err != nil {
		return false
	}

	binaryFile := getLocalVersionBinaryPath(version)
	if _, err := os.Stat(binaryFile); err != nil {
		return false
	}
	return true
}

func setLocalVersionInstalled(version string) error {
	_, err := os.Create(getLocalVersionInstallSuccessFile(version))
	return err
}

func getLocalVersions() (versions []*LocalVersion, err error) {
	rootDir := getLocalVersionsRootDir()
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return versions, nil
	}

	subDirs, err := os.ReadDir(rootDir)
	if err != nil {
		return versions, err
	}

	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			continue
		}

		version := subDir.Name()
		if !localVersionAlreadyInstalled(version) {
			continue
		}

		v := &LocalVersion{
			Version:    version,
			RootDir:    getLocalVersionRootDir(version),
			BinaryFile: getLocalVersionBinaryPath(version),
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func GetLocalVersions(stableOnly bool) (versions []*Version, err error) {
	localVersions, err := getLocalVersions()
	if err != nil {
		return versions, err
	}

	versions = make([]*Version, 0)
	for _, localVersion := range localVersions {
		v := &Version{
			Version: localVersion.Version,
			Stable:  true, // Assume stable by default
			Local:   localVersion,
		}

		if remoteVersion, err := getRemoteVersion(v.Version); err == nil {
			v.Remote = remoteVersion
			v.Stable = remoteVersion.Stable
		}

		if stableOnly && !v.Stable {
			continue
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func GetLocalVersion(version string) (v *LocalVersion, err error) {
	if localVersionAlreadyInstalled(version) {
		v = &LocalVersion{
			Version:    version,
			RootDir:    getLocalVersionRootDir(version),
			BinaryFile: getLocalVersionBinaryPath(version),
		}
	}
	return v, nil
}

func GetCurrentUseVersionBinaryFile() (binaryFile string, err error) {
	goBin := getGoBinaryLinkFile()
	if _, err = os.Stat(goBin); err != nil {
		return binaryFile, nil
	}
	return os.Readlink(goBin)
}
