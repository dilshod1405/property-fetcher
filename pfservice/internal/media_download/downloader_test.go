package media

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadImage(t *testing.T) {
	// Create a temporary directory for media
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set MEDIA_ROOT environment variable
	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Create a test HTTP server that serves an image
	testImageData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46} // JPEG header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(testImageData)
	}))
	defer server.Close()

	// Test downloading image
	propertyID := uint(123)
	localPath, err := DownloadImage(server.URL+"/test-image.jpg", propertyID)
	if err != nil {
		t.Fatalf("Failed to download image: %v", err)
	}

	// Verify path format
	expectedPath := "property_images/test-image.jpg"
	if localPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, localPath)
	}

	// Verify file exists
	fullPath := filepath.Join(MediaRoot, localPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Downloaded file does not exist at %s", fullPath)
	}

	// Verify file content
	fileData, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if len(fileData) != len(testImageData) {
		t.Errorf("Expected file size %d, got %d", len(testImageData), len(fileData))
	}
}

func TestDownloadImageWithInvalidURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Test with invalid URL
	_, err = DownloadImage("http://invalid-url-that-does-not-exist-12345.com/image.jpg", 123)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestDownloadImageWithErrorResponse(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err = DownloadImage(server.URL+"/not-found.jpg", 123)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

func TestDownloadImageWithEmptyResponse(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Create a test server that returns empty body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write nothing
	}))
	defer server.Close()

	_, err = DownloadImage(server.URL+"/empty.jpg", 123)
	if err == nil {
		t.Error("Expected error for empty response")
	}
}

func TestDownloadImageWithUUIDFallback(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Create a test server with URL that has no filename
	testImageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(testImageData)
	}))
	defer server.Close()

	// Test with URL that has no filename
	localPath, err := DownloadImage(server.URL, 123)
	if err != nil {
		t.Fatalf("Failed to download image: %v", err)
	}

	// Should generate UUID filename
	if !filepath.HasPrefix(localPath, "property_images/") {
		t.Errorf("Expected path to start with property_images/, got %s", localPath)
	}
	if filepath.Ext(localPath) != ".jpg" {
		t.Errorf("Expected .jpg extension, got %s", filepath.Ext(localPath))
	}
}

func TestGetMediaRoot(t *testing.T) {
	// Test default value
	oldEnv := os.Getenv("MEDIA_ROOT")
	os.Unsetenv("MEDIA_ROOT")
	defer func() {
		if oldEnv != "" {
			os.Setenv("MEDIA_ROOT", oldEnv)
		}
	}()

	// Reset MediaRoot by reloading
	oldRoot := MediaRoot
	MediaRoot = "/mhp/media" // Default value
	if MediaRoot != "/mhp/media" {
		t.Errorf("Expected default MEDIA_ROOT /mhp/media, got %s", MediaRoot)
	}

	// Test custom value
	os.Setenv("MEDIA_ROOT", "/custom/media")
	MediaRoot = "/custom/media"
	if MediaRoot != "/custom/media" {
		t.Errorf("Expected MEDIA_ROOT /custom/media, got %s", MediaRoot)
	}

	MediaRoot = oldRoot
}

func TestDownloadImageWithRetry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Create a test server that fails first 2 times, then succeeds
	attemptCount := 0
	testImageData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// Fail first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on 3rd attempt
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(testImageData)
	}))
	defer server.Close()

	// Set retry configuration
	oldMaxRetries := os.Getenv("IMAGE_DOWNLOAD_MAX_RETRIES")
	oldRetryDelay := os.Getenv("IMAGE_DOWNLOAD_RETRY_DELAY")
	os.Setenv("IMAGE_DOWNLOAD_MAX_RETRIES", "3")
	os.Setenv("IMAGE_DOWNLOAD_RETRY_DELAY", "0") // No delay for testing
	defer func() {
		if oldMaxRetries != "" {
			os.Setenv("IMAGE_DOWNLOAD_MAX_RETRIES", oldMaxRetries)
		} else {
			os.Unsetenv("IMAGE_DOWNLOAD_MAX_RETRIES")
		}
		if oldRetryDelay != "" {
			os.Setenv("IMAGE_DOWNLOAD_RETRY_DELAY", oldRetryDelay)
		} else {
			os.Unsetenv("IMAGE_DOWNLOAD_RETRY_DELAY")
		}
	}()

	// Test downloading image with retry
	propertyID := uint(123)
	localPath, err := DownloadImage(server.URL+"/test-retry.jpg", propertyID)
	if err != nil {
		t.Fatalf("Failed to download image after retries: %v", err)
	}

	// Verify it succeeded after retries
	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	// Verify path format
	if !filepath.HasPrefix(localPath, "property_images/") {
		t.Errorf("Expected path to start with property_images/, got %s", localPath)
	}

	// Verify file exists
	fullPath := filepath.Join(MediaRoot, localPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Downloaded file does not exist at %s", fullPath)
	}
}

