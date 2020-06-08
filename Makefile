
init-project:
	go mod init github.com/renegmed/learn-elas-search-go
.PHONY: init-project

up:
	docker-compose up --build -d 
.PHONY: up

build-web: 
	if [ -a ./web.exe ]; then  rm ./web.exe; fi;   # remove main if it exists 
	go build -o ./web.exe
	./web.exe
.PHONY: build-web 

.PHONY: hanap
hanap:  
	go install ./cmd/hanap
	
.PHONY: golang
golang:  
	hanap client destroy index -i golang
	hanap client reindex file -f ./index_file_go.csv -i golang -s .go

.PHONY: web
web:
	hanap client destroy index -i web
	hanap client reindex file -f ./index_file_web.csv -i web -s web

.PHONY: gopackage
gopackage:
	hanap client destroy index -i gopackage
	hanap client reindex file -f ./index_file_go_src.csv -i gopackage -s .go

.PHONY: solidity
solidity:
	hanap client destroy index -i solidity
	hanap client reindex file -f ./index_file_solidity.csv -i solidity -s .sol

.PHONY: rust
rust:
	hanap client destroy index -i rust
	hanap client reindex file -f ./index_file_rust.csv -i rust -s .rs

.PHONY: pdf
pdf:
	hanap client destroy index -i pdf
	hanap client reindex file -f ./index_file_pdf.csv -i pdf -s .pdf

.PHONY: note
note:
	hanap client destroy index -i note
	hanap client reindex file -f ./index_file_note.csv -i note -s .txt	
	hanap client reindex file -f ./index_file_note.csv -i note -s .md

.PHONY: kubernetes
kubernetes:
	hanap client destroy index -i kubernetes
	hanap client reindex file -f ./index_file_kubernetes.csv -i kubernetes -s .yml	
	hanap client reindex file -f ./index_file_kubernetes.csv -i kubernetes -s .yaml