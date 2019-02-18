package wfs

import (
	"os"
	"path/filepath"
)

var types map[string]string

func init() {
	types = map[string]string{
		"docx": "doc",

		"xls":  "excel",
		"xslx": "excel",

		"txt": "text",
		"md":  "text",

		"html": "code",
		"htm":  "code",
		"js":   "code",
		"json": "code",
		"css":  "code",
		"php":  "code",
		"sh":   "code",

		"mpg": "video",
		"mp4": "video",
		"avi": "video",
		"mkv": "video",

		"png": "image",
		"jpg": "image",
		"gif": "image",

		"mp3": "audio",
		"ogg": "audio",

		"zip": "archive",
		"rar": "archive",
		"7z":  "archive",
		"tar": "archive",
		"gz":  "archive",
	}
}

func getType(info os.FileInfo) string {
	if info.IsDir() {
		return "folder"
	}

	ext := filepath.Ext(info.Name())
	if ext == "" {
		return "file"
	}

	ftype, ok := types[ext[1:]]
	if !ok {
		return "file"
	}

	return ftype
}
