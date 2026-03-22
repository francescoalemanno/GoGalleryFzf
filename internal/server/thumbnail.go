package server

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

const (
	thumbWidth  = 400
	thumbHeight = 300
	maxCacheSize = 100 // Maximum number of cached thumbnails in memory
)

// ThumbnailCache manages cached thumbnails
type ThumbnailCache struct {
	mu     sync.RWMutex
	cache  map[string][]byte
	order  []string // LRU tracking
}

var thumbCache = &ThumbnailCache{
	cache: make(map[string][]byte),
	order: make([]string, 0),
}

func (c *ThumbnailCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.cache[key]
	return data, ok
}

func (c *ThumbnailCache) set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, remove from order
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}

	// Evict oldest if at capacity
	for len(c.order) >= maxCacheSize {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.cache, oldest)
	}

	c.cache[key] = data
	c.order = append(c.order, key)
}

func generateThumbnail(imgPath string) ([]byte, error) {
	file, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Resize maintaining aspect ratio
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	// Calculate new dimensions preserving aspect ratio
	var newWidth, newHeight uint
	if width*thumbHeight > height*thumbWidth {
		// Width is the limiting factor
		newWidth = thumbWidth
		newHeight = uint(height * thumbWidth / width)
	} else {
		// Height is the limiting factor
		newHeight = thumbHeight
		newWidth = uint(width * thumbHeight / height)
	}

	resized := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	// Encode to JPEG for smaller size using temp file
	tempFile, err := os.CreateTemp("", "thumb-*.jpg")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	err = jpeg.Encode(tempFile, resized, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, err
	}

	// Read the file back
	return os.ReadFile(tempFile.Name())
}

// ServeThumbnail generates and serves a thumbnail
func (s *GalleryServer) ServeThumbnail(w http.ResponseWriter, r *http.Request, path string) {
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.rootDir, cleanPath)

	if !strings.HasPrefix(fullPath, s.rootDir) {
		http.Error(w, "Accesso non consentito", http.StatusForbidden)
		return
	}

	ext := strings.ToLower(filepath.Ext(path))

	// For non-image files, serve raw
	if !isImageExt(ext) {
		http.ServeFile(w, r, fullPath)
		return
	}

	// Check cache
	cacheKey := cleanPath
	if data, ok := thumbCache.get(cacheKey); ok {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(data)
		return
	}

	// Generate thumbnail
	data, err := generateThumbnail(fullPath)
	if err != nil {
		// Fallback to original on error
		http.ServeFile(w, r, fullPath)
		return
	}

	// Cache it
	thumbCache.set(cacheKey, data)

	// Serve
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

// ThumbnailEncoder encodes thumbnail with proper format
func encodeThumbnail(img image.Image, ext string) ([]byte, error) {
	switch ext {
	case ".png":
		// Use temp file for PNG
		tempFile, err := os.CreateTemp("", "thumb-*.png")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()
		err = png.Encode(tempFile, img)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(tempFile.Name())
	case ".gif":
		tempFile, err := os.CreateTemp("", "thumb-*.gif")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()
		err = gif.Encode(tempFile, img, nil)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(tempFile.Name())
	default:
		// JPEG for everything else
		tempFile, err := os.CreateTemp("", "thumb-*.jpg")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()
		err = jpeg.Encode(tempFile, img, &jpeg.Options{Quality: 85})
		if err != nil {
			return nil, err
		}
		return os.ReadFile(tempFile.Name())
	}
}
