
init-project:
	go mod init github.com/renegmed/learn-elas-search-go
 
build:
	go build -o ingester . 


destroy:
	./ingester client destroy index -i golang

ingest_golang:
	./ingester client reindex file -f ../test-data/index_file_go.csv -i golang -s .go

search_test:
	curl -XGET 'localhost:9200/golang/_search?q=TestAdd&pretty'