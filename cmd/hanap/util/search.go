package util

import (
	"context"
	"fmt"
	"reflect"

	"github.com/olivere/elastic"
)

func Search(index, phrase string) error {

	// Create a elasticsearch client
	client, err := elastic.NewClient()
	if err != nil {
		return err
	}

	termQuery := elastic.NewMatchPhraseQuery("content", phrase)
	searchResult, err := client.Search().
		Index(index).     // search in index "tweets"
		Query(termQuery). // specify the query
		//Sort("topic.keyword", true). // sort by "topic" field, ascending
		From(0).Size(10).        // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		return err
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	//fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over iterating the hits, see below.
	var ttyp Content
	fmt.Printf("Phrase: %s\n", phrase)
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if c, ok := item.(Content); ok {
			fmt.Printf("%s\n", c.Source)
		}
	}
	// TotalHits is another convenience function that works even when something goes wrong.
	//fmt.Printf("Found a total of %d hits\n", searchResult.TotalHits())

	return nil
}
