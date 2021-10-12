package main

import (
	"io"
	"io/fs"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strings"
)

type File struct {
	Name string
	Size int64
}

type FileList struct {
	Device string
	Files  []File
}

func cleanUp() {
	Log("info", "cleaning up...")
	path := "./public/files/"
	files := fileList(path)

	for _, f := range files {
		if f.Name() == ".empty" {
			continue
		}
		Log("info", "removing "+path+f.Name())
		err := os.RemoveAll(path + f.Name())
		if err != nil {
			Fatal(err.Error())
		}
	}
}

func handleFile(file multipart.File, device string, fileName string) ([]FileList, error) {
	os.Mkdir("./public/files/"+device, 0775)
	f, err := os.OpenFile("./public/files/"+device+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return getFileList(), err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	return getFileList(), err
}

func getFileList() []FileList {
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

func readFile(path string, defaultStr string) string {
	s, err := ioutil.ReadFile(path)
	if err != nil {
		return defaultStr
	}
	return strings.TrimSpace(string(s))
}

func fileList(path string) []fs.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		Fatal(err.Error())
	}
	return files
}
