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
	"gallery/internal/models"
)

const DefaultPageSize = 100

type GalleryServer struct {
	rootDir string
	srv     *http.Server
}

func New(rootDir string) (*GalleryServer, error) {
	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("invalid root directory: %w", err)
	}
	return &GalleryServer{rootDir: absPath}, nil
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

func (s *GalleryServer) RootDir() string {
	return s.rootDir
}

func (s *GalleryServer) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.New("gallery").Parse(HTMLTemplate))
	tmpl.Execute(w, nil)
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

	cleanPath := filepath.Clean(folder)
	fullPath := filepath.Join(s.rootDir, cleanPath)

	if !strings.HasPrefix(fullPath, s.rootDir) {
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

func (s *GalleryServer) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "."
	}

	page, limit := parsePagination(r)

	cleanPath := filepath.Clean(folder)
	fullPath := filepath.Join(s.rootDir, cleanPath)

	if !strings.HasPrefix(fullPath, s.rootDir) {
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

	cleanPath := filepath.Clean(folder)
	fullPath := filepath.Join(s.rootDir, cleanPath)

	if !strings.HasPrefix(fullPath, s.rootDir) {
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
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.rootDir, cleanPath)

	if !strings.HasPrefix(fullPath, s.rootDir) {
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

	// Handle range requests for video/audio streaming
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		// No range request - serve entire file
		w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
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

	file.Seek(start, io.SeekStart)
	io.CopyN(w, file, contentLength)
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
