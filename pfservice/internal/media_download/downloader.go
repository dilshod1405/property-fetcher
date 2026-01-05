package media

import (
    "io"
    "net/http"
    "os"
    "path/filepath"
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
        return "", err
    }
    defer resp.Body.Close()

    saveDir := filepath.Join(MediaRoot, "property_images")
    if err := os.MkdirAll(saveDir, 0755); err != nil {
        return "", err
    }

    filename := filepath.Base(url)
    fullPath := filepath.Join(saveDir, filename)

    out, err := os.Create(fullPath)
    if err != nil {
        return "", err
    }
    defer out.Close()

    if _, err = io.Copy(out, resp.Body); err != nil {
        return "", err
    }
    return filepath.Join("property_images", filename), nil
}