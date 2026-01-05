package media

import (
    "io"
    "net/http"
    "os"
    "path/filepath"
    "fmt"
    "github.com/google/uuid"
)

var MediaRoot = getMediaRoot()

func getMediaRoot() string {
    if v := os.Getenv("MEDIA_ROOT"); v != "" {
        return v
    }
    return "/var/www/app/media"
}

func DownloadImage(url string, propertyID uint) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("http get failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("bad status %d for url %s", resp.StatusCode, url)
    }

    saveDir := filepath.Join(MediaRoot, "property_images")
    if err := os.MkdirAll(saveDir, 0755); err != nil {
        return "", fmt.Errorf("mkdir failed: %w", err)
    }

    filename := filepath.Base(resp.Request.URL.Path)
    if filename == "" || filename == "/" {
        filename = uuid.New().String() + ".jpg"
    }

    fullPath := filepath.Join(saveDir, filename)

    out, err := os.Create(fullPath)
    if err != nil {
        return "", fmt.Errorf("file create failed %s: %w", fullPath, err)
    }
    defer out.Close()

    if _, err := io.Copy(out, resp.Body); err != nil {
        return "", fmt.Errorf("file write failed %s: %w", fullPath, err)
    }

    return filepath.Join("property_images", filename), nil
}
