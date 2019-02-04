package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"

	"github.com/olivere/elastic"
)

const FIELD_TO_SEARCH = "topic"
const PHRASE_TO_SEARCH = "line"

type CsvLine struct {
	topic  string
	source string
}

type Content struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
	Source  string `json:"source"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func addToIndex(client *elastic.Client, id string, content Content) error {
	_, err := client.Index().
		Index("line").
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

func main() {

	file, err := os.Open("index_file.csv")
	check(err)
	defer file.Close()

	lines, err := csv.NewReader(file).ReadAll()
	check(err)

	csvFiles := make([]CsvLine, len(lines)-1)

	for _, line := range lines[1:] { // skip first line
		//data := CsvLine{fmt.Sprintf("\"%v\"", line[0]), fmt.Sprintf("\"%v\"", line[1])}
		data := new(CsvLine)
		data.topic = line[0]
		data.source = line[1]
		csvFiles = append(csvFiles, *data)
	}

	// read each file and index their contents(lines)

	content, err := ioutil.ReadFile("/Users/rene/learn/go-workspace/src/build-microservice-go/chapter3/main.go")
	check(err)
	//defer file.Close()
	//fmt.Println(string(content))

	// var fileLines []string

	// scanner := bufio.NewScanner(content)

	// scanner.Split(bufio.ScanLines)

	// for scanner.Scan() {
	// 	fileLines = append(fileLines, scanner.Text())
	// }
	// fmt.Println(fileLines)
	// }

	// Create a client
	client, err := elastic.NewClient()
	check(err)

	// Create an index
	_, err = client.CreateIndex("golang").Do(context.Background())
	check(err)

	content2 := Content{Topic: "golang", Content: string(content), Source: "/Users/rene/learn/go-workspace/src/build-microservice-go/chapter3/main.go"}

	counter := 1

	addToIndex(client, strconv.Itoa(counter), content2)

	// Search with a term query

	termQuery := elastic.NewTermQuery("topic", "golang")
	searchResult, err := client.Search().
		Index("golang").  // search in index "tweets"
		Query(termQuery). // specify the query
		//Sort("topic.keyword", true). // sort by "user" field, ascending
		From(0).Size(10).        // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	check(err)

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over iterating the hits, see below.
	var ttyp Content
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if c, ok := item.(Content); ok {
			fmt.Printf("Source  source: %s\n", c.Source)
		}
	}
	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d hits\n", searchResult.TotalHits())

	// // Here's how you iterate through results with full control over each step.
	// if searchResult.Hits.TotalHits > 0 {
	// 	fmt.Printf("Found a total of %d tweets\n", searchResult.Hits.TotalHits)

	// 	// Iterate through results
	// 	for _, hit := range searchResult.Hits.Hits {
	// 		// hit.Index contains the name of the index

	// 		// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
	// 		var t Tweet
	// 		err := json.Unmarshal(*hit.Source, &t)
	// 		if err != nil {
	// 			// Deserialization failed
	// 		}

	// 		// Work with tweet
	// 		fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
	// 	}
	// } else {
	// 	// No hits
	// 	fmt.Print("Found no tweets\n")
	// }

	// Delete the index again
	_, err = client.DeleteIndex("golang").Do(context.Background())
	check(err)
}
