package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ParseCallback reads the HTTP request sent by e-Mola and decodes
// it into a CallbackRequest struct. It limits the body size to prevent
// resource exhaustion attacks.
func ParseCallback(r *http.Request) (*CallbackRequest, error) {
	if r.Method != http.MethodPost {
		return nil, fmt.Errorf("expected POST method, got %s", r.Method)
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" || contentType != "application/json" {
		// Sometimes gateways might send 'application/json; charset=utf-8'
		// It's safer to just check if it contains the json string if strict parsing fails
		// but for now let's read the body safely anyway.
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	defer r.Body.Close()

	var cbReq CallbackRequest
	if err := json.Unmarshal(body, &cbReq); err != nil {
		return nil, fmt.Errorf("failed to parse JSON callback: %w. Body: %s", err, string(body))
	}

	return &cbReq, nil
}

// AcknowledgeCallback writes the required JSON acknowledgment back to e-Mola
// so they know you successfully received the callback.
func AcknowledgeCallback(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := CallbackResponse{
		ResponseCode: "0",
		Message:      "Success callback", // Standard message
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return fmt.Errorf("failed to encode acknowledgment callback: %w", err)
	}

	return nil
}
