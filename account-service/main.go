package main

import (
	"context"
	"encoding/json"
	"github.com/MaciejKaras/gcp-trace/shared"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"time"
)

var traceExporter *shared.TraceExporter

func main() {
	var err error
	traceExporter, err = shared.InitTrace()
	if err != nil {
		log.Fatalf("Error initializing trace exporter: %v", err)
	}
	defer traceExporter.Flush()

	http.HandleFunc("/account", account)

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

func account(w http.ResponseWriter, r *http.Request) {
	ctx, span := shared.StartRequestSpan(r)
	defer span.End()

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

	accountID := createAccount(ctx, request)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(shared.CreateAccountResponse{AccountID: accountID}); err != nil {
		log.Printf("Error while encoding response: %v", err)
	}
}

func createAccount(ctx context.Context, request shared.CreateAccountRequest) string {
	ctx, span := shared.StartSpan(ctx, "createAccount")
	defer span.End()

	time.Sleep(time.Millisecond * 200)

	accountID := uuid.New().String()
	log.Printf("Creating account with id %s for user %s", accountID, request.Name)
	return accountID
}
