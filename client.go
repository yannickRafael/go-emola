package emola

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/coffebit/go-emola/internal/soap"
	"github.com/coffebit/go-emola/pkg/config"
	"github.com/coffebit/go-emola/pkg/payment"
)

// Client is the main interface for interacting with the e-Mola service.
type Client struct {
	config     *config.Config
	httpClient *http.Client
}

// Payment returns the Customer-to-Business (C2B) service instance.
func (c *Client) Payment() *payment.Service {
	return payment.NewService(c)
}

// NewClient creates a new e-Mola API client with the given configuration.
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}

// Config returns the client's configuration.
func (c *Client) Config() *config.Config {
	return c.config
}

// CallSOAP is the internal method for dispatching a SOAP request.
func (c *Client) CallSOAP(ctx context.Context, wscode, paramXML string) (*soap.DetailResponse, error) {
	envelope := soap.NewEnvelope(c.config.Username, c.config.Password, wscode, paramXML)

	xmlBytes, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	payload := []byte(xml.Header + string(xmlBytes))

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.URL(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var soapResp soap.ResponseEnvelope
	if err := xml.Unmarshal(bodyBytes, &soapResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SOAP response: %w", err)
	}

	// Double-parse the 'return' element which contains nested XML
	var detail soap.DetailResponse
	if soapResp.Body.ProcessResp.Return == "" {
		return nil, fmt.Errorf("empty return data in SOAP response")
	}

	if err := xml.Unmarshal([]byte(soapResp.Body.ProcessResp.Return), &detail); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inner detail response: %w", err)
	}

	return &detail, nil
}
