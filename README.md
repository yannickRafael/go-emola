# go-emola

A modern, type-safe Go client for the Movitel e-Mola (BCCS Gateway) API.

[![Go Reference](https://pkg.go.dev/badge/github.com/yannickRafael/go-emola.svg)](https://pkg.go.dev/github.com/yannickRafael/go-emola)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🚀 Overview

Integrating with Movitel's e-Mola API usually involves a "labirinto de XMLs", complex VPN routing, and legacy SOAP protocols. `go-emola` abstracts all that complexity away, providing a clean, idiomatic Go SDK.

- **Zero XML:** Never write a line of SOAP/XML again.
- **Type-Safe:** Work with native Go structs and clear error codes.
- **Webhook Support:** Built-in parser for asynchronous payment callbacks.
- **Environment Aware:** Toggle between UAT and Production effortlessly via environment variables.

## 📦 Installation

```bash
go get github.com/yannickRafael/go-emola
```

## 🛠️ Usage

### 1. Configure the Client

```go
import (
    "github.com/yannickRafael/go-emola"
    "github.com/yannickRafael/go-emola/pkg/config"
)

cfg := &config.Config{
    Environment: config.EnvPROD, // OR config.EnvUAT
    PartnerCode: "...",
    PartnerKey:  "...",
    Username:    "...",
    Password:    "...",
}

client, err := emola.NewClient(cfg)
```

### 2. Request a Payment (C2B)

```go
resp, err := client.Payment().Receive(ctx, &payment.Request{
    Phone:     "861234567",
    Amount:    "100",
    Reference: "ORDER-123", // Library enforces valid length automatically
})

if resp.ErrorCode == "22" {
    fmt.Println("Waiting for user USSD PIN confirmation...")
}
```

### 3. Handle Webhooks (Callbacks)

```go
http.HandleFunc("/emola/callback", func(w http.ResponseWriter, r *http.Request) {
    callback, err := webhook.ParseCallback(r)
    if err == nil && callback.ErrorCode == "0" {
        fmt.Printf("Payment for %s successful!\n", callback.TransID)
        webhook.AcknowledgeCallback(w)
    }
})
```

## 🧪 Testing with the CLI Tool

We include a ready-to-use CLI tool for rapid testing in the `examples/c2b_payment` directory.

1. `cp examples/c2b_payment/.env.example .env`
2. Fill in your credentials.
3. `go run examples/c2b_payment/main.go --phone 86... --amount 10 --verbose`

## ⚖️ License

Distributed under the MIT License. See `LICENSE` for more information.
