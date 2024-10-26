package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var messageDelim string = "\r\n"

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
