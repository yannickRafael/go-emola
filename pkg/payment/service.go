package payment

import (
	"context"
	"fmt"
	"strings"

	"github.com/coffebit/go-emola/internal/soap"
	"github.com/coffebit/go-emola/pkg/config"
)

// SOAPCaller defines the interface the Service requires from the base client.
type SOAPCaller interface {
	Config() *config.Config
	CallSOAP(ctx context.Context, wscode, paramXML string) (*soap.DetailResponse, error)
}

// Service handles Customer-to-Business (C2B) payments via USSD Push.
type Service struct {
	client SOAPCaller
}

// NewService creates a new Payment Service bound to the base client.
func NewService(client SOAPCaller) *Service {
	return &Service{client: client}
}

// Receive initiates the payment request. Depending on your Movitel account
// setup, this will either block until the user enters their PIN (Sync)
// or return immediately with code 22 (Async processing).
func (s *Service) Receive(ctx context.Context, req *Request) (*Response, error) {
	if req.Phone == "" || req.Amount == "" || req.Reference == "" {
		return nil, fmt.Errorf("phone, amount, and reference are required")
	}

	cfg := s.client.Config()
	content := req.Content
	if content == "" {
		content = cfg.PartnerCode // Default to Partner Name if empty
	}

	// Construct the raw XML for the 'param' field
	paramXML := fmt.Sprintf(
		"<msisdn>%s</msisdn><transId>%s</transId><transAmount>%s</transAmount><partnerCode>%s</partnerCode><smsContent>%s</smsContent><language>en</language><key>%s</key><refNo>%s</refNo>",
		escapeXML(req.Phone),
		escapeXML(req.Reference), // Using Reference as TransID
		escapeXML(req.Amount),
		escapeXML(cfg.PartnerCode),
		escapeXML(content),
		escapeXML(cfg.PartnerKey),
		escapeXML(req.Reference),
	)

	// Call the generic SOAP dispatcher
	detail, err := s.client.CallSOAP(ctx, "pushUssdMessage", paramXML)
	if err != nil {
		return nil, fmt.Errorf("failed to execute payment: %w", err)
	}

	return &Response{
		TransID:   detail.TransID,
		ErrorCode: detail.ErrorCode,
		Message:   detail.Message,
	}, nil
}

// escapeXML is a simple helper to prevent XML injection in basic fields.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
