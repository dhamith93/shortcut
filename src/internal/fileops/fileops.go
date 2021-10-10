package fileops

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"github.com/dhamith93/shortcut/internal/logger"
)

type File struct {
	Name string
	Size int64
}

type FileList struct {
	Device string
	Files  []File
}

// CleanUp removes all files in private/files dir
func CleanUp() {
	fmt.Println("Cleaning up...")
	path := "./public/files/"
	files := fileList(path)

	for _, f := range files {
		if f.Name() == ".empty" {
			continue
		}
		logger.Log("info", "removing "+path+f.Name())
		err := os.RemoveAll(path + f.Name())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func HandleFile(file multipart.File, device string, fileName string) ([]FileList, error) {
	os.Mkdir("./public/files/"+device, 0775)
	f, err := os.OpenFile("./public/files/"+device+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return GetFileList(), err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	return GetFileList(), err
}

// GetFileList returns an array of FileList structs with files in private/files dir
func GetFileList() []FileList {
	path := "./public/files/"
	files := fileList(path)

	out := []FileList{}

	for _, f := range files {
		if f.IsDir() {
			dirName := f.Name()
			fileArr := []File{}
			filesInDir := fileList(path + "/" + dirName)
			for _, file := range filesInDir {
				if !file.IsDir() {
					fileArr = append(fileArr, File{
						Name: file.Name(), Size: file.Size(),
					})
				}
			}
			out = append(out, FileList{
				Device: dirName, Files: fileArr,
			})
		}
	}

	return out
}

// ReadFile read from given file
func ReadFile(path string, defaultStr string) string {
	s, err := ioutil.ReadFile(path)
	if err != nil {
		return defaultStr
	}
	return strings.TrimSpace(string(s))
}

func fileList(path string) []fs.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	return files
}
