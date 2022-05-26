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
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/renegmed/learn-elas-search-go/pkg/utils"

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

type workDispatcher struct {
	numOfWorkers int
	taskChan     chan string
	stopChan     chan struct{}
}

func (d *workDispatcher) Dispatch(filename string) {
	d.taskChan <- filename
}

func NewWorkDispatcher(
	num_of_workers int,
	taskCh chan string,
	stopCh chan struct{}) *workDispatcher {

	return &workDispatcher{
		numOfWorkers: num_of_workers,
		taskChan:     taskCh,
		stopChan:     stopCh,
	}
}

func (s *searcher) Destroy(index string) error {
	ctx := context.Background()

	exists, err := s.client.IndexExists(index).Do(ctx)
	if err != nil {
		return utils.Error(err)
	}

	if exists {
		// Delete an index.
		deleteIndex, err := s.client.DeleteIndex(index).Do(ctx)
		if err != nil {
			check(utils.Error(err))
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
		return "", utils.Error(fmt.Errorf("Error on checking existence of index, %v", err))
	}

	if !exists {
		createIndex, err := s.client.CreateIndex(index).Body(Mapping).Do(ctx)
		if err != nil {
			return "", utils.Error(fmt.Errorf("Error on creating new index, %v", err))
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}

	exists, err = s.client.IndexExists(index).Do(ctx)
	if err != nil {
		return "", utils.Error(fmt.Errorf("Error 2 on checking existence of index, %v", err))
	}

	var fileList []csvLine
	csvFileList, err := csvFileList(file)
	if err != nil {
		return "", utils.Error(err)
	}
	if suffix != "web" {
		num_of_workers := runtime.NumCPU() - 2
		taskCh := make(chan string)
		stopCh := make(chan struct{})

		wd := NewWorkDispatcher(num_of_workers, taskCh, stopCh)

		fileList, err = walkFileList(s.client, csvFileList, suffix, wd)
		if err != nil {
			return "", utils.Error(err)
		}
	} else {
		fileList = convertFileList(csvFileList)
	}

	fmt.Println("Point 0....file size..", len(fileList))

	return "", nil
}

func csvFileList(file string) ([][]string, error) {
	csvFile, err := os.Open(file)
	if err != nil {
		return nil, utils.Error(err)
	}
	defer csvFile.Close()
	lines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return nil, utils.Error(err)
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

func walkFileList(
	es *elastic.Client,
	lines [][]string,
	suffix string,
	wd *workDispatcher) ([]csvLine, error) {

	// setup the infrastructure

	for i := 1; i <= wd.numOfWorkers; i++ {
		go func(i int) {
			for {
				select {
				case filename, ok := <-wd.taskChan:
					if !ok {
						return
					}

					data := csvLine{}
					data.topic = lines[1][0] // 2nd row, 1st column value
					data.source = filename   //--- file name listed in cvs
					fmt.Printf("Worker: %d, Topic: %s, Source: %s\n", i, data.topic, data.source)
					ingestResp := indexFile(es, data)
					fmt.Println("INGEST RESP:", ingestResp)
				}
			}
		}(i)
	}

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

			// fmt.Println(" strings.HasSuffix(searchDir, suffix) -",
			// 	strings.HasSuffix(searchDir, suffix))

			if strings.HasSuffix(searchDir, suffix) {
				fileList = append(fileList, path)
			} else {

				if strings.HasSuffix(path, suffix) {

					//fmt.Println(path)

					wd.taskChan <- path // this is the dispatcher

					fileList = append(fileList, path)
					//------ ingest file content to elasticsearch ------//

					// data := csvLine{}
					// data.topic = line[0]
					// data.source = path //--- file name listed in cvs

					// ingestResp := indexFile(es, data)

					// fmt.Println("Ingest response:", ingestResp)
				}
			}
			return nil
		})
		if err != nil {
			return nil, utils.Error(err)
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
		res.err = utils.Error(err)
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
			return "", utils.Error(fmt.Errorf("Error while scraping html, %v", err))
		}

		content := Content{Topic: csvLine.topic, Content: string(byteContents), Source: csvLine.source}

		msg, err := addToIndex(client, csvLine.topic, content)
		if err != nil {
			return "", utils.Error(fmt.Errorf("Error on add http document to index, %v", err))
		}
		return msg, nil

	} else {
		byteContents, err := ioutil.ReadFile(csvLine.source)
		if err != nil {
			//fmt.Printf("%v\n", err)
			return "", utils.Error(fmt.Errorf("Error on reading file: %v", err))
		}

		content := Content{Topic: csvLine.topic, Content: string(byteContents), Source: csvLine.source}

		msg, err := addToIndex(client, csvLine.topic, content)
		if err != nil {
			return "", utils.Error(fmt.Errorf("Error on add file document to index, %v", err))
		}

		return msg, nil
	}

}

func addToIndex(client *elastic.Client, topic string, content Content) (string, error) {

	//time.Sleep(100 * time.Millisecond)

	//fmt.Println("Add to index:", content.Source)

	dataJSON, err := json.Marshal(content)
	if err != nil {
		return "", utils.Error(fmt.Errorf("Error on content marshalling, %v", err))
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
		return "", utils.Error(fmt.Errorf("Error on es indexing, %v", err))
	}
	return fmt.Sprintf("Added to index: %s", content.Source), nil
}

func scrapeHtml(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", utils.Error(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", utils.Error(fmt.Errorf("status code error: %d %s %s", res.StatusCode, res.Status, url))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(utils.Error(err))
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
