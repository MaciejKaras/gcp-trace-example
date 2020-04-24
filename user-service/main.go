package main

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/MaciejKaras/gcp-trace/shared"
	"github.com/google/uuid"
	"google.golang.org/appengine"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var sendMailTopic *pubsub.Topic
var httpClient *http.Client
var accountServiceEndpoint string

func main() {
	ctx := context.Background()

	projectID := appengine.AppID(ctx)

	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Error initializing pubsub: %v", err)
	}
	defer pubSubClient.Close()

	sendMailTopic = pubSubClient.Topic("send-mail")
	defer sendMailTopic.Stop()

	httpClient = &http.Client{
		Timeout: time.Second * 30,
	}
	defer httpClient.CloseIdleConnections()

	accountServiceEnv := os.Getenv("ACCOUNT_SERVICE")
	accountServiceEndpoint = strings.Replace(accountServiceEnv, "{PROJECT_ID}", projectID, 1)

	http.HandleFunc("/user", user)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)

	httpServer := &http.Server{Addr: ":" + port}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Error while runing server: %v", err)
	}

	log.Printf("Server exiting")
}

func user(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request shared.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error while decoding request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := sendMailEvent(ctx, request); err != nil {
		log.Printf("Error while sending mail event: %v", err)
	} else {
		log.Print("Event sent to sendMailTopic")
	}

	if accountID, err := sendCreateAccount(ctx, request); err != nil {
		log.Printf("Error creating account: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Printf("Account created with id %s", accountID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(shared.RegisterUserResponse{UserID: accountID}); err != nil {
			log.Printf("Error while encoding response: %v", err)
		}
	}
}

func sendMailEvent(ctx context.Context, request shared.RegisterUserRequest) error {
	message := shared.MailMessage{
		ID:        uuid.New().String(),
		Recipient: request.Email,
		Subject:   fmt.Sprintf("Welcome %s", request.Name),
		Content:   "Some welcome email content, lorem ipsum etc.",
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error while marshaling MailMessage: %v", err)
	}

	if _, err := sendMailTopic.Publish(ctx, &pubsub.Message{Data: data}).Get(ctx); err != nil {
		return fmt.Errorf("error while publishing to sendMailTopic: %v", err)
	}

	return nil
}

func sendCreateAccount(ctx context.Context, userRequest shared.RegisterUserRequest) (string, error) {
	jsonRequest, err := json.Marshal(shared.CreateAccountRequest{
		Name:        userRequest.Name,
		Email:       userRequest.Email,
		AccountType: "USER_TYPE",
	})

	if err != nil {
		return "", fmt.Errorf("json.Marshal: %v", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, accountServiceEndpoint+"/account", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return "", fmt.Errorf("http.NewRequestWithContext: %v", err)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("httpClient.Do: %v", err)
	}

	if response.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("invalid response status: '%s'", response.Status)
	}

	var createAccountResponse shared.CreateAccountResponse
	err = json.NewDecoder(response.Body).Decode(&createAccountResponse)
	if err != nil {
		return "", fmt.Errorf("invalid response body json: '%s'", err)
	}

	return createAccountResponse.AccountID, nil
}
