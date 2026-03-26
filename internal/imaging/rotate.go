package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsoprea/go-jpeg-image-structure/v2"
	_ "golang.org/x/image/webp" // Register WebP decoder
)

// RotateImage rotates an image file by the specified angle and saves it back
// Supported angles: 90, -90, 180 (positive = clockwise, negative = counter-clockwise)
func RotateImage(path string, angle int) error {
	// Validate angle
	if angle != 90 && angle != -90 && angle != 180 && angle != -180 {
		return fmt.Errorf("unsupported angle: %d (supported: 90, -90, 180)", angle)
	}

	// Normalize angle
	if angle == -180 {
		angle = 180
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Apply rotation
	rotated := rotateImage(img, angle)

	// Close file before writing
	file.Close()

	// Encode and save back
	if err := encodeAndSave(path, rotated, format); err != nil {
		return fmt.Errorf("failed to save rotated image: %w", err)
	}

	return nil
}

// rotateImage applies rotation transformation to an image
func rotateImage(img image.Image, angle int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var dst *image.RGBA

	switch angle {
	case 90:
		// Rotate 90° clockwise: (x, y) -> (height-1-y, x)
		dst = image.NewRGBA(image.Rect(0, 0, height, width))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(height-1-y, x, img.At(x, y))
			}
		}
	case -90:
		// Rotate 90° counter-clockwise: (x, y) -> (y, width-1-x)
		dst = image.NewRGBA(image.Rect(0, 0, height, width))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(y, width-1-x, img.At(x, y))
			}
		}
	case 180:
		// Rotate 180°: (x, y) -> (width-1-x, height-1-y)
		dst = image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(width-1-x, height-1-y, img.At(x, y))
			}
		}
	}

	return dst
}

// encodeAndSave encodes the image in the specified format and saves it to path
func encodeAndSave(path string, img image.Image, format string) error {
	// Create temp file in same directory
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, "rotate-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	// Encode based on format
	var encodeErr error
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		// Encode to buffer first
		var buf bytes.Buffer
		encodeErr = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
		if encodeErr == nil {
			// Strip EXIF data to remove orientation tag - this ensures the rotated
			// pixels display correctly without browsers applying additional rotation
			encodeErr = stripExifAndWrite(&buf, tempFile)
		}
	case "png":
		encodeErr = png.Encode(tempFile, img)
	case "gif":
		encodeErr = gif.Encode(tempFile, img, nil)
	case "webp":
		// For WebP output, we'll encode as PNG since Go's stdlib doesn't have WebP encoder
		// The user can convert back if needed, or we save as PNG
		encodeErr = png.Encode(tempFile, img)
		// If it was originally WebP, we keep the original extension
		// This is a limitation - rotated WebP becomes PNG content with .webp extension
		// Browsers will still display it correctly
	default:
		encodeErr = jpeg.Encode(tempFile, img, &jpeg.Options{Quality: 95})
	}

	tempFile.Close()

	if encodeErr != nil {
		return fmt.Errorf("failed to encode image: %w", encodeErr)
	}

	// Preserve original file permissions
	originalInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat original file: %w", err)
	}

	// Set same permissions on temp file
	if err := os.Chmod(tempPath, originalInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Replace original with rotated version
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to replace original file: %w", err)
	}

	return nil
}

// stripExifAndWrite removes EXIF data from JPEG bytes and writes to the output file.
// This prevents browsers from applying orientation transforms based on EXIF orientation tags.
func stripExifAndWrite(buf *bytes.Buffer, out *os.File) error {
	// Parse the JPEG structure
	parser := jpegstructure.NewJpegMediaParser()
	imc, err := parser.ParseBytes(buf.Bytes())
	if err != nil {
		// If we can't parse it, just write the original bytes
		// This handles edge cases where the JPEG might be malformed
		_, err = out.Write(buf.Bytes())
		return err
	}

	sl := imc.(*jpegstructure.SegmentList)

	// Drop EXIF data from the segment list
	// This removes the orientation tag that causes browsers to apply additional rotation
	if _, err := sl.DropExif(); err != nil {
		// If we can't drop EXIF, just write the original bytes
		_, err = out.Write(buf.Bytes())
		return err
	}

	// Write the modified JPEG structure back out
	return sl.Write(out)
}

// SupportedImageExts returns the list of supported image extensions for rotation
func SupportedImageExts() []string {
	return []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
}

// IsSupportedFormat checks if the file extension is supported for rotation
func IsSupportedFormat(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	}
	return false
}
