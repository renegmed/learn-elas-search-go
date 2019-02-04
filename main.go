package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/olivere/elastic"
)

type Tweet struct {
	User    string `json:"user"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

func addToIndex(client *elastic.Client, id string, tweet Tweet) error {
	_, err := client.Index().
		Index("tweets").
		Type("doc").
		Id(id).
		BodyJson(tweet).
		Refresh("wait_for").
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

const FIELD_TO_SEARCH = "message"
const PHRASE_TO_SEARCH = "four"

func main() {
	// Create a client
	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
	}

	// Create an index
	_, err = client.CreateIndex("tweets").Do(context.Background())
	if err != nil {
		// Handle error
		panic(err)
	}

	tweets := []Tweet{
		Tweet{User: "olivere", Message: "Take One", Source: "/user/file1.txt"},
		Tweet{User: "olivere", Message: "Take Two", Source: "/user/file2.txt"},
		Tweet{User: "olivere", Message: "Take Three", Source: "/user/file3.txt"},
		Tweet{User: "olivere", Message: "Take Four", Source: "/user/file4.txt"},
		Tweet{User: "olivere", Message: "Take Five", Source: "/user/file5.txt"},
		Tweet{User: "olivere", Message: "Take Six", Source: "/user/file6.txt"},
	}

	counter := 1
	for _, tweet := range tweets {
		addToIndex(client, strconv.Itoa(counter), tweet)
		counter++
	}
	// Search with a term query
	//termQuery := elastic.NewTermQuery("user", "olivere")
	termQuery := elastic.NewTermQuery(FIELD_TO_SEARCH, PHRASE_TO_SEARCH)
	searchResult, err := client.Search().
		Index("tweets").            // search in index "tweets"
		Query(termQuery).           // specify the query
		Sort("user.keyword", true). // sort by "user" field, ascending
		From(0).Size(10).           // take documents 0-9
		Pretty(true).               // pretty print request and response JSON
		Do(context.Background())    // execute
	if err != nil {
		// Handle error
		panic(err)
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over iterating the hits, see below.
	var ttyp Tweet
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(Tweet); ok {
			fmt.Printf("Tweet by %s: %s source: %s\n", t.User, t.Message, t.Source)
		}
	}
	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits.TotalHits > 0 {
		fmt.Printf("Found a total of %d tweets\n", searchResult.Hits.TotalHits)

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// hit.Index contains the name of the index

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var t Tweet
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				// Deserialization failed
			}

			// Work with tweet
			fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
		}
	} else {
		// No hits
		fmt.Print("Found no tweets\n")
	}

	// Delete the index again
	_, err = client.DeleteIndex("tweets").Do(context.Background())
	if err != nil {
		// Handle error
		panic(err)
	}
}
