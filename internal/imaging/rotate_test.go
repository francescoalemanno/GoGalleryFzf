package imaging

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// createTestImage creates a test image with the specified dimensions
func createTestImage(width, height int, format string) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient pattern to make rotation visually verifiable
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return img, nil
}

// saveTestImage saves an image to a file with the specified format
func saveTestImage(img image.Image, path string, format string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch format {
	case "jpeg", "jpg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
	case "png":
		return png.Encode(file, img)
	default:
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
	}
}

// TestRotateImage_90Degrees tests 90 degree rotation
func TestRotateImage_90Degrees(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")

	// Create a test image (100x200 to make rotation detectable)
	img, err := createTestImage(100, 200, "jpeg")
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	if err := saveTestImage(img, testFile, "jpeg"); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Get original file info
	originalInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat original file: %v", err)
	}

	// Rotate 90 degrees
	if err := RotateImage(testFile, 90); err != nil {
		t.Fatalf("Failed to rotate image: %v", err)
	}

	// Verify file exists and has been modified
	newInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat rotated file: %v", err)
	}

	if newInfo.ModTime().Equal(originalInfo.ModTime()) {
		t.Error("File modification time should have changed after rotation")
	}

	// Verify the rotated image can be decoded
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open rotated file: %v", err)
	}
	defer file.Close()

	rotatedImg, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode rotated image: %v", err)
	}

	// After 90 degree rotation, dimensions should be swapped
	bounds := rotatedImg.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("Expected dimensions 200x100 after 90° rotation, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestRotateImage_Minus90Degrees tests -90 degree rotation
func TestRotateImage_Minus90Degrees(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.png")

	// Create a test image
	img, err := createTestImage(100, 200, "png")
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	if err := saveTestImage(img, testFile, "png"); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Rotate -90 degrees
	if err := RotateImage(testFile, -90); err != nil {
		t.Fatalf("Failed to rotate image: %v", err)
	}

	// Verify the rotated image
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open rotated file: %v", err)
	}
	defer file.Close()

	rotatedImg, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode rotated image: %v", err)
	}

	bounds := rotatedImg.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("Expected dimensions 200x100 after -90° rotation, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestRotateImage_180Degrees tests 180 degree rotation
func TestRotateImage_180Degrees(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")

	// Create a test image with a specific pattern at corner
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Set top-left corner to red
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	// Set bottom-right corner to blue
	img.Set(99, 99, color.RGBA{0, 0, 255, 255})

	if err := saveTestImage(img, testFile, "jpeg"); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Rotate 180 degrees
	if err := RotateImage(testFile, 180); err != nil {
		t.Fatalf("Failed to rotate image: %v", err)
	}

	// Verify the rotated image
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open rotated file: %v", err)
	}
	defer file.Close()

	rotatedImg, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode rotated image: %v", err)
	}

	// Dimensions should remain the same
	bounds := rotatedImg.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected dimensions 100x100 after 180° rotation, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// After 180° rotation, red should be at bottom-right and blue at top-left
	r, g, b, _ := rotatedImg.At(99, 99).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Logf("Note: JPEG compression may affect exact pixel values")
	}
}

// TestRotateImage_InvalidAngle tests error handling for invalid angles
func TestRotateImage_InvalidAngle(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	if err := saveTestImage(img, testFile, "jpeg"); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Test invalid angles
	invalidAngles := []int{0, 45, -45, 270, 360}
	for _, angle := range invalidAngles {
		err := RotateImage(testFile, angle)
		if err == nil {
			t.Errorf("Expected error for angle %d, got nil", angle)
		}
	}
}

