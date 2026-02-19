package payment

// Request represents the payload required to initiate a C2B payment.
type Request struct {
	Phone     string // Customer MSISDN (e.g., "861234567")
	Amount    string // Transaction Amount (e.g., "500")
	Reference string // Unique request ID or Order ID
	Content   string // Optional SMS content/remark shown to user
}

// Response contains the result from Movitel after sending the push request.
type Response struct {
	TransID   string // The transaction ID
	ErrorCode string // "0" = Success, "11" = Timeout, "22" = Async Processing
	Message   string // Detailed message from the API
}
