package util

type CsvLine struct {
	Topic  string
	Source string
}

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

func check(e error) {
	if e != nil {
		panic(e)
	}
}
