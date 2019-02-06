package util

type CsvLine struct {
	topic  string
	source string
}

type Content struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
	Source  string `json:"source"`
}

const mapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"golang":{
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}