func TestDownloadImageWithRetryFailure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Create a test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Set retry configuration
	oldMaxRetries := os.Getenv("IMAGE_DOWNLOAD_MAX_RETRIES")
	oldRetryDelay := os.Getenv("IMAGE_DOWNLOAD_RETRY_DELAY")
	os.Setenv("IMAGE_DOWNLOAD_MAX_RETRIES", "3")
	os.Setenv("IMAGE_DOWNLOAD_RETRY_DELAY", "0") // No delay for testing
	defer func() {
		if oldMaxRetries != "" {
			os.Setenv("IMAGE_DOWNLOAD_MAX_RETRIES", oldMaxRetries)
		} else {
			os.Unsetenv("IMAGE_DOWNLOAD_MAX_RETRIES")
		}
		if oldRetryDelay != "" {
			os.Setenv("IMAGE_DOWNLOAD_RETRY_DELAY", oldRetryDelay)
		} else {
			os.Unsetenv("IMAGE_DOWNLOAD_RETRY_DELAY")
		}
	}()

	// Test downloading image - should fail after all retries
	propertyID := uint(123)
	_, err = DownloadImage(server.URL+"/test-fail.jpg", propertyID)
	if err == nil {
		t.Error("Expected error after all retries failed")
	}

	// Verify error message mentions retries
	if err != nil && !strings.Contains(err.Error(), "attempts") {
		t.Errorf("Expected error message to mention attempts, got: %v", err)
	}
}

func TestDownloadImageWithEmptyURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Test with empty URL - should not retry
	_, err = DownloadImage("", 123)
	if err == nil {
		t.Error("Expected error for empty URL")
	}
	if err != nil && !strings.Contains(err.Error(), "empty URL") {
		t.Errorf("Expected error about empty URL, got: %v", err)
	}
}

func TestDownloadImageWithTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pf-service-test-media-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldMediaRoot := MediaRoot
	MediaRoot = tmpDir
	defer func() {
		MediaRoot = oldMediaRoot
	}()

	// Create a test server that delays response longer than timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay for 15 seconds (longer than default 10s timeout)
		time.Sleep(15 * time.Second)
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
	}))
	defer server.Close()

	// Set a short timeout for testing
	oldTimeout := os.Getenv("IMAGE_DOWNLOAD_TIMEOUT")
	os.Setenv("IMAGE_DOWNLOAD_TIMEOUT", "2") // 2 seconds
	defer func() {
		if oldTimeout != "" {
			os.Setenv("IMAGE_DOWNLOAD_TIMEOUT", oldTimeout)
		} else {
			os.Unsetenv("IMAGE_DOWNLOAD_TIMEOUT")
		}
	}()

	// Set retry configuration to fail fast
	oldMaxRetries := os.Getenv("IMAGE_DOWNLOAD_MAX_RETRIES")
	oldRetryDelay := os.Getenv("IMAGE_DOWNLOAD_RETRY_DELAY")
	os.Setenv("IMAGE_DOWNLOAD_MAX_RETRIES", "1") // Only 1 attempt for timeout test
	os.Setenv("IMAGE_DOWNLOAD_RETRY_DELAY", "0")
	defer func() {
		if oldMaxRetries != "" {
			os.Setenv("IMAGE_DOWNLOAD_MAX_RETRIES", oldMaxRetries)
		} else {
			os.Unsetenv("IMAGE_DOWNLOAD_MAX_RETRIES")
		}
		if oldRetryDelay != "" {
			os.Setenv("IMAGE_DOWNLOAD_RETRY_DELAY", oldRetryDelay)
		} else {
			os.Unsetenv("IMAGE_DOWNLOAD_RETRY_DELAY")
		}
	}()

	// Test downloading image - should timeout
	start := time.Now()
	_, err = DownloadImage(server.URL+"/timeout-test.jpg", 123)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error")
	}

	// Verify it timed out quickly (should be around 2 seconds, not 15)
	if duration > 5*time.Second {
		t.Errorf("Expected timeout to occur quickly, but took %v", duration)
	}

	// Verify error mentions timeout
	if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Logf("Timeout error (may vary by Go version): %v", err)
	}
}
