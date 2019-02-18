package test

import (
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/xbsoftware/wfs-local"
)

func TestLocalTestSuite(t *testing.T) {
	suite.Run(t, new(LocalTestSuite))
}

type LocalTestSuite struct {
	suite.Suite
	drive          wfs.Drive
	collisionDrive wfs.Drive
}

func (suite *LocalTestSuite) BeforeTest(name, test string) {
	suite.drive, _ = wfs.NewLocalDrive("./sandbox", &wfs.DriveConfig{Verbose: true})
	suite.collisionDrive = suite.drive.WithOperationConfig(&wfs.OperationConfig{PreventNameCollision: true})
}

func (suite *LocalTestSuite) TestLocalFilesExists() {
	suite.True(suite.drive.Exists("/sub"))
	suite.False(suite.drive.Exists("/sub2"))
}

func (suite *LocalTestSuite) TestInitModes() {
	_, err := wfs.NewLocalDrive("./sandbox", nil)
	suite.Nil(err)
	_, err = wfs.NewLocalDrive("./sandbox", &wfs.DriveConfig{})
	suite.Nil(err)
}

func (suite *LocalTestSuite) TestLocalFilesInfo() {
	info1, err := suite.drive.Info("/sub")
	suite.Nil(err)
	suite.Equal("sub", info1.Name)
	suite.Equal("folder", info1.Type)

	info2, err := suite.drive.Info("a.txt")
	suite.Nil(err)
	suite.Equal("a.txt", info2.Name)
	suite.Equal("text", info2.Type)
}

func (suite *LocalTestSuite) TestLocalFilesRemove() {
	path, err := suite.drive.Write("/sub/test.doc", []byte("some"))
	suite.Nil(err)
	suite.Equal("/sub/test.doc", path)
	suite.True(suite.drive.Exists("/sub/test.doc"))

	err = suite.drive.Remove("/sub/test.doc")
	suite.Nil(err)
	suite.False(suite.drive.Exists("/sub/test/doc"))
}

func (suite *LocalTestSuite) writeAndCheck(target, text, check string, clean bool, drive wfs.Drive) {
	path, err := drive.Write(target, []byte(text))
	suite.Nil(err)
	suite.Equal(check, path)
	suite.True(drive.Exists(check))

	data, err := drive.Read(check)
	suite.Nil(err)
	suite.Equal(text, string(data))

	if clean {
		err = drive.Remove(check)
		suite.Nil(err)
	}
}
func (suite *LocalTestSuite) TestLocalFilesWrite() {
	suite.writeAndCheck("/sub/deep/copy.doc", "some", "/sub/deep/copy.doc", false, suite.drive)
	suite.writeAndCheck("/sub/deep/copy.doc", "other", "/sub/deep/copy.doc", true, suite.drive)

	suite.writeAndCheck("/sub/deep/deep.doc", "none", "/sub/deep/deep.doc.new", true, suite.collisionDrive)
}

func (suite *LocalTestSuite) TestLocalFilesRead() {
	info1, err := suite.drive.Read("/sub/deep/deep.doc")
	suite.Nil(err)
	suite.Equal("test", strings.TrimSpace(string(info1)))
}

func (suite *LocalTestSuite) TestLocalFilesMkdir() {
	path, err := suite.drive.Mkdir("/alfa/123/a")
	suite.Nil(err)
	suite.Equal("/alfa/123/a", path)
	suite.True(suite.drive.Exists("/alfa/123/a"))

	err = suite.drive.Remove("/alfa")
	suite.Nil(err)

	// prevent overwrite mode
	path, err = suite.collisionDrive.Mkdir("/sub/deep")
	suite.Nil(err)
	suite.Equal("/sub/deep.new", path)
	suite.True(suite.collisionDrive.Exists("/sub/deep.new"))

	err = suite.collisionDrive.Remove(path)
	suite.Nil(err)
}

