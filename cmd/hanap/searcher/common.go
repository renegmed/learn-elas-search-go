package searcher

import elastic "github.com/olivere/elastic/v7"

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
		return searcher{}, err
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