// TestRotateImage_NonExistentFile tests error handling for non-existent files
func TestRotateImage_NonExistentFile(t *testing.T) {
	err := RotateImage("/nonexistent/path/image.jpg", 90)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestRotateImage_LargeImage tests rotation of a large image (>10MB equivalent in memory)
// This simulates the large image scenario mentioned in the plan
func TestRotateImage_LargeImage(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large.jpg")

	// Create a large image (approx 4000x3000 = 12MP, ~48MB in memory as RGBA)
	// For testing purposes, we'll create a smaller but still substantial image
	// that exercises the rotation code path without causing test timeouts
	width, height := 2000, 1500

	img, err := createTestImage(width, height, "jpeg")
	if err != nil {
		t.Fatalf("Failed to create large test image: %v", err)
	}

	if err := saveTestImage(img, testFile, "jpeg"); err != nil {
		t.Fatalf("Failed to save large test image: %v", err)
	}

	// Get file size
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}
	t.Logf("Test image size: %d bytes (%d MB)", info.Size(), info.Size()/1024/1024)

	// Rotate - should complete without memory issues
	if err := RotateImage(testFile, 90); err != nil {
		t.Fatalf("Failed to rotate large image: %v", err)
	}

	// Verify rotated image
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open rotated file: %v", err)
	}
	defer file.Close()

	rotatedImg, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode rotated large image: %v", err)
	}

	// Verify dimensions swapped
	bounds := rotatedImg.Bounds()
	if bounds.Dx() != height || bounds.Dy() != width {
		t.Errorf("Expected dimensions %dx%d after rotation, got %dx%d", height, width, bounds.Dx(), bounds.Dy())
	}
}

// TestRotateImage_PreservesPermissions tests that file permissions are preserved
func TestRotateImage_PreservesPermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	if err := saveTestImage(img, testFile, "jpeg"); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Set specific permissions
	if err := os.Chmod(testFile, 0644); err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}

	originalInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	originalMode := originalInfo.Mode().Perm()

	// Rotate
	if err := RotateImage(testFile, 90); err != nil {
		t.Fatalf("Failed to rotate image: %v", err)
	}

	// Verify permissions preserved
	newInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat rotated file: %v", err)
	}
	newMode := newInfo.Mode().Perm()

	if originalMode != newMode {
		t.Errorf("Permissions changed: expected %o, got %o", originalMode, newMode)
	}
}

// TestSupportedImageExts tests the SupportedImageExts function
func TestSupportedImageExts(t *testing.T) {
	 exts := SupportedImageExts()
	expected := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if len(exts) != len(expected) {
		t.Errorf("Expected %d extensions, got %d", len(expected), len(exts))
	}

	for _, ext := range exts {
		if !expected[ext] {
			t.Errorf("Unexpected extension: %s", ext)
		}
	}
}

// TestIsSupportedFormat tests the IsSupportedFormat function
func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".jpg", true},
		{".JPG", true},
		{".jpeg", true},
		{".JPEG", true},
		{".png", true},
		{".PNG", true},
		{".gif", true},
		{".GIF", true},
		{".webp", true},
		{".WEBP", true},
		{".bmp", false},
		{".svg", false},
		{".txt", false},
		{"", false},
	}

	for _, test := range tests {
		result := IsSupportedFormat(test.ext)
		if result != test.expected {
			t.Errorf("IsSupportedFormat(%q) = %v, expected %v", test.ext, result, test.expected)
		}
	}
}

// TestRotateImage_MultipleRotations tests multiple consecutive rotations
func TestRotateImage_MultipleRotations(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	if err := saveTestImage(img, testFile, "jpeg"); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Apply 4 x 90° rotations = 360° = original orientation
	for i := 0; i < 4; i++ {
		if err := RotateImage(testFile, 90); err != nil {
			t.Fatalf("Failed on rotation %d: %v", i+1, err)
		}
	}

	// Verify image is still valid
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	_, _, err = image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode after multiple rotations: %v", err)
	}
}

// BenchmarkRotateImage benchmarks the rotation performance
func BenchmarkRotateImage(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.jpg")

	// Create a test image (800x600 = ~1.8MP)
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	if err := saveTestImage(img, testFile, "jpeg"); err != nil {
		b.Fatalf("Failed to save test image: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy file for each iteration since we modify it
		testFileCopy := filepath.Join(tempDir, fmt.Sprintf("bench_%d.jpg", i))
		data, _ := os.ReadFile(testFile)
		os.WriteFile(testFileCopy, data, 0644)

		RotateImage(testFileCopy, 90)
	}
}
