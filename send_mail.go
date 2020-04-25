package gcp_trace_example

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MaciejKaras/gcp-trace/shared"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var traceExporter *shared.TraceExporter

type PubSubRequest struct {
	Message struct {
		MessageID  string            `json:"messageId"`
		Attributes map[string]string `json:"attributes"`
		Data       string            `json:"data"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func init() {
	var err error
	traceExporter, err = shared.InitTrace()
	if err != nil {
		log.Fatalf("Error initializing trace exporter: %v", err)
	}
}

func SendMail(w http.ResponseWriter, r *http.Request) {
	defer traceExporter.Flush()

	var pubSubRequest PubSubRequest
	if err := json.NewDecoder(r.Body).Decode(&pubSubRequest); err != nil {
		log.Printf("error decoding message: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, span := shared.StartCloudEventSpan(r.Context(), "SendMail", pubSubRequest.Message.Attributes)
	defer span.End()

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

	if err := sendMailRequest(ctx, mailMessage); err != nil {
		log.Printf("error while sending mail to: %s: %v", mailMessage.Recipient, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func sendMailRequest(ctx context.Context, message shared.MailMessage) error {
	ctx, span := shared.StartSpan(ctx, "sendMailRequest")
	defer span.End()

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
