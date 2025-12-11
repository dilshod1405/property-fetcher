package media

import (
    "io"
    "net/http"
    "os"
    "path/filepath"
)

const MediaRoot = "/var/www/mhp-api/media" // Django MEDIA_ROOT

func DownloadImage(url string, propertyID uint) (string, error) {

    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    // Django media folder
    saveDir := filepath.Join(MediaRoot, "property_images")
    os.MkdirAll(saveDir, 0755)

    // filename
    filename := filepath.Base(url)

    fullPath := filepath.Join(saveDir, filename)

    out, err := os.Create(fullPath)
    if err != nil {
        return "", err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return "", err
    }

    //"property_images/filename.jpg for django"
    return filepath.Join("property_images", filename), nil
}
