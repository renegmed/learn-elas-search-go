package searcher

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	elastic "github.com/olivere/elastic/v7"
)

func (s *searcher) Search(index, phrase string, searchMethod string) ([]string, error) {

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
