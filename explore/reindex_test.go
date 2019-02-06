package explore

import (
	"fmt"
	"testing"
)

// $ go test -run=TestBarrier/Correct_endpoints -v
// $ go test -run=TestBarrier/One_endpoint_incorrect -v
func TestBarrier(t *testing.T) {

	t.Run("Correct processing", func(t *testing.T) {
		f := "./index_file.csv"
		result, err := captureBarrierOutput(f, ".go")
		if err != nil {
			fmt.Printf("+++Failed:\n%v\n", err)
			t.Fail()
		}
		// if !strings.Contains(result, "Accept-Encoding") || !strings.Contains(result, "user-agent") {
		// 	t.Fail()
		// }
		t.Log(result)
	})

	// t.Run("One endpoint incorrect", func(t *testing.T) {
	// 	endpoints := []string{"http://malformed-url",
	// 		"http://httpbin.org/User-Agent"}

	// 	result := captureBarrierOutput(endpoints...)
	// 	if !strings.Contains(result, "ERROR") {
	// 		t.Fail()
	// 	}

	// 	t.Log(result)
	// })

	// t.Run("Very short timeout", func(t *testing.T) {
	// 	endpoints := []string{"http://httpbin.org/headers",
	// 		"http://httpbin.org/User-Agent"}
	// 	timeoutMilliseconds = 1

	// 	result := captureBarrierOutput(endpoints...)
	// 	if !strings.Contains(result, "Timeout") {
	// 		t.Fail()
	// 	}

	// 	t.Log(result)
	// })

}
