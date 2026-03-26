package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gallery/internal/fzf"
	"gallery/internal/imaging"
	"gallery/internal/models"
	"gallery/internal/version"
)

const DefaultPageSize = 100

type GalleryServer struct {
	rootDir    string
	rootDirResolved string // rootDir with symlinks resolved
	srv        *http.Server
}

func New(rootDir string) (*GalleryServer, error) {
	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("invalid root directory: %w", err)
	}
	
	// Resolve symlinks in rootDir for consistent path comparison
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If we can't resolve symlinks, use the absolute path
		resolvedPath = absPath
	}
	
	return &GalleryServer{
		rootDir:    absPath,
		rootDirResolved: resolvedPath,
	}, nil
}

func (s *GalleryServer) SetServer(srv *http.Server) {
	s.srv = srv
}

func (s *GalleryServer) HandleShutdown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "shutting_down"})

	// Shutdown the server gracefully
	go func() {
		if s.srv != nil {
			s.srv.Close()
		}
	}()
}

// HandleRotate handles image rotation requests
func (s *GalleryServer) HandleRotate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path  string `json:"path"`
		Angle int    `json:"angle"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON body",
		})
		return
	}

	// Validate path (prevent directory traversal)
	fullPath, cleanPath, err := s.resolveAndValidatePath(req.Path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	// Check if file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "File not found",
		})
		return
	}

	if info.IsDir() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Path is a directory",
		})
		return
	}

	// Check if file is a supported image
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if !imaging.IsSupportedFormat(ext) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unsupported image format",
		})
		return
	}

	// Perform rotation
	if err := imaging.RotateImage(fullPath, req.Angle); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Invalidate thumbnail cache
	invalidateThumbnailCache(cleanPath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"path":    cleanPath,
		"angle":   req.Angle,
	})
}

// invalidateThumbnailCache removes a specific path from the thumbnail cache
func invalidateThumbnailCache(path string) {
	thumbCache.delete(path)
}

// HandleRename handles file rename requests
func (s *GalleryServer) HandleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path    string `json:"path"`
		NewName string `json:"newName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON body",
		})
		return
	}

	// Validate newName
	if err := validateFileName(req.NewName); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate path (prevent directory traversal)
	fullPath, cleanPath, err := s.resolveAndValidatePath(req.Path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	// Check if file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "File not found",
		})
		return
	}

	if info.IsDir() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Cannot rename directories",
		})
		return
	}

	// Build new path
	dir := filepath.Dir(cleanPath)
	newPath := filepath.Join(dir, req.NewName)
	newFullPath, _, err := s.resolveAndValidatePath(newPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid new name",
		})
		return
	}

	// Check if destination already exists
	if _, err := os.Stat(newFullPath); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "File already exists",
		})
		return
	}

	// Perform rename
	if err := os.Rename(fullPath, newFullPath); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Invalidate thumbnail cache for old path
	invalidateThumbnailCache(cleanPath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"oldPath": cleanPath,
		"newPath": newPath,
	})
}

// validateFileName checks if a filename is valid
func validateFileName(name string) error {
	if name == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}

	// Check for invalid characters: / \ : * ? " < > |
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("filename contains invalid character: %s", char)
		}
	}

	// Check for . and .. which are reserved
	if name == "." || name == ".." {
		return fmt.Errorf("invalid filename")
	}

	return nil
}

func (s *GalleryServer) RootDir() string {
	return s.rootDir
}

func (s *GalleryServer) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.New("gallery").Parse(HTMLTemplate))
	data := map[string]string{
		"Version": version.Version(),
	}
	tmpl.Execute(w, data)
}

func parsePagination(r *http.Request) (page, limit int) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, _ = strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ = strconv.Atoi(limitStr)
	if limit < 1 || limit > 500 {
		limit = DefaultPageSize
	}

	return page, limit
}

func paginateFiles(files []models.FileInfo, page, limit int) models.PaginatedResponse {
	total := len(files)
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	start := (page - 1) * limit
	end := start + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	// Always return empty slice, never nil (for JSON array)
	paginated := []models.FileInfo{}
	if start < total {
		paginated = files[start:end]
	}

	return models.PaginatedResponse{
		Files:      paginated,
		Total:      total,
		Page:       page,
		Limit:      limit,
		HasMore:    end < total,
		TotalPages: totalPages,
	}
}

