package media

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadImageUniqueFilenames(t *testing.T) {
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

	// Create test server with UUID in URL path
	testImageData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(testImageData)
	}))
	defer server.Close()

	propertyID := uint(123)
	
	// Test multiple images with same "original.jpg" ending - should get unique names
	urls := []string{
		server.URL + "/media/images/listing/ID1/ce5950dd-d4b0-478e-ad32-176b8900bef1/original.jpg",
		server.URL + "/media/images/listing/ID2/991f8733-ac63-4b7b-8a44-d8586cae82a0/original.jpg",
		server.URL + "/media/images/listing/ID3/741275b0-4d08-4129-a86c-7a69537e7aba/original.jpg",
	}

	downloadedFiles := make(map[string]bool)
	
	for idx, url := range urls {
		localPath, err := DownloadImage(url, propertyID, idx)
		if err != nil {
			t.Fatalf("Failed to download image %d: %v", idx, err)
		}

		// Check filename is unique
		filename := filepath.Base(localPath)
		if downloadedFiles[filename] {
			t.Errorf("Duplicate filename detected: %s", filename)
		}
		downloadedFiles[filename] = true

		// Verify file exists
		fullPath := filepath.Join(MediaRoot, localPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Downloaded file does not exist at %s", fullPath)
		}

		// Verify filename contains UUID or property ID
		if !strings.Contains(filename, "ce5950dd-d4b0-478e-ad32-176b8900bef1") && 
		   !strings.Contains(filename, "991f8733-ac63-4b7b-8a44-d8586cae82a0") &&
		   !strings.Contains(filename, "741275b0-4d08-4129-a86c-7a69537e7aba") {
			// If UUID not in filename, should have property ID
			if !strings.Contains(filename, "pf_123") {
				t.Errorf("Filename should contain UUID or property ID, got: %s", filename)
			}
		}
	}

	// Verify we got 3 unique files
	if len(downloadedFiles) != 3 {
		t.Errorf("Expected 3 unique files, got %d", len(downloadedFiles))
	}
}

func TestDownloadImageFilenameWithUUID(t *testing.T) {
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

	testImageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(testImageData)
	}))
	defer server.Close()

	// Test URL with UUID
	url := server.URL + "/media/images/listing/Z1XHGC2QB0ARA317TMC2F5K2ZW/ce5950dd-d4b0-478e-ad32-176b8900bef1/original.jpg"
	localPath, err := DownloadImage(url, 1061, 0)
	if err != nil {
		t.Fatalf("Failed to download image: %v", err)
	}

	filename := filepath.Base(localPath)
	// Should extract UUID from URL
	if !strings.Contains(filename, "ce5950dd-d4b0-478e-ad32-176b8900bef1") {
		t.Errorf("Expected filename to contain UUID, got: %s", filename)
	}
}

func TestDownloadImageFilenameWithoutUUID(t *testing.T) {
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

	testImageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(testImageData)
	}))
	defer server.Close()

	// Test URL with UUID in filename but not in path (tmp path)
	// URL: /media/tmp/979aa4e2-0ce0-4717-b230-81a2babeeac6/f4440543-f51c-4227-bb92-6b24bbf466a3_original.jpg
	// The UUID f4440543-f51c-4227-bb92-6b24bbf466a3 should be extracted
	url := server.URL + "/media/tmp/979aa4e2-0ce0-4717-b230-81a2babeeac6/f4440543-f51c-4227-bb92-6b24bbf466a3_original.jpg"
	localPath, err := DownloadImage(url, 1052, 5)
	if err != nil {
		t.Fatalf("Failed to download image: %v", err)
	}

	filename := filepath.Base(localPath)
	// Should extract UUID from filename or use property ID + index + hash
	// The UUID f4440543-f51c-4227-bb92-6b24bbf466a3 might be extracted, or property ID format used
	hasUUID := strings.Contains(filename, "f4440543-f51c-4227-bb92-6b24bbf466a3")
	hasPropertyID := strings.Contains(filename, "pf_1052")
	if !hasUUID && !hasPropertyID {
		t.Errorf("Expected filename to contain UUID or property ID, got: %s", filename)
	}
}
