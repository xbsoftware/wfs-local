package wfs

import (
	"os"
	"path/filepath"
	"strings"
)

var types map[string]string

func init() {
	types = map[string]string{
		"docx": "document",
		"doc":  "document",
		"odt":  "document",
		"xls":  "document",
		"xslx": "document",
		"pdf":  "document",
		"djvu": "document",
		"djv":  "document",
		"pptx": "document",
		"ppt":  "document",

		"html":   "code",
		"htm":    "code",
		"js":     "code",
		"json":   "code",
		"css":    "code",
		"scss":   "code",
		"sass":   "code",
		"php":    "code",
		"sh":     "code",
		"coffee": "code",
		"txt":    "code",
		"md":     "code",

		"mpg": "video",
		"mp4": "video",
		"avi": "video",
		"mkv": "video",
		"ogv": "video",

		"png":  "image",
		"jpg":  "image",
		"jpeg": "image",
		"gif":  "image",
		"tiff": "image",
		"tif":  "image",
		"svg":  "image",

		"mp3":  "audio",
		"ogg":  "audio",
		"flac": "audio",
		"wav":  "audio",

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

	ftype, ok := types[strings.ToLower(ext[1:])]
	if !ok {
		return "file"
	}

	return ftype
}
