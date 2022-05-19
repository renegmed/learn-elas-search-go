package ingester

import (
	elastic "github.com/olivere/elastic/v7"
	"github.com/renegmed/learn-elas-search-go/pkg/utils"
)

type Content struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
	Source  string `json:"source"`
}

const Mapping = `
{ 
	"mappings":{
		"properties":{
			"topic":{
				"type":"keyword"
			},
			"content":{
				"type":"text" 
			},
			"source":{
				"type":"text"
			} 
		}
	 }
}`

type searcher struct {
	client *elastic.Client
}

func NewSearcher() (searcher, error) {
	client, err := elastic.NewClient()
	if err != nil {
		return searcher{}, utils.Error(err)
	}
	return searcher{
		client: client,
	}, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
