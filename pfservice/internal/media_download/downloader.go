package media

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

var MediaRoot = getMediaRoot()

func getMediaRoot() string {
	if v := os.Getenv("MEDIA_ROOT"); v != "" {
		return v
	}
	return "/mhp/media"
}


func DownloadImage(url string, propertyID uint) (string, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status %d for url %s", resp.StatusCode, url)
	}

	saveDir := filepath.Join(MediaRoot, "property_images")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir failed %s: %w", saveDir, err)
	}

	filename := filepath.Base(resp.Request.URL.Path)
	if filename == "" || filename == "/" || len(filename) < 4 {
		filename = uuid.New().String() + ".jpg"
	}

	fullPath := filepath.Join(saveDir, filename)

	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("file create failed %s: %w", fullPath, err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("file write failed %s: %w", fullPath, err)
	}

	if written == 0 {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("empty file downloaded from %s", url)
	}

	info, err := os.Stat(fullPath)
	if err != nil || info.Size() == 0 {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("file verification failed %s", fullPath)
	}

	return filepath.Join("property_images", filename), nil
}
