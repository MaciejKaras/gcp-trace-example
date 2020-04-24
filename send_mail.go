package gcp_trace_example

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MaciejKaras/gcp-trace/shared"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type PubSubRequest struct {
	Message struct {
		MessageID  string            `json:"messageId"`
		Attributes map[string]string `json:"attributes"`
		Data       string            `json:"data"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func SendMail(w http.ResponseWriter, r *http.Request) {
	var pubSubRequest PubSubRequest
	if err := json.NewDecoder(r.Body).Decode(&pubSubRequest); err != nil {
		log.Printf("error decoding message: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("PubSubRequest: %+v", pubSubRequest)

	data, err := base64.StdEncoding.DecodeString(pubSubRequest.Message.Data)
	if err != nil {
		log.Printf("error decoding base64 data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Data: %+v", data)

	var mailMessage shared.MailMessage
	if err := json.Unmarshal(data, &mailMessage); err != nil {
		log.Printf("error decoding message send mail data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("SendMailData: %+v", mailMessage)

	if err := sendMail(mailMessage); err != nil {
		log.Printf("error while sending mail to: %s: %v", mailMessage.Recipient, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func sendMail(message shared.MailMessage) error {
	log.Printf("Sending email to %s", message.Recipient)

	//Imitate request latency
	time.Sleep(time.Millisecond * 250)

	if !checkIfSuccess() {
		return fmt.Errorf("sms client error")
	}

	return nil
}

func checkIfSuccess() bool {
	//Error rate (percent)
	const errorRate = 30

	rand.Seed(time.Now().UnixNano())

	return rand.Intn(100)%100 >= errorRate
}
