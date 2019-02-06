package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	searchDir := "/Users/rene/learn/go-workspace/src/elasticsearch-olivere"

	fileList := []string{}

	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".go") {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, file := range fileList {
		fmt.Println(file)
	}
}