func (s *GalleryServer) HandleFiles(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "."
	}

	page, limit := parsePagination(r)

	fullPath, cleanPath, err := s.resolveAndValidatePath(folder)
	if err != nil {
		http.Error(w, "Accesso non consentito", http.StatusForbidden)
		return
	}

	// Check if directory exists
	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "Cartella non trovata", http.StatusNotFound)
		return
	}
	if !info.IsDir() {
		http.Error(w, "Non è una cartella", http.StatusBadRequest)
		return
	}

	files, err := s.scanDirectory(fullPath, cleanPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := paginateFiles(files, page, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSearch handles search requests and returns paginated results
// sorted by fzf relevance score (highest first)
func (s *GalleryServer) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "."
	}

	page, limit := parsePagination(r)

	fullPath, cleanPath, err := s.resolveAndValidatePath(folder)
	if err != nil {
		http.Error(w, "Accesso non consentito", http.StatusForbidden)
		return
	}

	allFiles, err := s.scanDirectoryRecursive(fullPath, cleanPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build file list for fzf matching (skip directories)
	var fileList []models.FileInfo
	var filePaths []string // Use full paths for matching
	fileIndex := make(map[string]int) // path -> index in fileList

	for _, f := range allFiles {
		if f.IsDir {
			continue
		}
		fileIndex[f.Path] = len(fileList)
		fileList = append(fileList, f)
		filePaths = append(filePaths, f.Path) // Use full path for search
	}

	// Use fzf algorithm for matching and scoring
	var results []models.FileInfo
	if query == "" {
		results = fileList
	} else {
		matches := fzf.FuzzySearch(query, filePaths, 0)
		for _, match := range matches {
			if idx, ok := fileIndex[match.Text]; ok {
				results = append(results, fileList[idx])
			}
		}
	}

	response := paginateFiles(results, page, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *GalleryServer) HandleFolders(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "."
	}

	fullPath, cleanPath, err := s.resolveAndValidatePath(folder)
	if err != nil {
		http.Error(w, "Accesso non consentito", http.StatusForbidden)
		return
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var folders []models.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			info, _ := entry.Info()
			relPath := filepath.Join(cleanPath, entry.Name())
			folders = append(folders, models.FileInfo{
				Name:    entry.Name(),
				Path:    relPath,
				ModTime: info.ModTime(),
				IsDir:   true,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folders)
}

func (s *GalleryServer) HandleRaw(w http.ResponseWriter, r *http.Request) {
	path, _ := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/raw/"))
	s.serveFile(w, r, path)
}

func (s *GalleryServer) HandleThumb(w http.ResponseWriter, r *http.Request) {
	path, _ := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/thumb/"))
	s.ServeThumbnail(w, r, path)
}

func (s *GalleryServer) serveFile(w http.ResponseWriter, r *http.Request, path string) {
	fullPath, _, err := s.resolveAndValidatePath(path)
	if err != nil {
		http.Error(w, "Accesso non consentito", http.StatusForbidden)
		return
	}

	// Open file to get info and serve with range support
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if stat.IsDir() {
		http.Error(w, "Accesso non consentito", http.StatusForbidden)
		return
	}

	// Detect content type
	contentType := getContentType(filepath.Ext(fullPath))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")

	// Enable keep-alive for media streaming (video/audio)
	if isVideoExt(filepath.Ext(fullPath)) || isAudioExt(filepath.Ext(fullPath)) {
		w.Header().Set("Connection", "keep-alive")
	}

	// Handle range requests for video/audio streaming
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		// No range request - serve entire file
		w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, file)
		// Ignore write errors (client disconnects are common during streaming)
		return
	}

	// Parse range header
	start, end, err := parseRange(rangeHeader, stat.Size())
	if err != nil {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", stat.Size()))
		http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Serve partial content
	contentLength := end - start + 1
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, stat.Size()))
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.WriteHeader(http.StatusPartialContent)

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return
	}
	_, err = io.CopyN(w, file, contentLength)
	// Ignore write errors (client disconnects are common during streaming)
}

