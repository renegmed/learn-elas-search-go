package publisher

import "net/http"

type Publisher interface {
	HandlePublishMessage(rw http.ResponseWriter, req *http.Request)
}
