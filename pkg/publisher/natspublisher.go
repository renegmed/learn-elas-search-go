package publisher

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"log"

	nats "github.com/nats-io/nats.go"
	uuid "github.com/satori/go.uuid"
)

type Message struct {
	MessageID string `json:"message-id"`
	Topic     string `json:"topic"`
	Message   string `json:"message"`
}

type NatsPublisher struct {
	js nats.JetStreamContext
}

func NewNatsPublisher(js nats.JetStreamContext) *NatsPublisher {
	return &NatsPublisher{js}
}
func (s *NatsPublisher) HandlePublishMessage(rw http.ResponseWriter, req *http.Request) {

	log.Println("Request method:", req.Method)
	switch req.Method {
	case "POST":
		s.publishMessage(rw, req)
	default:
		log.Printf("Invalid request method: %s", req.Method)
		http.Error(rw, "Invalid request", http.StatusBadRequest)
	}
}

func (s *NatsPublisher) publishMessage(rw http.ResponseWriter, req *http.Request) {

	var msg Message

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "Bad Request", http.StatusBadRequest)
		return
	}

	log.Println("...Request Body:", string(body))

	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Failed to read request: %v", err)
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		return
	}

	messageID := uuid.NewV4().String()
	msg.MessageID = messageID

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message topic %s\n\t%v", msg.Topic, err)
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}
	_, err = s.js.Publish(msg.Topic, msgJSON)
	if err != nil {
		log.Printf("Failed to publish message onto queue '%s': %v", msg.Topic, err)
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}
}
