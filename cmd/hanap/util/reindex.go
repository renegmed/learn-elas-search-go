package util

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/olivere/elastic"
)

func Reindex(f, suffix string) {

	csvFile, err := os.Open(f)
	check(err)
	defer csvFile.Close()
	lines, err := csv.NewReader(csvFile).ReadAll()
	check(err)

	csvLines := make([]CsvLine, 0)
	for _, line := range lines[1:] { // skip first line

		searchDir := line[1]

		fileList := []string{}

		err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
			if suffix == "" {
				fileList = append(fileList, path)
			} else {
				if strings.HasSuffix(path, suffix) {
					fileList = append(fileList, path)
				}
			}

			return nil
		})

		check(err)

		for _, file := range fileList {
			data := new(CsvLine)
			data.Topic = line[0]
			data.Source = file
			csvLines = append(csvLines, *data)
		}
	}

	ctx := context.Background()
	// Create a elasticsearch client
	client, err := elastic.NewClient()
	check(err)

	exists, err := client.IndexExists("golang").Do(ctx)
	if err != nil {
		check(err)
	}
	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex("golang").BodyString(Mapping).Do(ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}
	//counter := 1
	// read each file and index their contents(lines)
	for _, f := range csvLines {
		byteContents, err := ioutil.ReadFile(f.Source)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		content2 := Content{Topic: "golang", Content: string(byteContents), Source: f.Source}
		//addToIndex(client, strconv.Itoa(counter), f.topic, content2)
		addToIndex(client, f.Topic, content2)
		//counter++

		fmt.Printf("Indexed: %s\n", f.Source)
	}
}

func addToIndex(client *elastic.Client, topic string, content Content) error {
	_, err := client.Index().
		Index(topic).
		Type("doc").
		//Id(id).
		BodyJson(content).
		Refresh("wait_for").
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}
