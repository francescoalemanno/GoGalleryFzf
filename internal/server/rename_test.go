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
	"runtime"
	"testing"
)

// TestHandleRename_Success tests successful rename
func TestHandleRename_Success(t *testing.T) {
	tempDir := t.TempDir()
	oldFile := "test.jpg"
	newFile := "renamed.jpg"

	// Create a test file
	fullPath := filepath.Join(tempDir, oldFile)
	if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	server, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	reqBody := map[string]string{
		"path":    oldFile,
		"newName": newFile,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["success"] != true {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	if response["oldPath"] != oldFile {
		t.Errorf("Expected oldPath=%s, got %v", oldFile, response["oldPath"])
	}

	if response["newPath"] != newFile {
		t.Errorf("Expected newPath=%s, got %v", newFile, response["newPath"])
	}

	// Verify old file doesn't exist
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Error("Old file should not exist after rename")
	}

	// Verify new file exists
	newFullPath := filepath.Join(tempDir, newFile)
	if _, err := os.Stat(newFullPath); os.IsNotExist(err) {
		t.Error("New file should exist after rename")
	}
}

// TestHandleRename_InvalidMethod tests that only POST is allowed
func TestHandleRename_InvalidMethod(t *testing.T) {
	server, _ := New(t.TempDir())

	req := httptest.NewRequest("GET", "/api/rename", nil)
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}
}

// TestHandleRename_InvalidJSON tests error handling for invalid JSON
func TestHandleRename_InvalidJSON(t *testing.T) {
	server, _ := New(t.TempDir())

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

// TestHandleRename_PathTraversal tests path traversal protection
func TestHandleRename_PathTraversal(t *testing.T) {
	server, _ := New(t.TempDir())

	reqBody := map[string]string{
		"path":    "../../../etc/passwd",
		"newName": "test.txt",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rec.Code)
	}
}

