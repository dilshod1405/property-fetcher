package media

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var MediaRoot = getMediaRoot()

const (
	defaultMaxRetries   = 3
	defaultRetryDelay   = 2 * time.Second
	defaultDownloadTimeout = 10 * time.Second
)

func getMediaRoot() string {
	if v := os.Getenv("MEDIA_ROOT"); v != "" {
		return v
	}
	return "/mhp/media"
}

func getMaxRetries() int {
	if v := os.Getenv("IMAGE_DOWNLOAD_MAX_RETRIES"); v != "" {
		if retries, err := strconv.Atoi(v); err == nil && retries > 0 {
			return retries
		}
	}
	return defaultMaxRetries
}

func getRetryDelay() time.Duration {
	if v := os.Getenv("IMAGE_DOWNLOAD_RETRY_DELAY"); v != "" {
		if delay, err := strconv.Atoi(v); err == nil && delay > 0 {
			return time.Duration(delay) * time.Second
		}
	}
	return defaultRetryDelay
}

func getDownloadTimeout() time.Duration {
	if v := os.Getenv("IMAGE_DOWNLOAD_TIMEOUT"); v != "" {
		if timeout, err := strconv.Atoi(v); err == nil && timeout > 0 {
			return time.Duration(timeout) * time.Second
		}
	}
	return defaultDownloadTimeout
}

// downloadImageAttempt performs a single download attempt
// Uses configurable timeout to prevent long-running downloads from blocking
func downloadImageAttempt(url string, propertyID uint) (string, error) {
	timeout := getDownloadTimeout()
	client := &http.Client{
		Timeout: timeout,
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

// DownloadImage downloads an image with retry logic
// It will retry up to maxRetries times if the download fails
// Returns the relative path to the downloaded image or an error
func DownloadImage(url string, propertyID uint) (string, error) {
	// Skip if URL is empty
	if url == "" {
		return "", fmt.Errorf("empty URL provided")
	}

	maxRetries := getMaxRetries()
	retryDelay := getRetryDelay()
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		path, err := downloadImageAttempt(url, propertyID)
		if err == nil {
			// Success on first attempt, no need to log
			if attempt > 1 {
				// Log successful retry
				fmt.Printf("Image download succeeded on attempt %d for URL: %s\n", attempt, url)
			}
			return path, nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt < maxRetries {
			fmt.Printf("Image download attempt %d/%d failed for URL %s: %v. Retrying in %v...\n",
				attempt, maxRetries, url, err, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	// All retries failed
	return "", fmt.Errorf("image download failed after %d attempts: %w", maxRetries, lastErr)
}
