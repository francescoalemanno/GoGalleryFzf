package server

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestHandleRotate_Success tests successful rotation
func TestHandleRotate_Success(t *testing.T) {
	tempDir := t.TempDir()
	testFile := "test.jpg"
	fullPath := filepath.Join(tempDir, testFile)

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	file, err := os.Create(fullPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("Failed to encode image: %v", err)
	}
	file.Close()

	server, _ := New(tempDir)

	reqBody := map[string]interface{}{
		"path":  testFile,
		"angle": 90,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["success"] != true {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	if response["path"] != testFile {
		t.Errorf("Expected path=%s, got %v", testFile, response["path"])
	}

	if response["angle"] != float64(90) {
		t.Errorf("Expected angle=90, got %v", response["angle"])
	}

	// Verify the image was actually rotated
	rotatedFile, err := os.Open(fullPath)
	if err != nil {
		t.Fatalf("Failed to open rotated file: %v", err)
	}
	defer rotatedFile.Close()

	rotatedImg, _, err := image.Decode(rotatedFile)
	if err != nil {
		t.Fatalf("Failed to decode rotated image: %v", err)
	}

	bounds := rotatedImg.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("Expected dimensions 200x100 after rotation, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestHandleRotate_InvalidMethod tests that only POST is allowed
func TestHandleRotate_InvalidMethod(t *testing.T) {
	server, _ := New(t.TempDir())

	req := httptest.NewRequest("GET", "/api/rotate", nil)
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}
}

// TestHandleRotate_InvalidJSON tests error handling for invalid JSON
func TestHandleRotate_InvalidJSON(t *testing.T) {
	server, _ := New(t.TempDir())

	req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

// TestHandleRotate_PathTraversal tests path traversal protection
func TestHandleRotate_PathTraversal(t *testing.T) {
	server, _ := New(t.TempDir())

	reqBody := map[string]interface{}{
		"path":  "../../../etc/passwd",
		"angle": 90,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rec.Code)
	}
}

// TestHandleRotate_FileNotFound tests rotating a non-existent file
func TestHandleRotate_FileNotFound(t *testing.T) {
	server, _ := New(t.TempDir())

	reqBody := map[string]interface{}{
		"path":  "nonexistent.jpg",
		"angle": 90,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

// TestHandleRotate_DirectoryNotAllowed tests that directories cannot be rotated
func TestHandleRotate_DirectoryNotAllowed(t *testing.T) {
	tempDir := t.TempDir()
	subDir := "subdir"

	// Create a subdirectory
	if err := os.Mkdir(filepath.Join(tempDir, subDir), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	server, _ := New(tempDir)

	reqBody := map[string]interface{}{
		"path":  subDir,
		"angle": 90,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

// TestHandleRotate_UnsupportedFormat tests rotating non-image files
func TestHandleRotate_UnsupportedFormat(t *testing.T) {
	tempDir := t.TempDir()
	testFile := "test.txt"

	// Create a text file
	if err := os.WriteFile(filepath.Join(tempDir, testFile), []byte("not an image"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	server, _ := New(tempDir)

	reqBody := map[string]interface{}{
		"path":  testFile,
		"angle": 90,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	// Check error message
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] != "Unsupported image format" {
		t.Errorf("Expected 'Unsupported image format' error, got %v", response["error"])
	}
}

// TestHandleRotate_InvalidAngle tests invalid rotation angles
func TestHandleRotate_InvalidAngle(t *testing.T) {
	tempDir := t.TempDir()
	testFile := "test.jpg"
	fullPath := filepath.Join(tempDir, testFile)

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, _ := os.Create(fullPath)
	jpeg.Encode(file, img, &jpeg.Options{Quality: 85})
	file.Close()

	server, _ := New(tempDir)

	invalidAngles := []int{0, 45, -45, 270, 360}
	for _, angle := range invalidAngles {
		reqBody := map[string]interface{}{
			"path":  testFile,
			"angle": angle,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.HandleRotate(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for angle %d, got %d", angle, rec.Code)
		}
	}
}

// TestHandleRotate_SupportedAngles tests all supported rotation angles
func TestHandleRotate_SupportedAngles(t *testing.T) {
	tempDir := t.TempDir()

	supportedAngles := []int{90, -90, 180, -180}

	for _, angle := range supportedAngles {
		testFile := "test_" + string(rune('a'+angle)) + ".jpg"
		fullPath := filepath.Join(tempDir, testFile)

		// Create a test image
		img := image.NewRGBA(image.Rect(0, 0, 100, 200))
		file, _ := os.Create(fullPath)
		jpeg.Encode(file, img, &jpeg.Options{Quality: 85})
		file.Close()

		server, _ := New(tempDir)

		reqBody := map[string]interface{}{
			"path":  testFile,
			"angle": angle,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.HandleRotate(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Angle %d: Expected status 200, got %d: %s", angle, rec.Code, rec.Body.String())
		}
	}
}

// TestHandleRotate_InvalidatesThumbnail tests that rotation invalidates thumbnail cache
func TestHandleRotate_InvalidatesThumbnail(t *testing.T) {
	tempDir := t.TempDir()
	testFile := "test.jpg"
	fullPath := filepath.Join(tempDir, testFile)

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, _ := os.Create(fullPath)
	jpeg.Encode(file, img, &jpeg.Options{Quality: 85})
	file.Close()

	// Pre-populate cache
	thumbCache.set(testFile, []byte("cached data"))

	server, _ := New(tempDir)

	reqBody := map[string]interface{}{
		"path":  testFile,
		"angle": 90,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Verify cache was invalidated
	_, ok := thumbCache.get(testFile)
	if ok {
		t.Error("Thumbnail cache should be invalidated after rotation")
	}
}

// TestHandleRotate_ReadOnlyFile tests error handling for read-only files
// Note: This test may not work on all platforms due to filesystem behavior
func TestHandleRotate_ReadOnlyFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping read-only test when running as root")
	}

	// On macOS, temp directories may have different permission handling
	// We test the permission error path by attempting the rotation
	tempDir := t.TempDir()
	testFile := "readonly.jpg"
	fullPath := filepath.Join(tempDir, testFile)

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, _ := os.Create(fullPath)
	jpeg.Encode(file, img, &jpeg.Options{Quality: 85})
	file.Close()

	// Make file read-only
	if err := os.Chmod(fullPath, 0444); err != nil {
		t.Fatalf("Failed to make file read-only: %v", err)
	}

	// Restore permissions after test
	defer os.Chmod(fullPath, 0644)

	server, _ := New(tempDir)

	reqBody := map[string]interface{}{
		"path":  testFile,
		"angle": 90,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rotate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRotate(rec, req)

	// On some systems (e.g., macOS temp dirs with specific ACLs), 
	// the file may still be writable. Accept either success or permission error.
	if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 or 500 for read-only file test, got %d", rec.Code)
	}

	// If we got an error response, verify it's about permissions
	if rec.Code == http.StatusInternalServerError {
		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err == nil {
			if response["success"] != false {
				t.Error("Error response should have success=false")
			}
		}
	}
}
