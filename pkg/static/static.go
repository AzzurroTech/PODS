// Package static implements static data serving functionality for the song project
package static

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StaticFile struct {
	Name        string                 `json:"name"`
	Path        string                 `json:"path"`
	Modified    time.Time              `json:"modified"`
	Size        int64                  `json:"size"`
	ContentType string                 `json:"content_type"`
	IsDir       bool                   `json:"is_dir"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type StaticDataManager struct {
	basePath     string
	allowedPaths map[string]bool
	indexHTML    *StaticFile
}

func NewStaticDataManager(basePath string, allowedPaths ...string) (*StaticDataManager, error) {
	if basePath == "" {
		basePath = "./static"
	}

	manager := &StaticDataManager{
		basePath:     basePath,
		allowedPaths: make(map[string]bool),
		indexHTML:    nil,
	}

	for _, p := range allowedPaths {
		manager.allowedPaths[p] = true
	}

	if manager.allowedPaths["/"] {
		manager.allowedPaths[""] = true
	}

	if err := manager.loadIndexHTML(); err != nil {
		return nil, fmt.Errorf("failed to load index.html: %w", err)
	}

	return manager, nil
}

func (sdm *StaticDataManager) loadIndexHTML() error {
	indexPath := filepath.Join(sdm.basePath, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		info, err := os.Stat(indexPath)
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		contentType := mime.TypeByExtension(filepath.Ext(indexPath))
		if contentType == "" {
			contentType = "text/html"
		}

		sdm.indexHTML = &StaticFile{
			Name:        "index.html",
			Path:        "/index.html",
			Modified:    info.ModTime(),
			Size:        info.Size(),
			ContentType: contentType,
			IsDir:       false,
			Metadata:    make(map[string]interface{}),
		}
	}

	return nil
}

func (sdm *StaticDataManager) isPathAllowed(reqPath string) bool {
	if len(sdm.allowedPaths) == 0 {
		return true
	}

	for allowedPath := range sdm.allowedPaths {
		if allowedPath == "" && (reqPath == "/" || reqPath == "") {
			return true
		}
		if strings.HasPrefix(reqPath, allowedPath) {
			return true
		}
	}

	return false
}

func (sdm *StaticDataManager) ListFiles(queryPath string) ([]*StaticFile, error) {
	if !sdm.isPathAllowed(queryPath) {
		return nil, fmt.Errorf("path not allowed: %s", queryPath)
	}

	fullPath := filepath.Join(sdm.basePath, strings.TrimPrefix(queryPath, "/"))

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []*StaticFile
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := queryPath
		if queryPath == "/" || queryPath == "" {
			filePath = "/" + entry.Name()
		} else {
			filePath = queryPath + "/" + entry.Name()
		}

		staticFile := &StaticFile{
			Name:        entry.Name(),
			Path:        filePath,
			Modified:    info.ModTime(),
			Size:        info.Size(),
			ContentType: mime.TypeByExtension(filepath.Ext(entry.Name())),
			IsDir:       entry.IsDir(),
			Metadata:    make(map[string]interface{}),
		}

		if staticFile.ContentType == "" {
			if entry.IsDir() {
				staticFile.ContentType = "application/directory"
			} else {
				staticFile.ContentType = "application/octet-stream"
			}
		}

		files = append(files, staticFile)
	}

	return files, nil
}

func (sdm *StaticDataManager) GetFile(path string) (*StaticFile, error) {
	if !sdm.isPathAllowed(path) {
		return nil, fmt.Errorf("path not allowed: %s", path)
	}

	if path == "/" {
		if sdm.indexHTML != nil {
			return sdm.indexHTML, nil
		}
		return nil, fmt.Errorf("file not found: %s", path)
	}

	fullPath := filepath.Join(sdm.basePath, strings.TrimPrefix(path, "/"))

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s with error: %s", path, err)
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(fullPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	metadata := make(map[string]interface{})

	return &StaticFile{
		Name:        filepath.Base(fullPath),
		Path:        path,
		Modified:    info.ModTime(),
		Size:        info.Size(),
		ContentType: contentType,
		IsDir:       info.IsDir(),
		Metadata:    metadata,
	}, nil
}

func (sdm *StaticDataManager) ServeFile(w http.ResponseWriter, r *http.Request, filePath string) error {
	if !sdm.isPathAllowed(filePath) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return fmt.Errorf("path not allowed: %s", filePath)
	}

	fullPath := filepath.Join(sdm.basePath, strings.TrimPrefix(filePath, "/"))

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return fmt.Errorf("file not found: %s", filePath)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return fmt.Errorf("failed to get file info: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(fullPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	if strings.HasSuffix(fullPath, ".gz") {
		w.Header().Set("Content-Encoding", "gzip")
	}

	if strings.HasSuffix(fullPath, ".min.css") || strings.Contains(contentType, "javascript") || strings.Contains(contentType, "js") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else if strings.Contains(contentType, "html") {
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	_, err = io.Copy(w, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

func (sdm *StaticDataManager) GenerateStaticIndexHTML() ([]byte, error) {
	return []byte("<h1>AzzurroTech Song Static Content</h1>\n<p>Magic Link Authentication System - Static File Server</p>"), nil
}
