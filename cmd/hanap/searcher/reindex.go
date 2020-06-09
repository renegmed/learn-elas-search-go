package searcher

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	elastic "github.com/olivere/elastic/v7"
)

type csvLine struct {
	topic  string
	source string
}

type ingestResp struct {
	resp string
	err  error
}

func (s *searcher) Destroy(index string) error {
	ctx := context.Background()

	exists, err := s.client.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if exists {
		// Delete an index.
		deleteIndex, err := s.client.DeleteIndex(index).Do(ctx)
		if err != nil {
			check(err)
		}
		if !deleteIndex.Acknowledged {
			// Not acknowledged
		}
	}
	return nil
}

func (s *searcher) Reindex(file, suffix, index string) (string, error) {
	ctx := context.Background()
	exists, err := s.client.IndexExists(index).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("Error on checking existence of index, %v", err)
	}

	if !exists {
		createIndex, err := s.client.CreateIndex(index).Body(Mapping).Do(ctx)
		if err != nil {
			return "", fmt.Errorf("Error on creating new index, %v", err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}

	exists, err = s.client.IndexExists(index).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("Error 2 on checking existence of index, %v", err)
	}

	var fileList []csvLine
	csvFileList, err := csvFileList(file)
	if err != nil {
		return "", err
	}
	if suffix != "web" {
		fileList, err = walkFileList(csvFileList, suffix)
		if err != nil {
			return "", err
		}
	} else {
		fileList = convertFileList(csvFileList)
	}

	fmt.Println("Point 0....file size..", len(fileList))

	type token struct{}

	sem := make(chan token, len(fileList))
	batches := make(chan csvLine, len(fileList))

	for i, f := range fileList {
		go func(cline csvLine, i int, limit int) {
			batches <- cline
			sem <- token{}
		}(f, i, len(fileList)-1)
	}

	for i := len(fileList); i >= 0; i-- {
		if i > 0 {
			<-sem
		} else {
			close(batches)
		}
	}

	messages := []ingestResp{}

	batches2 := BatchCsvLine(batches, 200, 60000*time.Millisecond)
	for {
		select {
		case files, ok := <-batches2:
			if ok {
				for _, f := range files {
					resp := indexFile(s.client, f)
					messages = append(messages, resp)
				}
			} else {
				goto done
			}
		}
	}
done:
	var buffer bytes.Buffer
	count := 1
	for _, message := range messages {
		buffer.WriteString(strconv.Itoa(count))
		buffer.WriteString(". ")
		if message.err != nil {
			buffer.WriteString(fmt.Sprintf("%v", message.err))
		} else {
			buffer.WriteString(message.resp)
		}

		buffer.WriteString("\n")
		count++
	}

	return buffer.String(), nil

}

func BatchCsvLine(cline <-chan csvLine, maxItems int, maxTimeout time.Duration) chan []csvLine {
	batches := make(chan []csvLine)

	go func() {
		defer close(batches) // stops receiving signal when the channel is closed
		for keepGoing := true; keepGoing; {
			var batch []csvLine
			expire := time.After(maxTimeout) // expire is channel that receives a signal when timeout is reached
			for {
				select {
				case value, ok := <-cline:
					if !ok { // channel was closed
						keepGoing = false // this flag causes to exit out of the loop
						goto done
					}

					batch = append(batch, value)
					if len(batch) == maxItems { // max is reached before timeout, done, send the batch now regardless of content
						goto done
					}

				case <-expire: // timeout reached before reaching maximum items
					keepGoing = false
					goto done // causes to send batches to channel, but continue into the loop
				}
			}

		done:
			if len(batch) > 0 {
				batches <- batch
			}
		}
	}()

	return batches
}

func csvFileList(file string) ([][]string, error) {
	csvFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()
	lines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return nil, err
	}

	return lines, nil
}

func convertFileList(lines [][]string) []csvLine {
	csvLines := make([]csvLine, 0)

	for _, line := range lines[1:] { // skip first line
		data := csvLine{}
		data.topic = line[0]
		data.source = strings.TrimSpace(line[1])
		csvLines = append(csvLines, data)
	}
	return csvLines
}

func walkFileList(lines [][]string, suffix string) ([]csvLine, error) {
	csvLines := make([]csvLine, 0)

	for _, line := range lines[1:] { // skip first line

		searchDir := line[1]
		fileList := []string{}
		err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {

			if strings.HasSuffix(path, "/vendor") {
				return nil
			}

			if strings.Contains(path, "/vendor/") {
				return nil
			}

			if strings.Contains(path, "/Godeps") {
				return nil
			}

			if strings.Contains(path, "/node_modules") {
				return nil
			}

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
			return nil, err
		}
		for _, file := range fileList {
			data := csvLine{}
			data.topic = line[0]
			data.source = file
			csvLines = append(csvLines, data)
		}
	}
	return csvLines, nil
}

func indexFile(client *elastic.Client, csvLine csvLine) ingestResp {

	//fmt.Println("Index File")
	res := ingestResp{}
	msg, err := processIndex(client, csvLine)
	if err != nil {
		res.err = err
		res.resp = string(fmt.Sprintf("ERROR file: %s\n %v", csvLine.source, err))
		return res
	}
	res.resp = msg // string(fmt.Sprintf("file indexed: %s", csvLine.source))

	// fmt.Println(res)
	return res
}

func processIndex(client *elastic.Client, csvLine csvLine) (string, error) {

	//fmt.Println("Process Index....")
	if strings.HasPrefix(csvLine.source, "http") {
		byteContents, err := scrapeHtml(csvLine.source)
		if err != nil {
			//fmt.Printf("%v\n", err)
			return "", fmt.Errorf("Error while scraping html, %v", err)
		}

		content := Content{Topic: csvLine.topic, Content: string(byteContents), Source: csvLine.source}

		msg, err := addToIndex(client, csvLine.topic, content)
		if err != nil {
			return "", fmt.Errorf("Error on add http document to index, %v", err)
		}
		return msg, nil

	} else {
		byteContents, err := ioutil.ReadFile(csvLine.source)
		if err != nil {
			//fmt.Printf("%v\n", err)
			return "", fmt.Errorf("Error on reading file: %v", err)
		}

		content := Content{Topic: csvLine.topic, Content: string(byteContents), Source: csvLine.source}

		msg, err := addToIndex(client, csvLine.topic, content)
		if err != nil {
			return "", fmt.Errorf("Error on add file document to index, %v", err)
		}

		return msg, nil
	}

}

func addToIndex(client *elastic.Client, topic string, content Content) (string, error) {

	//time.Sleep(100 * time.Millisecond)

	//fmt.Println("Add to index:", content.Source)

	dataJSON, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("Error on content marshalling, %v", err)
	}
	js := string(dataJSON)

	_, err = client.Index().
		Index(topic).
		//Type("doc").
		//Id(id).
		BodyJson(js).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return "", fmt.Errorf("Error on es indexing, %v", err)
	}
	return fmt.Sprintf("Added to index: %s", content.Source), nil
}

func scrapeHtml(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s %s", res.StatusCode, res.Status, url)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var textBuilder bytes.Buffer

	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		content := s.Text()
		fmt.Printf("Title %d: %s \n", i, content)
		textBuilder.WriteString(content)
	})

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		content := s.Text()
		textBuilder.WriteString(content)
	})

	return textBuilder.String(), nil
}
