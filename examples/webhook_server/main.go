package main

import (
	"log"
	"net/http"

	"github.com/yannickRafael/go-emola/pkg/webhook"
)

func main() {
	// Define the HTTP handler that Movitel will ping
	http.HandleFunc("/emola/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received a callback from e-Mola...")

		// Step 1: Parse the incoming JSON callback Request
		cbReq, err := webhook.ParseCallback(r)
		if err != nil {
			log.Printf("Error parsing callback: %v\n", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Step 2: Handle the transaction result
		log.Printf("\n--- CALLBACK RECEIVED ---")
		log.Printf("Request ID: %s", cbReq.RequestID)
		log.Printf("Trans ID:   %s", cbReq.TransID)
		log.Printf("Code:       %s", cbReq.ErrorCode)
		log.Printf("Message:    %s\n", cbReq.Message)

		if cbReq.ErrorCode == "0" {
			log.Println("✅ The customer successfully entered their PIN and paid!")
			// Here you would find the order in your database using cbReq.TransID or cbReq.RefNo
			// and mark it as PAID.
		} else if cbReq.ErrorCode == "11" {
			log.Println("⚠️ The customer timed out and did not enter their PIN.")
			// Mark order as failed/retry.
		} else {
			log.Printf("❌ Payment failed for another reason (code %s).\n", cbReq.ErrorCode)
		}

		// Step 3: Tell Movitel's server we safely received the data
		// If you don't respond, Movitel's server might keep retrying the callback!
		if err := webhook.AcknowledgeCallback(w); err != nil {
			log.Printf("Error acknowledging callback: %v\n", err)
		}
	})

	// Start a simple server on port 8080
	log.Println("Listening for e-Mola webhooks on http://localhost:8080/emola/callback")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
