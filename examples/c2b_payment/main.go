package main

import (
	"context"
	"fmt"
	"log"
	"time"

	emola "github.com/coffeebit/go-emola"
	"github.com/coffeebit/go-emola/pkg/config"
	"github.com/coffeebit/go-emola/pkg/payment"
)

func main() {
	// Step 1: Configure the client
	// In a real application, you would load these from environment variables
	cfg := &config.Config{
		Environment: config.EnvUAT, // Use UAT for testing, EnvPROD for live
		PartnerCode: "YOUR_PARTNER_CODE",
		PartnerKey:  "YOUR_PARTNER_KEY",
		Username:    "YOUR_ENCRYPTED_USERNAME",
		Password:    "YOUR_ENCRYPTED_PASSWORD",
		Timeout:     65 * time.Second, // Important: Wait slightly longer than the 60s pin prompt
	}

	// Step 2: Initialize the client
	client, err := emola.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize e-Mola client: %v", err)
	}

	// Step 3: Prepare the payment request
	// Example: Charging 500 MT from phone 861234567
	req := &payment.Request{
		Phone:     "861234567",
		Amount:    "500",
		Reference: fmt.Sprintf("ORD-%d", time.Now().Unix()), // Unique order reference
		Content:   "Payment for Order #123",
	}

	fmt.Printf("Initiating C2B USSD Push for %s MT to %s...\n", req.Amount, req.Phone)
	fmt.Println("Waiting for customer to enter PIN...")

	// Step 4: Call the Payment service
	// This will block for up to 60 seconds while the customer enters their PIN
	ctx := context.Background()
	resp, err := client.Payment().Receive(ctx, req)
	if err != nil {
		log.Fatalf("Payment call failed (Network/VPN error?): %v", err)
	}

	// Step 5: Handle the response
	fmt.Println("\n--- Transaction Complete ---")
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
