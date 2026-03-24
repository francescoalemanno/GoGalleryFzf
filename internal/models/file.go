package models

import "time"

type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	IsDir   bool      `json:"isDir"`
	IsImage bool      `json:"isImage"`
	IsVideo bool      `json:"isVideo"`
	IsAudio bool      `json:"isAudio"`
	Ext     string    `json:"ext"`
}
