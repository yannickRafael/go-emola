package soap

import "encoding/xml"

// Envelope represents the root <soapenv:Envelope> tag.
type Envelope struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	Xmlns   string   `xml:"xmlns:soapenv,attr"`
	Web     string   `xml:"xmlns:web,attr"`
	Header  Header   `xml:"soapenv:Header"`
	Body    Body
}

// Header is empty but required by the protocol.
type Header struct {
	XMLName xml.Name `xml:"soapenv:Header"`
}

// Body contains the actual function call.
type Body struct {
	XMLName     xml.Name    `xml:"soapenv:Body"`
	GwOperation GwOperation `xml:"web:gwOperation"`
}

// GwOperation maps to the Movitel BCCS Gateway API format.
// The operation name is gwOperation as declared in the WSDL.
type GwOperation struct {
	Input Input `xml:"Input"`
}

// Input wraps the actual payload.
type Input struct {
	Username string  `xml:"username"`
	Password string  `xml:"password"`
	Wscode   string  `xml:"wscode"`
	Params   []Param `xml:"param"`
	RawData  string  `xml:"rawData"` // Can be empty or "?" depending on the call
}

// Param represents a single key-value request parameter.
type Param struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// ResponseEnvelope represents the envelope returned by Movitel.
type ResponseEnvelope struct {
	XMLName xml.Name     `xml:"Envelope"`
	Body    ResponseBody `xml:"Body"`
}

type ResponseBody struct {
	GwOperationResp GwOperationResponse `xml:"gwOperationResponse"`
}

// GwOperationResponse contains the outer <Result> tag from Movitel.
type GwOperationResponse struct {
	Result ResultPayload `xml:"Result"`
}

// ResultPayload contains the payload inside <Result>
// <error>0</error><return></return><original>... XML string ...</original><gwtransid>...</gwtransid>
type ResultPayload struct {
	Error     string `xml:"error"`
	Return    string `xml:"return"`
	Original  string `xml:"original"`
	GwTransID string `xml:"gwtransid"`
}

// DetailResponse represents the parsed content of the inner result string (from <original>).
// These are the transaction-level fields. The <return> tag is nested *inside* the <original> string payload.
type DetailResponse struct {
	// Let's grab the fields from the unescaped inner XML:
	// <ns2:pushUssdMessageResponse><return><errorCode>14</errorCode> ...
	Return DetailResponseContent `xml:"return"`
}

// DetailResponseContent holds the actual fields inside the inner <return> element.
type DetailResponseContent struct {
	ErrorCode string `xml:"errorCode"` // 0 = Success, 10 = ISDN not in white list, 22 = Async Push message done
	Message   string `xml:"message"`   // Human readable description
	RequestID string `xml:"reqeustId"` // Notice the typo from the API server: "reqeustId" instead of "requestId"!
}

// NewEnvelope creates a properly formatted SOAP Envelope.
func NewEnvelope(username, password, wscode string, params []Param) *Envelope {
	return &Envelope{
		Xmlns:  "http://schemas.xmlsoap.org/soap/envelope/",
		Web:    "http://webservice.bccsgw.viettel.com/",
		Header: Header{},
		Body: Body{
			GwOperation: GwOperation{
				Input: Input{
					Username: username,
					Password: password,
					Wscode:   wscode,
					Params:   params,
					RawData:  "?", // Postman uses "?" or "" for rawData. Using "?" to match withdrawing operations
				},
			},
		},
	}
}
