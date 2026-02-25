package emola

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/yannickRafael/go-emola/internal/soap"
	"github.com/yannickRafael/go-emola/pkg/config"
	"github.com/yannickRafael/go-emola/pkg/payment"
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
func (c *Client) CallSOAP(ctx context.Context, wscode string, params []soap.Param) (*soap.DetailResponse, error) {
	envelope := soap.NewEnvelope(c.config.Username, c.config.Password, wscode, params)

	xmlBytes, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	payload := []byte(xml.Header + string(xmlBytes))

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.URL(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if os.Getenv("EMOLA_VERBOSE") == "true" {
		fmt.Printf("\n--- [VERBOSE OUTGOING XML REQUEST] ---\n%s\n--------------------------------------\n", string(payload))
	}

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("SOAPAction", "#POST") // Must match Postman: SOAPAction: #POST

	if os.Getenv("EMOLA_VERBOSE") == "true" {
		fmt.Println("[VERBOSE] Sending HTTP request to:", c.config.URL())
	}

	resp, err := c.httpClient.Do(req)
	if os.Getenv("EMOLA_VERBOSE") == "true" {
		if err != nil {
			fmt.Printf("[VERBOSE] HTTP call returned an error: %v\n", err)
		} else {
			fmt.Printf("[VERBOSE] HTTP response received. Status: %s\n", resp.Status)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if os.Getenv("EMOLA_VERBOSE") == "true" {
		fmt.Printf("\n--- [VERBOSE RAW XML RESPONSE FROM MOVITEL] ---\n%s\n-----------------------------------------------\n", string(bodyBytes))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	// First layer: parse the SOAP envelope into our structures
	var soapResp soap.ResponseEnvelope
	if err := xml.Unmarshal(bodyBytes, &soapResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SOAP response: %w", err)
	}

	resultObj := soapResp.Body.GwOperationResp.Result

	// Check gateway-level error (e.g., login failed)
	if resultObj.Error != "0" {
		return nil, fmt.Errorf("gateway rejected request (errorCode: %s)", resultObj.Error)
	}

	if resultObj.Original == "" {
		return nil, fmt.Errorf("empty <original> payload in SOAP response despite gateway success")
	}

	// Second layer: unmarshal the inner XML found inside <original>
	var detail soap.DetailResponse
	if err := xml.Unmarshal([]byte(resultObj.Original), &detail); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inner detail response: %w", err)
	}

	return &detail, nil
}
