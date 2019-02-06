package util

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/olivere/elastic"
)

func Reindex(f string) {
	csvFile, err := os.Open(f)
	check(err)
	defer csvFile.Close()
	lines, err := csv.NewReader(csvFile).ReadAll()
	check(err)

	csvFiles := make([]CsvLine, 0)
	for _, line := range lines[1:] { // skip first line
		data := new(CsvLine)
		data.topic = line[0]
		data.source = line[1]
		csvFiles = append(csvFiles, *data)
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
		createIndex, err := client.CreateIndex("golang").BodyString(mapping).Do(ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}
	counter := 1
	// read each file and index their contents(lines)
	for _, f := range csvFiles {
		byteContents, err := ioutil.ReadFile(f.source)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		content2 := Content{Topic: "golang", Content: string(byteContents), Source: f.source}
		addToIndex(client, strconv.Itoa(counter), f.topic, content2)
		counter++
	}
}

func addToIndex(client *elastic.Client, id string, topic string, content Content) error {
	_, err := client.Index().
		Index(topic).
		Type("doc").
		Id(id).
		BodyJson(content).
		Refresh("wait_for").
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}
