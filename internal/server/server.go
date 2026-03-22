package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
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

	var paginated []models.FileInfo
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
	var fileNames []string
	fileIndex := make(map[string]int) // name -> index in fileList

	for _, f := range allFiles {
		if f.IsDir {
			continue
		}
		fileIndex[f.Name] = len(fileList)
		fileList = append(fileList, f)
		fileNames = append(fileNames, f.Name)
	}

	// Use fzf algorithm for matching and scoring
	var results []models.FileInfo
	if query == "" {
		results = fileList
	} else {
		matches := fzf.FuzzySearch(query, fileNames, 0)
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
	path := strings.TrimPrefix(r.URL.Path, "/raw/")
	s.serveFile(w, r, path)
}

func (s *GalleryServer) HandleThumb(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/thumb/")
	s.ServeThumbnail(w, r, path)
}

func (s *GalleryServer) serveFile(w http.ResponseWriter, r *http.Request, path string) {
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.rootDir, cleanPath)

	if !strings.HasPrefix(fullPath, s.rootDir) {
		http.Error(w, "Accesso non consentito", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, fullPath)
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

		// Mostra solo cartelle e file multimediali (immagini + video)
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		isImage := isImageExt(ext)
		isVideo := isVideoExt(ext)
		isMedia := isImage || isVideo

		// Salta file non multimediali e non cartelle
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

		// Mostra solo cartelle e file multimediali (immagini + video)
		ext := strings.ToLower(filepath.Ext(d.Name()))
		isImage := isImageExt(ext)
		isVideo := isVideoExt(ext)
		isMedia := isImage || isVideo

		// Salta file non multimediali e non cartelle
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
