package wfs

import "io"

// FsObject stores info about single file
type FsObject struct {
	Name  string     `json:"value"`
	ID    string     `json:"id"`
	Size  int64      `json:"size"`
	Date  int64      `json:"date"`
	Type  string     `json:"type"`
	Files []FsObject `json:"data,omitempty"`
}

// ListConfig contains file listing options
type ListConfig struct {
	SkipFiles  bool
	SubFolders bool
	Nested     bool
	Exclude    MatcherFunc
	Include    MatcherFunc
}

// DriveConfig contains drive configuration
type DriveConfig struct {
	Verbose   bool
	List      *ListConfig
	Operation *OperationConfig
	Policy    *Policy
}

// OperationConfig contains file operation options
type OperationConfig struct {
	PreventNameCollision bool
}

// MatcherFunc receives path and returns true if path matches the rule
type MatcherFunc func(string) bool

// Drive provides all file operations for some storage
type Drive interface {
	Exists(path string) bool
	Info(path string) (*FsObject, error)
	Read(path string) (io.ReadSeeker, error)
	List(folder string, config ...*ListConfig) ([]FsObject, error)

	Remove(path string) error
	Mkdir(path string) (string, error)
	Write(path string, source io.Reader) (string, error)
	Copy(source string, target string) (string, error)
	Move(source string, target string) (string, error)

	WithOperationConfig(config *OperationConfig) Drive
}

// Operation defines type of operation
type Operation int

// Supported operation modes
const (
	ReadOperation Operation = iota
	WriteOperation
)

// Policy is a rule which allows or denies operation
type Policy interface {
	// Comply method returns true is operation for the path is allowed
	Comply(path string, operation Operation) bool
}
