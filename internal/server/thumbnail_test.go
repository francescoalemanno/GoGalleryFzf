package server

import (
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// createTestJPEG creates a simple JPEG file for testing
func createTestJPEG(t *testing.T, path string, width, height int) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, image.Black)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("Failed to encode JPEG: %v", err)
	}
}

// TestThumbnailCache_SetAndGet tests basic cache operations
func TestThumbnailCache_SetAndGet(t *testing.T) {
	cache := &ThumbnailCache{
		cache: make(map[string][]byte),
		order: make([]string, 0),
	}

	// Test set and get
	data := []byte("test thumbnail data")
	cache.set("test.jpg", data)

	retrieved, ok := cache.get("test.jpg")
	if !ok {
		t.Error("Expected to find cached item")
	}
	if string(retrieved) != string(data) {
		t.Error("Retrieved data doesn't match")
	}

	// Test get for non-existent key
	_, ok = cache.get("nonexistent.jpg")
	if ok {
		t.Error("Should not find non-existent key")
	}
}

// TestThumbnailCache_Delete tests cache deletion
func TestThumbnailCache_Delete(t *testing.T) {
	cache := &ThumbnailCache{
		cache: make(map[string][]byte),
		order: make([]string, 0),
	}

	// Add items
	cache.set("image1.jpg", []byte("data1"))
	cache.set("image2.jpg", []byte("data2"))
	cache.set("image3.jpg", []byte("data3"))

	// Delete middle item
	cache.delete("image2.jpg")

	// Verify deleted
	_, ok := cache.get("image2.jpg")
	if ok {
		t.Error("Deleted item should not be found")
	}

	// Verify others still exist
	_, ok = cache.get("image1.jpg")
	if !ok {
		t.Error("image1.jpg should still be cached")
	}
	_, ok = cache.get("image3.jpg")
	if !ok {
		t.Error("image3.jpg should still be cached")
	}
}

// TestThumbnailCache_LRU tests LRU eviction
func TestThumbnailCache_LRU(t *testing.T) {
	cache := &ThumbnailCache{
		cache: make(map[string][]byte),
		order: make([]string, 0),
	}

	// Add more items than maxCacheSize
	for i := 0; i < maxCacheSize+5; i++ {
		key := filepath.Join("path", string(rune('a'+i%26))+".jpg")
		cache.set(key, []byte("data"))
	}

	// Cache should not exceed max size
	if len(cache.cache) > maxCacheSize {
		t.Errorf("Cache size %d exceeds max %d", len(cache.cache), maxCacheSize)
	}
	if len(cache.order) > maxCacheSize {
		t.Errorf("Order size %d exceeds max %d", len(cache.order), maxCacheSize)
	}
}

// TestThumbnailCache_InvalidateOnRotation tests that thumbnail cache is invalidated after rotation
func TestThumbnailCache_InvalidateOnRotation(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")

	// Create a test image
	createTestJPEG(t, testFile, 400, 300)

	// Add to cache
	data := []byte("cached thumbnail")
	thumbCache.set("test.jpg", data)

	// Verify it's cached
	cached, ok := thumbCache.get("test.jpg")
	if !ok {
		t.Fatal("Item should be in cache")
	}
	if string(cached) != string(data) {
		t.Error("Cached data doesn't match")
	}

	// Invalidate cache (simulating what happens after rotation)
	invalidateThumbnailCache("test.jpg")

	// Verify it's no longer cached
	_, ok = thumbCache.get("test.jpg")
	if ok {
		t.Error("Cache should be invalidated after rotation")
	}
}

// TestServeThumbnail_CacheHit tests serving a cached thumbnail
func TestServeThumbnail_CacheHit(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test image with relative path
	testFile := "test.jpg"
	fullPath := filepath.Join(tempDir, testFile)
	createTestJPEG(t, fullPath, 400, 300)

	// Pre-populate cache with a marker (using relative path as key)
	thumbCache.set(testFile, []byte("cached data"))

	server, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/thumb/"+testFile, nil)
	rec := httptest.NewRecorder()

	server.ServeThumbnail(rec, req, testFile)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Verify we got the cached data
	if string(rec.Body.Bytes()) != "cached data" {
		t.Error("Did not get cached data")
	}

	// Cleanup
	invalidateThumbnailCache(testFile)
}

// TestServeThumbnail_CacheMiss tests generating thumbnail on cache miss
func TestServeThumbnail_CacheMiss(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test image with relative path
	testFile := "test.jpg"
	fullPath := filepath.Join(tempDir, testFile)
	createTestJPEG(t, fullPath, 400, 300)

	server, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Ensure not cached
	invalidateThumbnailCache(testFile)

	req := httptest.NewRequest("GET", "/thumb/"+testFile, nil)
	rec := httptest.NewRecorder()

	server.ServeThumbnail(rec, req, testFile)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "image/jpeg" {
		t.Errorf("Expected content type image/jpeg, got %s", contentType)
	}

	// Verify it was cached
	_, ok := thumbCache.get(testFile)
	if !ok {
		t.Error("Thumbnail should be cached after generation")
	}

	// Cleanup
	invalidateThumbnailCache(testFile)
}

// TestServeThumbnail_PathTraversal tests path traversal protection
func TestServeThumbnail_PathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	server, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/thumb/../../../etc/passwd", nil)
	rec := httptest.NewRecorder()

	server.ServeThumbnail(rec, req, "../../../etc/passwd")

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for path traversal, got %d", rec.Code)
	}
}

// TestServeThumbnail_NonImage tests serving a non-image file
func TestServeThumbnail_NonImage(t *testing.T) {
	tempDir := t.TempDir()

	// Create a text file with relative path
	testFile := "test.txt"
	fullPath := filepath.Join(tempDir, testFile)
	if err := os.WriteFile(fullPath, []byte("not an image"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	server, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/thumb/"+testFile, nil)
	rec := httptest.NewRecorder()

	server.ServeThumbnail(rec, req, testFile)

	// Should serve the raw file for non-images
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// TestGenerateThumbnail tests the thumbnail generation function
func TestGenerateThumbnail(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")

	// Create a large test image
	createTestJPEG(t, testFile, 1000, 800)

	data, err := generateThumbnail(testFile)
	if err != nil {
		t.Fatalf("Failed to generate thumbnail: %v", err)
	}

	if len(data) == 0 {
		t.Error("Generated thumbnail is empty")
	}

	// Verify it's a valid JPEG by checking magic bytes
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		t.Error("Generated data is not a valid JPEG")
	}
}

// TestInvalidateThumbnailCache_Global tests the global cache invalidation
func TestInvalidateThumbnailCache_Global(t *testing.T) {
	// Add multiple items
	thumbCache.set("image1.jpg", []byte("data1"))
	thumbCache.set("image2.jpg", []byte("data2"))
	thumbCache.set("image3.jpg", []byte("data3"))

	// Invalidate specific item
	invalidateThumbnailCache("image2.jpg")

	// Verify only that item is removed
	_, ok := thumbCache.get("image1.jpg")
	if !ok {
		t.Error("image1.jpg should still be cached")
	}

	_, ok = thumbCache.get("image2.jpg")
	if ok {
		t.Error("image2.jpg should be invalidated")
	}

	_, ok = thumbCache.get("image3.jpg")
	if !ok {
		t.Error("image3.jpg should still be cached")
	}

	// Cleanup
	invalidateThumbnailCache("image1.jpg")
	invalidateThumbnailCache("image3.jpg")
}
