package webhook

// CallbackRequest represents the JSON payload pushed by the e-Mola server
// to our callback URL when an asynchronous transaction (like USSD Push) completes.
type CallbackRequest struct {
	RequestID string `json:"reqeustId"` // Notice the typo from the e-Mola API: "reqeustId" instead of "requestId"
	TransID   string `json:"transId"`   // The original transaction ID you passed in
	RefNo     string `json:"refNo"`     // The reference number
	ErrorCode string `json:"errorCode"` // "0" for success, "11" for timeout, etc.
	Message   string `json:"message"`   // Detailed description
}

// CallbackResponse represents the JSON we should reply to e-Mola with
// to acknowledge we received their callback successfully.
type CallbackResponse struct {
	ResponseCode string `json:"responseCode"` // "0" means successfully received
	Message      string `json:"message"`
}