// TestHandleRename_FileNotFound tests renaming a non-existent file
func TestHandleRename_FileNotFound(t *testing.T) {
	server, _ := New(t.TempDir())

	reqBody := map[string]string{
		"path":    "nonexistent.jpg",
		"newName": "newname.jpg",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

// TestHandleRename_DirectoryNotAllowed tests that directories cannot be renamed
func TestHandleRename_DirectoryNotAllowed(t *testing.T) {
	tempDir := t.TempDir()
	subDir := "subdir"

	// Create a subdirectory
	if err := os.Mkdir(filepath.Join(tempDir, subDir), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	server, _ := New(tempDir)

	reqBody := map[string]string{
		"path":    subDir,
		"newName": "renameddir",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

// TestHandleRename_InvalidCharacters tests invalid characters in new name
func TestHandleRename_InvalidCharacters(t *testing.T) {
	server, _ := New(t.TempDir())

	invalidNames := []string{
		"file/name.jpg",
		"file\\name.jpg",
		"file:name.jpg",
		"file*name.jpg",
		"file?name.jpg",
		"file\"name.jpg",
		"file<name.jpg",
		"file>name.jpg",
		"file|name.jpg",
	}

	for _, name := range invalidNames {
		reqBody := map[string]string{
			"path":    "test.jpg",
			"newName": name,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.HandleRename(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for %q, got %d", name, rec.Code)
		}
	}
}

// TestHandleRename_EmptyName tests empty new name
func TestHandleRename_EmptyName(t *testing.T) {
	server, _ := New(t.TempDir())

	reqBody := map[string]string{
		"path":    "test.jpg",
		"newName": "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

// TestHandleRename_ReservedNames tests reserved names . and ..
func TestHandleRename_ReservedNames(t *testing.T) {
	server, _ := New(t.TempDir())

	reservedNames := []string{".", ".."}
	for _, name := range reservedNames {
		reqBody := map[string]string{
			"path":    "test.jpg",
			"newName": name,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.HandleRename(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for %q, got %d", name, rec.Code)
		}
	}
}

// TestHandleRename_FileAlreadyExists tests renaming to an existing file
func TestHandleRename_FileAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create two test files
	oldFile := "old.jpg"
	newFile := "new.jpg"

	if err := os.WriteFile(filepath.Join(tempDir, oldFile), []byte("old content"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, newFile), []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	server, _ := New(tempDir)

	reqBody := map[string]string{
		"path":    oldFile,
		"newName": newFile,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rec.Code)
	}
}

// TestHandleRename_CaseOnly tests renaming to same name different case
// This is important on case-insensitive filesystems (macOS, Windows)
func TestHandleRename_CaseOnly(t *testing.T) {
	tempDir := t.TempDir()
	oldFile := "test.jpg"
	newFile := "TEST.jpg"

	// Create a test file
	fullPath := filepath.Join(tempDir, oldFile)
	if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	server, _ := New(tempDir)

	reqBody := map[string]string{
		"path":    oldFile,
		"newName": newFile,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	// On case-insensitive filesystems, this might return conflict (409)
	// On case-sensitive filesystems, this should succeed (200)
	// The test accepts either outcome
	if rec.Code != http.StatusOK && rec.Code != http.StatusConflict {
		t.Errorf("Expected status 200 or 409 on %s, got %d", runtime.GOOS, rec.Code)
	}
}

// TestHandleRename_NameTooLong tests name length limit
func TestHandleRename_NameTooLong(t *testing.T) {
	server, _ := New(t.TempDir())

	// Create a name longer than 255 characters
	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}
	longName += ".jpg"

	reqBody := map[string]string{
		"path":    "test.jpg",
		"newName": longName,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

// TestHandleRename_InSubdirectory tests renaming a file in a subdirectory
func TestHandleRename_InSubdirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create subdirectory and file
	subDir := "subdir"
	if err := os.Mkdir(filepath.Join(tempDir, subDir), 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	oldPath := filepath.Join(subDir, "test.jpg")
	newName := "renamed.jpg"
	expectedNewPath := filepath.Join(subDir, newName)

	if err := os.WriteFile(filepath.Join(tempDir, oldPath), []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	server, _ := New(tempDir)

	reqBody := map[string]string{
		"path":    oldPath,
		"newName": newName,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["newPath"] != expectedNewPath {
		t.Errorf("Expected newPath=%s, got %v", expectedNewPath, response["newPath"])
	}

	// Verify new file exists
	newFullPath := filepath.Join(tempDir, expectedNewPath)
	if _, err := os.Stat(newFullPath); os.IsNotExist(err) {
		t.Error("New file should exist after rename")
	}
}

// TestHandleRename_InvalidatesThumbnail tests that rename invalidates thumbnail cache
func TestHandleRename_InvalidatesThumbnail(t *testing.T) {
	tempDir := t.TempDir()
	oldFile := "test.jpg"

	// Create a test image
	fullPath := filepath.Join(tempDir, oldFile)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, _ := os.Create(fullPath)
	jpeg.Encode(file, img, &jpeg.Options{Quality: 85})
	file.Close()

	// Pre-populate cache
	thumbCache.set(oldFile, []byte("cached data"))

	server, _ := New(tempDir)

	reqBody := map[string]string{
		"path":    oldFile,
		"newName": "renamed.jpg",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/rename", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.HandleRename(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Verify old cache entry was invalidated
	_, ok := thumbCache.get(oldFile)
	if ok {
		t.Error("Old thumbnail cache should be invalidated after rename")
	}
}

// TestValidateFileName tests the filename validation function
func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid.jpg", false},
		{"file name with spaces.png", false},
		{"file-name_123.gif", false},
		{"", true},                              // empty
		{"file/name.jpg", true},                 // slash
		{"file\\\\name.jpg", true},              // backslash
		{"file:name.jpg", true},                 // colon
		{"file*name.jpg", true},                 // asterisk
		{"file?name.jpg", true},                 // question mark
		{"file\"name.jpg", true},                // quote
		{"file<name.jpg", true},                 // less than
		{"file>name.jpg", true},                 // greater than
		{"file|name.jpg", true},                 // pipe
		{".", true},                             // current dir
		{"..", true},                            // parent dir
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFileName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

// TestValidateFileName_TooLong tests the 255 character limit
func TestValidateFileName_TooLong(t *testing.T) {
	// Create a name with exactly 255 characters
	name := ""
	for i := 0; i < 255; i++ {
		name += "a"
	}

	// Should be valid at exactly 255
	if err := validateFileName(name); err != nil {
		t.Errorf("Name with exactly 255 chars should be valid: %v", err)
	}

	// 256 should fail
	name += "a"
	if err := validateFileName(name); err == nil {
		t.Error("Name with 256 chars should be invalid")
	}
}
