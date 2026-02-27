package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	emola "github.com/yannickRafael/go-emola"
	"github.com/yannickRafael/go-emola/pkg/config"
	"github.com/yannickRafael/go-emola/pkg/payment"
	"github.com/yannickRafael/go-emola/pkg/webhook"
)

func main() {
	// Attempt to load .env file if it exists. Ignore if it doesn't.
	_ = godotenv.Load()

	// Define command-line flags
	phone := flag.String("phone", "", "Customer phone number (e.g., 861234567)")
	amount := flag.String("amount", "", "Amount to charge (e.g., 500)")
	// e-Mola API Specification states transId minimum is 15 chars, and refNo is exactly 20 chars.
	// E.g., "ORD" (3) + Timestamp (10) + Random (7) = 20 chars
	defaultRef := fmt.Sprintf("ORD%d%07d", time.Now().Unix(), time.Now().Nanosecond()%10000000)
	refName := flag.String("ref", defaultRef, "Unique Reference string (should be exactly 20 chars)")
	content := flag.String("content", "Payment via CLI Tool", "SMS content to show")
	envStr := flag.String("env", "UAT", "Environment: UAT or PROD")
	verbose := flag.Bool("verbose", false, "Print raw HTTP request and response for debugging")
	callbackPort := flag.String("callback-port", "", "If set, starts a local webhook listener on this port (e.g. 8080)")
	flag.Parse()

	if *phone == "" || *amount == "" {
		fmt.Println("Error: --phone and --amount are required flags.")
		fmt.Println("Usage: ./emola-cli --phone 861234567 --amount 103 --ref test1234 [--callback-port 8080]")
		os.Exit(1)
	}

	env := config.EnvUAT
	if *envStr == "PROD" {
		env = config.EnvPROD
	}

	// Step 1: Configure the client
	cfg := &config.Config{
		Environment: env,
		PartnerCode: os.Getenv("EMOLA_PARTNER_CODE"),
		PartnerKey:  os.Getenv("EMOLA_PARTNER_KEY"),
		Username:    os.Getenv("EMOLA_USERNAME"),
		Password:    os.Getenv("EMOLA_PASSWORD"),
		Timeout:     65 * time.Second,
	}

	if cfg.PartnerCode == "" || cfg.Username == "" {
		log.Fatal("Error: EMOLA_PARTNER_CODE, EMOLA_PARTNER_KEY, EMOLA_USERNAME, and EMOLA_PASSWORD must be set in .env or environment.")
	}

	if *verbose {
		fmt.Println("[DEBUG] Verbose mode enabled")
		os.Setenv("EMOLA_VERBOSE", "true")
	}

	// Step 2: (Optional) Start webhook listener in a background goroutine
	// This lets a single binary both send the request and receive the callback.
	if *callbackPort != "" {
		callbackReceived := make(chan *webhook.CallbackRequest, 1)

		http.HandleFunc("/emola/callback", func(w http.ResponseWriter, r *http.Request) {
			cb, err := webhook.ParseCallback(r)
			if err != nil {
				log.Printf("[WEBHOOK] Failed to parse callback: %v\n", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			webhook.AcknowledgeCallback(w)
			callbackReceived <- cb
		})

		go func() {
			fmt.Printf("[WEBHOOK] Listening for callback on :%s/emola/callback\n", *callbackPort)
			if err := http.ListenAndServe(":"+*callbackPort, nil); err != nil {
				log.Fatalf("[WEBHOOK] Server error: %v", err)
			}
		}()

		// Give the server a moment to start
		time.Sleep(100 * time.Millisecond)

		// Step 3: Prepare and send the payment request
		req := &payment.Request{
			Phone:     *phone,
			Amount:    *amount,
			Reference: *refName,
			Content:   *content,
			Language:  "pt",
		}

		fmt.Printf("[→] Sending C2B request: %s MT to %s (Ref: %s)...\n", req.Amount, req.Phone, req.Reference)

		client, err := emola.NewClient(cfg)
		if err != nil {
			log.Fatalf("Failed to initialize e-Mola client: %v", err)
		}

		ctx := context.Background()
		start := time.Now()
		resp, err := client.Payment().Receive(ctx, req)
		elapsed := time.Since(start)

		if err != nil {
			log.Fatalf("\n❌ Payment call failed after %v: %v", elapsed, err)
		}

		fmt.Printf("[✓] Gateway responded in %v\n", elapsed)
		fmt.Printf("    Error Code: %s\n", resp.ErrorCode)
		fmt.Printf("    Message:    %s\n", resp.Message)
		fmt.Printf("    Request ID: %s\n\n", resp.RequestID)

		if resp.ErrorCode == "0" {
			fmt.Println("✅ SYNC SUCCESS: Payment completed immediately!")
			return
		} else if resp.ErrorCode == "22" {
			fmt.Printf("[⏳] Async mode: Waiting for customer to enter PIN on phone %s...\n", req.Phone)
			fmt.Printf("     (Will wait up to 2 minutes. Press Ctrl+C to abort.)\n\n")
		} else {
			fmt.Printf("❌ Payment rejected by gateway (code %s). Exiting.\n", resp.ErrorCode)
			return
		}

		// Step 4: Wait for the callback and print the result
		select {
		case cb := <-callbackReceived:
			fmt.Println("\n🔔 CALLBACK RECEIVED FROM MOVITEL!")
			fmt.Printf("    Request ID: %s\n", cb.RequestID)
			fmt.Printf("    Trans ID:   %s\n", cb.TransID)
			fmt.Printf("    Error Code: %s\n", cb.ErrorCode)
			fmt.Printf("    Message:    %s\n", cb.Message)

			if cb.ErrorCode == "0" {
				fmt.Println("\n✅ SUCCESS: The customer paid successfully!")
			} else if cb.ErrorCode == "11" {
				fmt.Println("\n⚠️  TIMEOUT: The customer did not enter their PIN in time.")
			} else {
				fmt.Printf("\n❌ Payment failed. Code: %s\n", cb.ErrorCode)
			}

		case <-time.After(2 * time.Minute):
			fmt.Println("\n⏰ No callback received within 2 minutes. Exiting.")
		}

	} else {
		// No callback port — simple fire-and-forget mode (just shows the gateway's immediate response)
		client, err := emola.NewClient(cfg)
		if err != nil {
			log.Fatalf("Failed to initialize e-Mola client: %v", err)
		}

		req := &payment.Request{
			Phone:     *phone,
			Amount:    *amount,
			Reference: *refName,
			Content:   *content,
			Language:  "pt",
		}

		fmt.Printf("[→] Sending C2B request: %s MT to %s (Ref: %s)...\n", req.Amount, req.Phone, req.Reference)

		ctx := context.Background()
		start := time.Now()
		resp, err := client.Payment().Receive(ctx, req)
		elapsed := time.Since(start)

		if err != nil {
			log.Fatalf("\n❌ Payment call failed after %v: %v", elapsed, err)
		}

		fmt.Printf("[✓] Gateway responded in %v\n", elapsed)
		fmt.Println("\n--- Transaction Result ---")
		fmt.Printf("    Trans ID:   %s\n", resp.TransID)
		fmt.Printf("    Request ID: %s\n", resp.RequestID)
		fmt.Printf("    Error Code: %s\n", resp.ErrorCode)
		fmt.Printf("    Message:    %s\n", resp.Message)

		if resp.ErrorCode == "0" {
			fmt.Println("\n✅ SUCCESS: Payment completed immediately!")
		} else if resp.ErrorCode == "22" {
			fmt.Println("\n⏳ ASYNC: Request submitted. Customer will receive a USSD prompt.")
			fmt.Println("   Run with --callback-port 8080 to receive the final result via webhook.")
		} else if resp.ErrorCode == "11" {
			fmt.Println("\n⚠️  TIMEOUT: The customer did not enter their PIN in time.")
		} else {
			fmt.Printf("\n❌ FAILED (code %s)\n", resp.ErrorCode)
		}
	}
}
