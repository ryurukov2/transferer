package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var messageDelim string = "\r\n"
var logDir string = "logs/"

func setLogDir() error {
	stat, err := os.Stat(logDir)
	// 3. file/dir doesn't exist
	if err != nil && os.IsNotExist(err) {
		err := os.Mkdir(logDir, 0777)
		if err != nil {
			return err
		}
		return nil
	}
	// 2. file exists but is not dir
	if err == nil && !stat.IsDir() {
		logDir := logDir + "(1)"
		setReceivedFilesDir(logDir)
	}
	// 4. permission error
	if stat.IsDir() && os.IsPermission(err) {
		logDir := logDir + "(1)"
		setReceivedFilesDir(logDir)
	}
	return nil
}

// reads files from an open directory and formats them to a string, deliminated by two commas.
// Returns "filename//isDir" for all files.
func readFilesToStr(f *os.File) (string, error) {
	files, err := f.Readdir(0)
	if err != nil {
		fmt.Printf("unable to read files in directory - %v\n", err)
		return "", err
	}
	fileStr := ""
	for _, file := range files {
		fileStr += fmt.Sprintf("%v//%v,,", file.Name(), file.IsDir())
	}
	return fileStr, nil
}

// Parses a string of files deliminated by two commas and returns a slice of fileData objects
func parseFileStr(fileStr string) ([]fileData, error) {
	filesObj := []fileData{}
	files := strings.Split(fileStr, ",,")
	for _, file := range files {
		if file == "\n" {
			break
		}
		fileAttrs := strings.Split(file, "//")
		fileIsFolder, err := strconv.ParseBool(fileAttrs[1])
		if err != nil {
			return filesObj, fmt.Errorf("file parsing error %v", err)
		}
		tmpObj := fileData{Name: fileAttrs[0],
			IsFolder: bool(fileIsFolder)}
		filesObj = append(filesObj, tmpObj)
	}
	return filesObj, nil
}

func writeLog(filename string, logdata string) {
	fPath := filepath.Join("logs/", filename)
	f, err := os.OpenFile(fPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)
	logger.Println(logdata)
}

func selectFolder(a *App, prompt string) (string, error) {
	folderPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: prompt,
	})
	if err != nil {
		return "", err
	}
	return folderPath, nil

}
