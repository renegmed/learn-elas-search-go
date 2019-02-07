
build: 
	if [ -a ./web.exe ]; then  rm ./web.exe; fi;   # remove main if it exists 
	go build -o ./web.exe
	./web.exe

hanap:  
	go install ./cmd/hanap
	