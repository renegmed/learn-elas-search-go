// Error wrapper for better logging
package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

func Error(err interface{}) error {
	if os.Getenv("DEBUG") == "true" && err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.Printf("ERROR: [%s:%d] %v \n", fn, line, err)

		switch err.(type) {
		case string:
			return fmt.Errorf(err.(string))
		case error:
			return err.(error)
		default:
			return fmt.Errorf("%v", err)
		}
	}

	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}