func getContentType(ext string) string {
	ext = strings.ToLower(ext)
	mimeTypes := map[string]string{
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mov":  "video/quicktime",
		".avi":  "video/x-msvideo",
		".mkv":  "video/x-matroska",
		".flv":  "video/x-flv",
		".wmv":  "video/x-ms-wmv",
		".m4v":  "video/mp4",
		".3gp":  "video/3gpp",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
		".wma":  "audio/x-ms-wma",
		".opus": "audio/opus",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
	}
	if ct, ok := mimeTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func parseRange(rangeHeader string, fileSize int64) (start, end int64, err error) {
	const prefix = "bytes="
	if !strings.HasPrefix(rangeHeader, prefix) {
		return 0, 0, fmt.Errorf("invalid range header")
	}

	rangeStr := strings.TrimPrefix(rangeHeader, prefix)
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	startStr, endStr := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

	if startStr == "" {
		// Suffix range: -500 means last 500 bytes
		suffix, _ := strconv.ParseInt(endStr, 10, 64)
		start = fileSize - suffix
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		if endStr == "" {
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(endStr, 10, 64)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	if start < 0 {
		start = 0
	}
	if end >= fileSize {
		end = fileSize - 1
	}
	if start > end {
		return 0, 0, fmt.Errorf("invalid range")
	}

	return start, end, nil
}

// resolveAndValidatePath resolves the given path relative to rootDir and validates it.
// It returns the full filesystem path, the clean relative path, and an error if validation fails.
// This handles symlinks and complex path scenarios properly.
func (s *GalleryServer) resolveAndValidatePath(userPath string) (fullPath string, cleanRelPath string, err error) {
	// Clean the input path
	cleanRelPath = filepath.Clean(userPath)

	// Prevent path traversal at the input level - reject paths that escape the root
	if strings.HasPrefix(cleanRelPath, "..") || strings.Contains(cleanRelPath, string(filepath.Separator)+"..") {
		return "", "", fmt.Errorf("path traversal detected")
	}

	// Join with rootDir to get the full path
	fullPath = filepath.Join(s.rootDir, cleanRelPath)

	// Resolve symlinks in the full path (this also gives us the canonical absolute path)
	resolvedPath, resolveErr := filepath.EvalSymlinks(fullPath)
	if resolveErr != nil {
		// If the path doesn't exist, we can't resolve symlinks.
		// Try to resolve the parent directory to validate the path.
		parentDir := filepath.Dir(fullPath)
		resolvedParent, parentErr := filepath.EvalSymlinks(parentDir)
		if parentErr == nil {
			// Validate that the resolved parent is within resolved root
			rootPrefix := s.rootDirResolved + string(filepath.Separator)
			if resolvedParent != s.rootDirResolved && !strings.HasPrefix(resolvedParent+string(filepath.Separator), rootPrefix) {
				return "", "", fmt.Errorf("path traversal detected")
			}
			// Parent is valid. Construct the resolved full path by joining resolved parent with the file name
			// This ensures the resolved path has the same prefix as rootDirResolved
			baseName := filepath.Base(fullPath)
			resolvedPath = filepath.Join(resolvedParent, baseName)
		} else {
			// Can't resolve parent either - fall back to basic prefix check on unresolved paths
			if !strings.HasPrefix(fullPath, s.rootDir+string(filepath.Separator)) && fullPath != s.rootDir {
				return "", "", fmt.Errorf("path traversal detected")
			}
			resolvedPath = fullPath
		}
	}

	// Ensure the resolved path is within the resolved root directory
	// Add trailing separator to prevent partial matches like /rootDirFoo matching /rootDir
	rootPrefix := s.rootDirResolved + string(filepath.Separator)
	if resolvedPath != s.rootDirResolved && !strings.HasPrefix(resolvedPath+string(filepath.Separator), rootPrefix) {
		return "", "", fmt.Errorf("path traversal detected")
	}

	return fullPath, cleanRelPath, nil
}

func (s *GalleryServer) scanDirectory(dirPath, relPath string) ([]models.FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []models.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Mostra solo cartelle e file multimediali (immagini + video + audio)
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		isImage := isImageExt(ext)
		isVideo := isVideoExt(ext)
		isAudio := isAudioExt(ext)
		isMedia := isImage || isVideo || isAudio

		// Salta file non multimediali
		if !entry.IsDir() && !isMedia {
			continue
		}

		entryRelPath := filepath.Join(relPath, entry.Name())
		files = append(files, models.FileInfo{
			Name:    entry.Name(),
			Path:    entryRelPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
			IsImage: isImage,
			IsVideo: isVideo,
			IsAudio: isAudio,
			Ext:     ext,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	return files, nil
}

func (s *GalleryServer) scanDirectoryRecursive(dirPath, relPath string) ([]models.FileInfo, error) {
	var allFiles []models.FileInfo

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Detect file type for all files
		ext := strings.ToLower(filepath.Ext(d.Name()))
		isImage := isImageExt(ext)
		isVideo := isVideoExt(ext)
		isAudio := isAudioExt(ext)
		isMedia := isImage || isVideo || isAudio

		// Salta file non multimediali
		if !d.IsDir() && !isMedia {
			return nil
		}

		rel, _ := filepath.Rel(s.rootDir, path)
		allFiles = append(allFiles, models.FileInfo{
			Name:    d.Name(),
			Path:    rel,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   d.IsDir(),
			IsImage: isImage,
			IsVideo: isVideo,
			IsAudio: isAudio,
			Ext:     ext,
		})

		return nil
	})

	return allFiles, err
}

func isImageExt(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".ico"}
	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

func isVideoExt(ext string) bool {
	videoExts := []string{".mp4", ".webm", ".mov", ".avi", ".mkv", ".flv", ".wmv", ".m4v", ".3gp"}
	for _, e := range videoExts {
		if ext == e {
			return true
		}
	}
	return false
}

func isAudioExt(ext string) bool {
	audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a", ".wma", ".opus"}
	for _, e := range audioExts {
		if ext == e {
			return true
		}
	}
	return false
}
