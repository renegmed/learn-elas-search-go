package explore

import (
	"fmt"
	"testing"
)

// $ go test -run=TestBarrier/Correct_processing -v >> results.txt

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

}
