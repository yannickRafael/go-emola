package payment

import (
	"context"
	"fmt"

	"github.com/yannickRafael/go-emola/internal/soap"
	"github.com/yannickRafael/go-emola/pkg/config"
)

// SOAPCaller defines the interface the Service requires from the base client.
type SOAPCaller interface {
	Config() *config.Config
	CallSOAP(ctx context.Context, wscode string, params []soap.Param) (*soap.DetailResponse, error)
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

	lang := req.Language
	if lang == "" {
		lang = "pt" // Default to Portuguese for Mozambique
	}

	// Construct the parameters matching the <param name="x" value="y"/> structure
	params := []soap.Param{
		{Name: "msisdn", Value: req.Phone},
		{Name: "smsContent", Value: content},
		{Name: "transAmount", Value: req.Amount},
		{Name: "transId", Value: req.Reference},
		{Name: "language", Value: lang},
		{Name: "refNo", Value: req.Reference},
		{Name: "partnerCode", Value: cfg.PartnerCode},
		{Name: "key", Value: cfg.PartnerKey},
	}

	// Call the generic SOAP dispatcher
	detail, err := s.client.CallSOAP(ctx, "pushUssdMessage", params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute payment: %w", err)
	}

	return &Response{
		TransID:   req.Reference, // Movitel doesn't echo transId back reliably in this call, so we use our own request ref
		RequestID: detail.Return.RequestID,
		ErrorCode: detail.Return.ErrorCode,
		Message:   detail.Return.Message,
	}, nil
}

// We no longer need escapeXML here because Go's xml.Marshal handles escaping
// attributes safely automatically when encoding soap.Param structures!
