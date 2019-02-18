package wfs

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// NewLocalDrive returns new LocalDrive object
// which represents the file folder on local drive
// due to ForceRootPolicy all operations outside of the root folder will be blocked
func NewLocalDrive(root string, config *DriveConfig) (*LocalDrive, error) {
	absroot, err := filepath.Abs(root)
	if err != nil {
		return nil, errors.New("Invalid path: " + root)
	}

	drive := LocalDrive{
		root:   absroot,
		policy: &ForceRootPolicy{absroot},
	}

	if config != nil {
		drive.verbose = config.Verbose
		drive.list = config.List
		drive.operation = config.Operation

		if config.Policy != nil {
			drive.policy = CombinedPolicy{
				[]Policy{
					*config.Policy,
					drive.policy,
				},
			}
		}
	}

	if drive.list == nil {
		drive.list = &ListConfig{}
	}
	if drive.operation == nil {
		drive.operation = &OperationConfig{}
	}

	return &drive, nil
}

// LocalDrive represents a folder on local drive
type LocalDrive struct {
	root      string
	list      *ListConfig
	operation *OperationConfig
	policy    Policy
	verbose   bool
}

// List method returns array of files from the target folder
func (drive LocalDrive) List(path string, config ...*ListConfig) ([]FsObject, error) {
	fullpath := drive.idToPath(path)

	if drive.verbose {
		log.Printf("List %s", path)
	}

	if !drive.policy.Comply(fullpath, ReadOperation) {
		return nil, errors.New("Access Denied")
	}

	list := drive.list
	if len(config) > 0 {
		list = config[0]
	}
	if list == nil {
		list = drive.list
	}

	if drive.verbose {
		log.Printf("with config %+v", config)
	}

	return drive.listFolder(fullpath, path, list, nil)
}

// Remove deletes a file or a folder
func (drive LocalDrive) Remove(path string) error {
	path = drive.idToPath(path)

	if drive.verbose {
		log.Printf("Remove %s", path)
	}

	if !drive.policy.Comply(path, WriteOperation) {
		return errors.New("Access Denied")
	}

	return os.RemoveAll(path)
}

// Read returns content of a file
func (drive LocalDrive) Read(path string) ([]byte, error) {
	if drive.verbose {
		log.Printf("Read %s", path)
	}

	path = drive.idToPath(path)
	if !drive.policy.Comply(path, ReadOperation) {
		return nil, errors.New("Access Denied")
	}

	return ioutil.ReadFile(path)
}

// Write saves content to a file
func (drive LocalDrive) Write(path string, data []byte) (string, error) {
	if drive.verbose {
		log.Printf("Write %s", path)
	}

	path = drive.idToPath(path)
	if !drive.policy.Comply(path, WriteOperation) {
		return "", errors.New("Access Denied")
	}

	if drive.operation.PreventNameCollision {
		var err error
		path, err = drive.checkName(path)
		if err != nil {
			return "", err
		}
	}

	return drive.pathToID(path), ioutil.WriteFile(path, data, 0600)
}

func (drive LocalDrive) Exists(id string) bool {
	_, err := drive.Info(id)
	if err != nil {
		return false
	}

	return true
}

// Info returns info about a single file
func (drive LocalDrive) Info(id string) (*FsObject, error) {
	path := drive.idToPath(id)
	if !drive.policy.Comply(path, ReadOperation) {
		return nil, errors.New("Access Denied")
	}

	file, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	fs := FsObject{file.Name(), id, file.Size(), file.ModTime().Unix(), getType(file), nil}
	return &fs, nil
}

//Mkdir creates a new folder
func (drive LocalDrive) Mkdir(path string) (string, error) {
	if drive.verbose {
		log.Printf("Make folder %s", path)
	}

	path = drive.idToPath(path)
	if !drive.policy.Comply(path, WriteOperation) {
		return "", errors.New("Access Denied")
	}

	if drive.operation.PreventNameCollision {
		var err error
		path, err = drive.checkName(path)
		if err != nil {
			return "", err
		}
	}

	return drive.pathToID(path), os.MkdirAll(path, os.FileMode(int(0700)))
}

