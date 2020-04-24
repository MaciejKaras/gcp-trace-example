package main

import (
	"encoding/json"
	"github.com/MaciejKaras/gcp-trace/shared"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	http.HandleFunc("/account", account)

	log.Printf("Listening on port %s", port)

	httpServer := &http.Server{Addr: ":" + port}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Error while runing server: %v", err)
	}

	log.Printf("Server exiting")
}

func account(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	log.Printf("Started creating account")

	var request shared.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error while decoding request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accountID := createAccount(request)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(shared.CreateAccountResponse{AccountID: accountID}); err != nil {
		log.Printf("Error while encoding response: %v", err)
	}
}

func createAccount(request shared.CreateAccountRequest) string {
	time.Sleep(time.Millisecond * 200)

	accountID := uuid.New().String()
	log.Printf("Creating account with id %s for user %s", accountID, request.Name)
	return accountID
}
