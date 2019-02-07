package util

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/olivere/elastic"
)

type barrierResp struct {
	Resp string
	Err  error
}

func Reindex(f, suffix string) (string, error) {

	result, err := captureBarrierOutput(f, suffix)
	if err != nil {
		return "", err
	}
	return result, nil

}

// capture the output from std output
func captureBarrierOutput(f, suffix string) (string, error) {

	csvFile, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer csvFile.Close()
	lines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return "", err
	}

	csvLines := make([]CsvLine, 0)
	for _, line := range lines[1:] { // skip first line

		searchDir := line[1]

		fileList := []string{}

		err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {

			if strings.HasSuffix(searchDir, suffix) {
				fileList = append(fileList, path)
			} else {
				if strings.HasSuffix(path, suffix) {
					fileList = append(fileList, path)
				}
			}

			return nil
		})

		if err != nil {
			return "", err
		}

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
	if err != nil {
		return "", err
	}

	exists, err := client.IndexExists("golang").Do(ctx)
	if err != nil {
		return "", err
	}

	fmt.Printf("--------- client.IndexExists(\"golang\"): %v\n", exists)

	if !exists {
		// Create a new index.
		//createIndex, err := client.CreateIndex("doc").BodyString(util.Mapping).Do(ctx)
		createIndex, err := client.CreateIndex("golang").Body(Mapping).Do(ctx)
		if err != nil {
			return "", err
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}

		fmt.Printf("--------- created index 'golang'\n")
	}

	fmt.Println("Point 1")

	reader, writer, err := os.Pipe()
	if err != nil {
		return "", err
	}

	fmt.Println("Point 2")

	os.Stdout = writer // make the writer as output handler

	fmt.Println("Point 3")

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, reader)
		outChan <- buf.String() // send read string to channel
	}()

	fmt.Println("Point 4")

	barrier(client, csvLines...)

	fmt.Println("Point 5")

	writer.Close()
	temp := <-outChan // waits all the results here - messages from outChan are all in.

	return temp, nil
}

func barrier(client *elastic.Client, files ...CsvLine) {
	fmt.Printf("------ barrier:  files size - %d\n", len(files))

	requestNumber := len(files)

	in := make(chan barrierResp, requestNumber)
	defer close(in)

	responses := make([]barrierResp, requestNumber) // each endpoint has its own response

	for _, file := range files {
		go indexFile(client, in, file) // call each enpoint and put into channel the response
	}

	var hasError bool
	for i := 0; i < requestNumber; i++ {
		resp := <-in // resp is a barrierResp
		if resp.Err != nil {
			fmt.Println("ERROR: ", resp.Err)
			hasError = true
		}
		responses[i] = resp
	}

	if !hasError {
		for _, resp := range responses {
			fmt.Println(resp.Resp)
		}
	}
}

// Make http request and process the response/error
func indexFile(client *elastic.Client, out chan<- barrierResp, csvLine CsvLine) { // sending channel

	fmt.Printf("------ file to index: %s\n", csvLine.Source)

	res := barrierResp{}

	err := processIndex(client, csvLine)
	if err != nil {
		res.Err = err
		out <- res
		return
	}

	res.Resp = string(fmt.Sprintf("file indexed: %s", csvLine.Source))
	out <- res
}

func processIndex(client *elastic.Client, csvLine CsvLine) error {

	byteContents, err := ioutil.ReadFile(csvLine.Source)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	content := Content{Topic: "golang", Content: string(byteContents), Source: csvLine.Source}

	err = addToIndex(client, csvLine.Topic, content)
	if err != nil {
		return err
	}

	fmt.Printf("++++++++ Indexed: %s\n", csvLine.Source)

	return nil
}

func addToIndex(client *elastic.Client, topic string, content Content) error {

	fmt.Printf("======== addToIndex: \n")

	_, err := client.Index().
		Index(topic).
		Type("doc").
		//Id(id).
		BodyJson(content).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}
