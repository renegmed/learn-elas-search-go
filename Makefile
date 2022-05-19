
init-project:
	go mod init github.com/renegmed/learn-elas-search-go
 

up:
	docker-compose up  --build -d  

down:
	docker-compose down 
 



# ----- web ------

build-web: 
	if [ -a ./web.exe ]; then  rm ./web.exe; fi;   # remove main if it exists 
	go build -race -o ./web.exe web/*.go

run-web: build-web
	./web.exe 


# ----- cli ------

build-cli: 
	#if [ -a ./hanap ]; then  rm ./hanap; fi;   # remove main if it exists 
	go build -race -o ./hanap cmd/cli/*.go

run-cli: build-cli
	./hanap
  

# ---- rebuild elastic search topics -----

install_app:  # build and install to $GOPATH/bin, no need to call ./hanap
	go install ./cmd/hanap
	 
index_golang:  
	./hanap client destroy index -i golang
	GOMAXPROCS=4 hanap client reindex file -f ./cvs-files/test-data/index_file_go.csv -i golang -s .go
 	#GOMAXPROCS=4 hanap client reindex file -f ./cvs-files/index_file_go.csv -i golang -s .go
index_web:
	hanap client destroy index -i web
	hanap client reindex file -f ./index_file_web.csv -i web -s web
 
index_gopackage:
	hanap client destroy index -i gopackage
	hanap client reindex file -f ./index_file_go_src.csv -i gopackage -s .go
 
index_solidity:
	hanap client destroy index -i solidity
	hanap client reindex file -f ./index_file_solidity.csv -i solidity -s .sol

 
index_rust:
	hanap client destroy index -i rust
	hanap client reindex file -f ./index_file_rust.csv -i rust -s .rs
 
index_pdf:
	hanap client destroy index -i pdf
	hanap client reindex file -f ./index_file_pdf.csv -i pdf -s .pdf
 
index_note:
	hanap client destroy index -i note
	hanap client reindex file -f ./index_file_note.csv -i note -s .txt	
	hanap client reindex file -f ./index_file_note.csv -i note -s .md
 
index_kubernetes:
	hanap client destroy index -i kubernetes
	hanap client reindex file -f ./index_file_kubernetes.csv -i kubernetes -s .yml	
	hanap client reindex file -f ./index_file_kubernetes.csv -i kubernetes -s .yaml