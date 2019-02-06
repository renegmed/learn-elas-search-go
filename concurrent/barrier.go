//
// Barrier Concurrency Pattern
// Purpose: put up a barrier so that nobody passes until we have all the results we need
//
package barrier

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var timeoutMilliseconds int = 5000

type barrierResp struct {
	Resp string
	Err  error
}

// capture the output from std output
func captureBarrierOutput(endpoints ...string) string {
	reader, writer, _ := os.Pipe() // Pipe() connects reader ouput into writer input thus reader and writer acts as one

	os.Stdout = writer // make the writer as output handler

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		fmt.Printf("++++ buf: %v", &buf) // e.g. ++++ buf: ERROR:  Get http://malformed-url: dial tcp: lookup malformed-url: no such host
		io.Copy(&buf, reader)
		outChan <- buf.String() // send read string to channel
	}()

	barrier(endpoints...)

	writer.Close()
	temp := <-outChan // waits all the results here - messages from outChan are all in.

	return temp
}

func barrier(endpoints ...string) {
	requestNumber := len(endpoints)

	in := make(chan barrierResp, requestNumber)
	defer close(in)

	responses := make([]barrierResp, requestNumber) // each endpoint has its own response

	for _, endpoint := range endpoints {
		go makeRequest(in, endpoint) // call each enpoint and put into channel the response
	}

	var hasError bool
	for i := 0; i < requestNumber; i++ {
		resp := <-in // resp is a barrierResp
		if resp.Err != nil {
			fmt.Println("ERROR: ", resp.Err)
			hasError = true
		}
		responses[i] = resp
	}

	if !hasError {
		for _, resp := range responses {
			fmt.Println(resp.Resp)
		}
	}
}

// Make http request and process the response/error
func makeRequest(out chan<- barrierResp, url string) { // sending channel
	res := barrierResp{}
	client := http.Client{
		Timeout: time.Duration(time.Duration(timeoutMilliseconds) * time.Millisecond),
	}

	resp, err := client.Get(url)
	if err != nil {
		res.Err = err
		out <- res
		return
	}

	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		res.Err = err
		out <- res
		return
	}

	res.Resp = string(byt)
	out <- res
}
