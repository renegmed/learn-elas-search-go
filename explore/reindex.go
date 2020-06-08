//
// Barrier Concurrency Pattern
// Purpose: put up a barrier so that nobody passes until we have all the results we need
//
package explore

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/renegmed/learn-elas-search-go/cmd/hanap/util"

	elastic "gopkg.in/olivere/elastic.v7"
)

const mapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"doc":{
			"properties":{
				"topic":{
					"type":"keyword"
				},
				"content":{
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"source":{
					"type":"text"
				} 
			}
		}
	}
}`

var timeoutMilliseconds int = 5000

type barrierResp struct {
	Resp string
	Err  error
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

	csvLines := make([]util.CsvLine, 0)
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
			data := new(util.CsvLine)
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
		createIndex, err := client.CreateIndex("golang").Body(mapping).Do(ctx)

		if err != nil {
			return "", err
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}

		fmt.Printf("--------- created index 'golang'\n")
	}

	reader, writer, _ := os.Pipe()

	os.Stdout = writer // make the writer as output handler

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, reader)
		outChan <- buf.String() // send read string to channel
	}()

	barrier(client, csvLines...)

	writer.Close()
	temp := <-outChan // waits all the results here - messages from outChan are all in.

	return temp, nil
}

func barrier(client *elastic.Client, files ...util.CsvLine) {
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
func indexFile(client *elastic.Client, out chan<- barrierResp, csvLine util.CsvLine) { // sending channel
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

func processIndex(client *elastic.Client, csvLine util.CsvLine) error {

	byteContents, err := ioutil.ReadFile(csvLine.Source)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	content := util.Content{Topic: "golang", Content: string(byteContents), Source: csvLine.Source}

	err = addToIndex(client, csvLine.Topic, content)
	if err != nil {
		return err
	}

	fmt.Printf("++++++++ Indexed: %s\n", csvLine.Source)

	return nil
}

func addToIndex(client *elastic.Client, topic string, content util.Content) error {

	fmt.Println("++++++++ 1.1 +++++++++")
	dataJSON, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("Error on content marshalling, %v", err)
	}
	js := string(dataJSON)

	fmt.Println("++++++++ 1.2 +++++++++")

	_, err := client.Index().
		Index(topic).
		Type("doc").
		//Id(id).
		BodyJson(js).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}