func (suite *LocalTestSuite) copyAndTest(source, target, check string, drive wfs.Drive) {
	path, err := drive.Copy(source, target)

	suite.Nil(err)
	suite.Equal(check, path)
	suite.True(drive.Exists(check))

	err = drive.Remove(check)
	suite.Nil(err)
}

func (suite *LocalTestSuite) compareLists(as, bs string) {
	a, err1 := suite.drive.List(as)
	b, err2 := suite.drive.List(bs)

	suite.Nil(err1)
	suite.Nil(err2)
	suite.Equal(len(a), len(b))
	for i := range a {
		suite.Equal(a[i].Name, b[i].Name)
		suite.Equal(a[i].Type, b[i].Type)
		suite.Equal(a[i].Size, b[i].Size)
	}
}
func (suite *LocalTestSuite) copyAndTestFolder(source, target, check string, drive wfs.Drive) {
	path, err := drive.Copy(source, target)

	suite.Nil(err)
	suite.Equal(check, path)
	suite.True(drive.Exists(check))

	suite.compareLists(source, path)

	err = drive.Remove(check)
	suite.Nil(err)
}

func (suite *LocalTestSuite) TestLocalFileCopy() {
	//copy files
	suite.copyAndTest("/sub/deep/deep.doc", "/sub/deep/copy.doc", "/sub/deep/copy.doc", suite.drive)
	suite.copyAndTest("/sub/deep/deep.doc", "/sub", "/sub/deep.doc", suite.drive)
	suite.copyAndTest("/sub/deep/deep.doc", "/sub/", "/sub/deep.doc", suite.drive)

	//copy file and rename
	suite.copyAndTest("c.jpg", "/sub", "/sub/c.jpg.new", suite.collisionDrive)

	//copy folder
	suite.copyAndTestFolder("/sub", "/sub2", "/sub2", suite.drive)
	suite.copyAndTestFolder("/sub", "/test.folder", "/test.folder/sub", suite.drive)

	//can't copy folder into itself
	_, err := suite.drive.Copy("/sub", "/sub/")
	suite.Error(err)

	//copy and rename folder
	suite.copyAndTestFolder("/test.folder", "/sub", "/sub/test.folder.new", suite.collisionDrive)
}

func (suite *LocalTestSuite) TestLocalFileList() {
	//can read root
	data, err := suite.drive.List("/")
	suite.Nil(err)
	log.Printf("%+v", data)
	suite.Equal(5, len(data))
	first := data[0]
	suite.Equal("sub", first.Name)
	suite.Equal("/sub", first.ID)
	suite.Equal("folder", first.Type)
	suite.Nil(first.Files)

	suite.Equal("test.folder", data[1].Name)
	suite.Equal("a.txt", data[2].Name)
	suite.Equal("b.txt", data[3].Name)
	suite.Equal("c.jpg", data[4].Name)
	suite.Equal("image", data[4].Type)

	//can read sublevel
	data, err = suite.drive.List("/sub")
	suite.Nil(err)
	suite.Equal(3, len(data))
	suite.Equal("deep", data[0].Name)
	suite.Nil(data[0].Files)
	suite.Equal("c.jpg", data[2].Name)

	//can read folders only
	data, err = suite.drive.List("/", &wfs.ListConfig{SkipFiles: true})
	suite.Nil(err)
	suite.Equal(2, len(data))
	suite.Equal("sub", data[0].Name)
	suite.Equal("test.folder", data[1].Name)

	//can read nested folders
	data, err = suite.drive.List("/", &wfs.ListConfig{SkipFiles: true, SubFolders: true, Nested: true})
	suite.Nil(err)
	suite.Equal(2, len(data))
	suite.Equal("sub", data[0].Name)
	suite.Equal("test.folder", data[1].Name)
	suite.Equal(2, len(data[0].Files))
	suite.Equal("deep", data[0].Files[0].Name)
	suite.Equal("test.folder", data[0].Files[1].Name)
	suite.Nil(data[0].Files[0].Files)

	//can read nested files and folders
	data, err = suite.drive.List("/", &wfs.ListConfig{SubFolders: true, Nested: true})
	suite.Nil(err)
	suite.Equal(5, len(data))
	suite.Equal(3, len(data[0].Files))
	suite.Equal(1, len(data[0].Files[0].Files))

	//can include by mask
	data, err = suite.drive.List("/", &wfs.ListConfig{
		SubFolders: true,
		Include: func(file string) bool {
			r, _ := regexp.MatchString("\\.(txt|doc)$", file)
			return r
		},
	})
	suite.Nil(err)
	suite.Equal(3, len(data))
	suite.Equal("a.txt", data[0].Name)
	suite.Equal("b.txt", data[1].Name)
	suite.Equal("deep.doc", data[2].Name)

	//can exclude by mask
	data, err = suite.drive.List("/", &wfs.ListConfig{
		Exclude: func(file string) bool { return file == "a.txt" },
	})
	suite.Nil(err)
	suite.Equal(4, len(data))
	suite.Equal("sub", data[0].Name)
	suite.Equal("test.folder", data[1].Name)
	suite.Equal("b.txt", data[2].Name)
	suite.Equal("c.jpg", data[3].Name)
}

