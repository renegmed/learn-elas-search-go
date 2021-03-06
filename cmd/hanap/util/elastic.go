package util

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
	"reflect"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	elastic "github.com/olivere/elastic/v7"
)

type Searcher struct {
	client *elastic.Client
}

type ingestResp struct {
	Resp string
	Err  error
}

func NewSearcher() (Searcher, error) {

	client, err := elastic.NewClient()
	if err != nil {
		return Searcher{}, err
	}
	return Searcher{
		client: client,
	}, nil

}
func (s *Searcher) Search(index, phrase string, searchMethod string) ([]string, error) {

	var termQuery elastic.Query

	if strings.Contains(searchMethod, "prefix") {
		termQuery = elastic.NewMatchPhrasePrefixQuery("content", phrase)
	} else if strings.Contains(searchMethod, "fuzzy") {
		termQuery = elastic.NewFuzzyQuery("content", phrase).Boost(1.5).Fuzziness(2).PrefixLength(0).MaxExpansions(100)
	} else {
		// all words must belong to a document
		words := strings.Split(strings.Trim(phrase, " "), " ")
		fmt.Printf("++++ words: %v\n", words)

		tQuery := elastic.NewBoolQuery()

		for _, word := range words {
			termQuery = tQuery.Must(elastic.NewTermQuery("content", word))
		}
	}

	//termQuery := elastic.NewMatchPhraseQuery("content", phrase)

	searchResult, err := s.client.Search().
		Index(index).     // search in index "tweets"
		Query(termQuery). // specify the query
		//Sort("topic.keyword", true). // sort by "topic" field, ascending
		From(0).Size(2000).      // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		return nil, err
	}

	sourceList := []string{}

	var ttyp Content
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if c, ok := item.(Content); ok {
			sourceList = append(sourceList, c.Source)
		}
	}
	return sourceList, nil
}

func (s *Searcher) Destroy(index string) error {
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

func (s *Searcher) Reindex(file, suffix, index string) (string, error) {

	var fileList []CsvLine
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

	fmt.Println(fileList)

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

	messages := ingestFiles(s.client, fileList)
	var buffer bytes.Buffer
	count := 1
	for _, message := range messages {
		buffer.WriteString(strconv.Itoa(count))
		buffer.WriteString(". ")
		buffer.WriteString(message.Resp)
		buffer.WriteString("\n")
		count++
	}

	return buffer.String(), nil
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

func convertFileList(lines [][]string) []CsvLine {
	csvLines := make([]CsvLine, 0)

	for _, line := range lines[1:] { // skip first line
		data := new(CsvLine)
		data.Topic = line[0]
		data.Source = strings.TrimSpace(line[1])
		csvLines = append(csvLines, *data)
	}
	return csvLines
}
func walkFileList(lines [][]string, suffix string) ([]CsvLine, error) {
	csvLines := make([]CsvLine, 0)

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
			data := new(CsvLine)
			data.Topic = line[0]
			data.Source = file
			csvLines = append(csvLines, *data)
		}
	}
	return csvLines, nil
}

func ingestFiles(client *elastic.Client, files []CsvLine) []ingestResp {
	responses := []ingestResp{}
	for _, file := range files {
		resp := indexFile(client, file)
		responses = append(responses, resp)
	}
	return responses
}

func indexFile(client *elastic.Client, csvLine CsvLine) ingestResp {
	res := ingestResp{}
	err := processIndex(client, csvLine)
	if err != nil {
		res.Err = err
		res.Resp = string(fmt.Sprintf("ERROR file: %s\n %v", csvLine.Source, err))
		return res
	}
	res.Resp = string(fmt.Sprintf("file indexed: %s", csvLine.Source))
	return res
}

func processIndex(client *elastic.Client, csvLine CsvLine) error {

	if strings.HasPrefix(csvLine.Source, "http") {
		byteContents, err := scrapeHtml(csvLine.Source)
		if err != nil {
			fmt.Printf("%v\n", err)
			return err
		}

		content := Content{Topic: csvLine.Topic, Content: string(byteContents), Source: csvLine.Source}

		err = addToIndex(client, csvLine.Topic, content)
		if err != nil {
			return fmt.Errorf("Error on add http document to index, %v", err)
		}

	} else {
		byteContents, err := ioutil.ReadFile(csvLine.Source)
		if err != nil {
			fmt.Printf("%v\n", err)
			return err
		}

		content := Content{Topic: csvLine.Topic, Content: string(byteContents), Source: csvLine.Source}

		err = addToIndex(client, csvLine.Topic, content)
		if err != nil {
			return fmt.Errorf("Error on add file document to index, %v", err)
		}
	}
	return nil
}

func addToIndex(client *elastic.Client, topic string, content Content) error {

	dataJSON, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("Error on content marshalling, %v", err)
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
		return err
	}
	return nil
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
