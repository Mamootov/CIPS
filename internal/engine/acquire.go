package engine

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// XrayBinaryName returns the platform-specific binary name.
func XrayBinaryName() string {
	if runtime.GOOS == "windows" {
		return "xray.exe"
	}
	return "xray"
}

// FindXrayBinary searches for xray in app directory and system PATH.
func FindXrayBinary() (string, bool) {
	binaryName := XrayBinaryName()

	// 1. Check current directory and subdirectories
	candidates := []string{
		binaryName,
		filepath.Join("bin", binaryName),
		filepath.Join(".", binaryName),
	}

	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), binaryName))
	}

	for _, path := range candidates {
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			abs, err := filepath.Abs(path)
			if err == nil {
				return abs, true
			}
			return path, true
		}
	}

	// 2. Check system PATH
	if path, err := exec.LookPath(binaryName); err == nil {
		return path, true
	}

	return "", false
}

// ProgressFunc reports download progress percentage and status text.
type ProgressFunc func(percent float64, message string)

// EnsureXrayBinary verifies xray presence, or automatically downloads and extracts it.
func EnsureXrayBinary(onProgress ProgressFunc) (string, error) {
	if existing, found := FindXrayBinary(); found {
		if onProgress != nil {
			onProgress(100.0, fmt.Sprintf("Found existing Xray core at %s", existing))
		}
		return existing, nil
	}

	targetDir := "bin"
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bin directory: %w", err)
	}

	destBinary := filepath.Join(targetDir, XrayBinaryName())

	// Determine archive name
	var zipName string
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "arm64" {
			zipName = "Xray-windows-arm64-v8a.zip"
		} else {
			zipName = "Xray-windows-64.zip"
		}
	case "linux":
		if runtime.GOARCH == "arm64" {
			zipName = "Xray-linux-arm64-v8a.zip"
		} else {
			zipName = "Xray-linux-64.zip"
		}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			zipName = "Xray-macos-arm64-v8a.zip"
		} else {
			zipName = "Xray-macos-64.zip"
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Download sources (official GitHub release with reliable mirrors)
	urls := []string{
		fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/latest/download/%s", zipName),
		fmt.Sprintf("https://ghproxy.net/https://github.com/XTLS/Xray-core/releases/latest/download/%s", zipName),
		fmt.Sprintf("https://mirror.ghproxy.com/https://github.com/XTLS/Xray-core/releases/latest/download/%s", zipName),
	}

	tmpZip := filepath.Join(os.TempDir(), fmt.Sprintf("xray_download_%d.zip", time.Now().UnixNano()))
	defer os.Remove(tmpZip)

	var downloadErr error
	client := &http.Client{Timeout: 180 * time.Second}

	for _, downloadURL := range urls {
		if onProgress != nil {
			onProgress(10.0, fmt.Sprintf("Downloading Xray core from %s...", downloadURL))
		}

		err := downloadFile(client, downloadURL, tmpZip, onProgress)
		if err == nil {
			downloadErr = nil
			break
		}
		downloadErr = err
	}

	if downloadErr != nil {
		return "", fmt.Errorf("failed to download Xray core from available mirrors: %w", downloadErr)
	}

	// Unpack archive
	if onProgress != nil {
		onProgress(90.0, "Extracting Xray binary...")
	}

	if err := extractZipBinary(tmpZip, targetDir, XrayBinaryName()); err != nil {
		return "", fmt.Errorf("failed to extract Xray archive: %w", err)
	}

	// Set executable permissions on Unix
	if runtime.GOOS != "windows" {
		_ = os.Chmod(destBinary, 0755)
	}

	absPath, err := filepath.Abs(destBinary)
	if err != nil {
		absPath = destBinary
	}

	if onProgress != nil {
		onProgress(100.0, "Xray core ready")
	}

	return absPath, nil
}

func downloadFile(client *http.Client, url, destPath string, onProgress ProgressFunc) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "CIPS-Installer/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download server returned HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	totalSize := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024)
	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			nw, ew := out.Write(buf[0:nr])
			if ew != nil {
				return ew
			}
			downloaded += int64(nw)
			if totalSize > 0 && onProgress != nil {
				pct := 10.0 + (float64(downloaded)/float64(totalSize))*75.0
				onProgress(pct, fmt.Sprintf("Downloading Xray: %.1f MB / %.1f MB", float64(downloaded)/(1024*1024), float64(totalSize)/(1024*1024)))
			}
		}
		if er != nil {
			if er != io.EOF {
				return er
			}
			break
		}
	}

	return nil
}

func extractZipBinary(zipPath, destDir, binaryName string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		// Extract binary and asset dat files if present
		if strings.EqualFold(name, binaryName) || strings.HasSuffix(strings.ToLower(name), ".dat") {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			targetFilePath := filepath.Join(destDir, name)
			outFile, err := os.OpenFile(targetFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				rc.Close()
				return err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
