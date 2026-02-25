package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	emola "github.com/yannickRafael/go-emola"
	"github.com/yannickRafael/go-emola/pkg/config"
	"github.com/yannickRafael/go-emola/pkg/payment"
)

func main() {
	// Attempt to load .env file if it exists. Ignore if it doesn't.
	_ = godotenv.Load()

	// Define command-line flags
	phone := flag.String("phone", "", "Customer phone number (e.g., 861234567)")
	amount := flag.String("amount", "", "Amount to charge (e.g., 500)")
	refName := flag.String("ref", fmt.Sprintf("ORD-%d", time.Now().Unix()), "Unique Reference string (default is timestamp-based)")
	content := flag.String("content", "Payment via CLI Tool", "SMS content to show")
	envStr := flag.String("env", "UAT", "Environment: UAT or PROD")
	flag.Parse()

	if *phone == "" || *amount == "" {
		fmt.Println("Error: --phone and --amount are required flags.")
		fmt.Println("Usage: ./c2b_payment --phone 861234567 --amount 103 --ref test1234")
		os.Exit(1)
	}

	env := config.EnvUAT
	if *envStr == "PROD" {
		env = config.EnvPROD
	}

	// Step 1: Configure the client
	// It is highly recommended to pass credentials via ENV vars for compiled tools.
	cfg := &config.Config{
		Environment: env,
		PartnerCode: os.Getenv("EMOLA_PARTNER_CODE"),
		PartnerKey:  os.Getenv("EMOLA_PARTNER_KEY"),
		Username:    os.Getenv("EMOLA_USERNAME"),
		Password:    os.Getenv("EMOLA_PASSWORD"),
		Timeout:     65 * time.Second, // Wait slightly longer than the 60s pin prompt
	}

	// Check if environment variables were actually provided
	if cfg.PartnerCode == "" || cfg.Username == "" {
		log.Fatal("Error: EMOLA_PARTNER_CODE, EMOLA_PARTNER_KEY, EMOLA_USERNAME, and EMOLA_PASSWORD environment variables must be set.")
	}

	// Step 2: Initialize the client
	client, err := emola.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize e-Mola client: %v", err)
	}

	// Step 3: Prepare the payment request
	req := &payment.Request{
		Phone:     *phone,
		Amount:    *amount,
		Reference: *refName, // Dynamically set via CLI flag
		Content:   *content,
		Language:  "pt",
	}

	fmt.Printf("Initiating C2B USSD Push for %s MT to %s (Ref: %s)...\n", req.Amount, req.Phone, req.Reference)
	fmt.Println("Connecting to Movitel Gateway...")
	fmt.Println("Waiting for customer to enter PIN (or e-Mola async response)...")

	// Step 4: Call the Payment service
	// This will block for up to 60 seconds while the customer enters their PIN
	ctx := context.Background()

	startTime := time.Now()
	resp, err := client.Payment().Receive(ctx, req)
	elapsed := time.Since(startTime)

	if err != nil {
		log.Fatalf("\n❌ Payment call failed after %v: %v", elapsed, err)
	}

	// Step 5: Handle the response
	fmt.Printf("\n--- Transaction Complete (Took %v) ---\n", elapsed)
	fmt.Printf("Transaction ID: %s\n", resp.TransID)
	fmt.Printf("Error Code:     %s\n", resp.ErrorCode)
	fmt.Printf("Message:        %s\n", resp.Message)

	if resp.ErrorCode == "0" {
		fmt.Println("\n✅ SUCCESS: The payment was completed successfully!")
	} else if resp.ErrorCode == "11" {
		fmt.Println("\n⚠️ TIMEOUT: The customer did not enter their PIN in time.")
	} else {
		fmt.Printf("\n❌ FAILED: The transaction failed with code %s.\n", resp.ErrorCode)
	}
}