func (suite *LocalTestSuite) TestLocalFileSecurity() {
	//access outside of root
	data, err := suite.drive.List("../")
	suite.Error(err)
	suite.Nil(data)
}

func (suite *LocalTestSuite) TestFileNames() {
	_, er := suite.drive.Write("/.test", []byte("1"))
	suite.Nil(er)
	_, er = suite.drive.Write("/test", []byte("2"))
	suite.Nil(er)
	_, err := suite.drive.List("/", nil)
	suite.Nil(err)
	er = suite.drive.Remove("/.test")
	suite.Nil(er)
	er = suite.drive.Remove("/test")
	suite.Nil(er)
}

func (suite *LocalTestSuite) TestLocalFileMove() {

	// move file
	data, err := suite.drive.Move("/sub/deep/deep.doc", "/sub/deep/copy.doc")
	suite.Nil(err)
	suite.Equal("/sub/deep/copy.doc", data)
	suite.True(suite.drive.Exists("/sub/deep/copy.doc"))
	suite.False(suite.drive.Exists("/sub/deep/deeo.doc"))

	suite.drive.Move("/sub/deep/copy.doc", "/sub/deep/deep.doc")

	// move and rename file
	data, err = suite.collisionDrive.Move("/c.jpg", "/sub/")
	suite.Nil(err)
	suite.Equal("/sub/c.jpg.new", data)
	suite.True(suite.collisionDrive.Exists("/sub/c.jpg.new"))
	suite.False(suite.collisionDrive.Exists("/c.jpg"))

	suite.collisionDrive.Move("/sub/c.jpg.new", "/c.jpg")

	// move folder
	data, err = suite.drive.Copy("/sub", "/sub3")
	suite.Nil(err)
	suite.Equal("/sub3", data)
	suite.compareLists("/sub", "/sub3")
	data, err = suite.drive.Move("/sub3", "/sub2")
	suite.Nil(err)
	suite.Equal("/sub2", data)
	suite.compareLists("/sub", "/sub2")
	data, err = suite.drive.Move("/sub2", "/sub/deep")
	suite.Nil(err)
	suite.Equal("/sub/deep/sub2", data)
	suite.compareLists("/sub", "/sub/deep/sub2")

	err = suite.drive.Remove("/sub/deep/sub2")
	suite.Nil(err)

	// move and rename folder
	data, err = suite.collisionDrive.Copy("/test.folder", "/sub/deep")
	suite.Nil(err)
	suite.Equal("/sub/deep/test.folder", data)
	suite.compareLists("/sub/deep/test.folder", "/sub/test.folder")
	data, err = suite.collisionDrive.Move("/sub/deep/test.folder", "/")
	suite.Nil(err)
	suite.Equal("/test.folder.new", data)
	suite.compareLists("/test.folder.new", "/test.folder")

	err = suite.collisionDrive.Remove("/test.folder.new")
	suite.Nil(err)
}