// Copy makes a copy of file or a folder
func (drive LocalDrive) Copy(source string, target string) (string, error) {
	if drive.verbose {
		log.Printf("Copy %s to %s", source, target)
	}

	source = drive.idToPath(source)
	target = drive.idToPath(target)

	if !drive.policy.Comply(source, ReadOperation) || !drive.policy.Comply(target, WriteOperation) {
		return "", errors.New("Access Denied")
	}

	st, _ := drive.isFolder(source)
	et, ok := drive.isFolder(target)

	//file to folder
	if et {
		target = filepath.Join(target, filepath.Base(source))
	} else if st && ok {
		return "", errors.New("Can't copy folder to file")
	}

	if strings.HasPrefix(target, source+string(filepath.Separator)) {
		return "", errors.New("Can't copy folder into self")
	}

	if drive.operation.PreventNameCollision {
		var err error
		target, err = drive.checkName(target)
		if err != nil {
			return "", err
		}
	}

	//folder to folder
	if st {
		return drive.pathToID(target), CopyDir(source, target)
	}

	//file to file
	return drive.pathToID(target), CopyFile(source, target)

}

// Move renames(moves) a file or a folder
func (drive LocalDrive) Move(source string, target string) (string, error) {
	if drive.verbose {
		log.Printf("Move %s to %s", source, target)
	}

	source = drive.idToPath(source)
	target = drive.idToPath(target)

	if !drive.policy.Comply(source, WriteOperation) || !drive.policy.Comply(target, WriteOperation) {
		return "", errors.New("Access Denied")
	}

	st, _ := drive.isFolder(source)
	et, ok := drive.isFolder(target)

	//file to folder
	if et {
		target = filepath.Join(target, filepath.Base(source))
	} else if st && ok {
		return "", errors.New("Can't copy folder to file")
	}

	if strings.HasPrefix(target, source+string(filepath.Separator)) {
		return "", errors.New("Can't copy folder into self")
	}

	if drive.operation.PreventNameCollision {
		var err error
		target, err = drive.checkName(target)
		if err != nil {
			return "", err
		}
	}

	return drive.pathToID(target), os.Rename(source, target)
}

func (drive LocalDrive) isFolder(path string) (bool, bool) {
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		return true, true
	}

	return false, err == nil
}

func (drive LocalDrive) idToPath(id string) string {
	return filepath.Clean(filepath.Join(drive.root, id))
}

func (drive LocalDrive) pathToID(path string) string {
	return strings.Replace(path, drive.root, "", 1)
}

func (drive LocalDrive) listFolder(path string, prefix string, config *ListConfig, res []FsObject) ([]FsObject, error) {
	list, er := ioutil.ReadDir(path)
	if er != nil {
		return nil, er
	}

	needSortData := false
	if config.Nested || res == nil {
		res = make([]FsObject, 0, len(list))
		needSortData = true
	}

	for _, file := range list {
		skipFile := false
		if config.Exclude != nil && config.Exclude(file.Name()) {
			skipFile = true
		}
		if config.Include != nil && !config.Include(file.Name()) {
			skipFile = true
		}

		isDir := file.IsDir()
		if !isDir && (config.SkipFiles || skipFile) {
			continue
		}

		fullpath := filepath.Join(drive.root, prefix, file.Name())
		id := drive.pathToID(fullpath)
		fs := FsObject{file.Name(), id, file.Size(), file.ModTime().Unix(), getType(file), nil}

		if isDir && config.SubFolders {
			sub, err := drive.listFolder(
				filepath.Join(path, file.Name()),
				filepath.Join(prefix, file.Name()),
				config, res)

			fs.Type = "folder"
			if err != nil {
				return nil, err
			}

			if !config.Nested {
				res = sub
			} else if len(sub) > 0 {
				fs.Files = sub
			}
		}

		if !skipFile {
			res = append(res, fs)
		}
	}

	// sort files and folders by name, folders first
	if needSortData {
		sort.Slice(res, func(i, j int) bool {
			afolder := res[i].Type == "folder"
			bfolder := res[j].Type == "folder"
			if (afolder || bfolder) && res[i].Type != res[j].Type {
				return afolder
			}

			return strings.ToUpper(res[i].Name) < strings.ToUpper(res[j].Name)
		})
	}

	return res, nil
}

func (drive LocalDrive) checkName(path string) (string, error) {
	_, err := os.Stat(path)

	for !os.IsNotExist(err) {
		path = path + ".new"
		_, err = os.Stat(path)
	}

	return path, nil
}

// WithOperationConfig makes a copy of drive with new operation config
func (drive *LocalDrive) WithOperationConfig(config *OperationConfig) Drive {
	copy := *drive
	copy.operation = config

	return &copy
}
