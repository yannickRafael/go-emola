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

// InnerEnvelope represents the full inner SOAP envelope found inside <original>.
// The structure is:
//
//	<S:Envelope>
//	  <S:Body>
//	    <ns2:pushUssdMessageResponse>
//	      <return><errorCode>22</errorCode>...</return>
//	    </ns2:pushUssdMessageResponse>
//	  </S:Body>
//	</S:Envelope>
type InnerEnvelope struct {
	XMLName xml.Name  `xml:"Envelope"`
	Body    InnerBody `xml:"Body"`
}

type InnerBody struct {
	Response InnerResponse `xml:"pushUssdMessageResponse"`
}

type InnerResponse struct {
	Return DetailResponseContent `xml:"return"`
}

// DetailResponse is the final parsed result returned to the caller.
type DetailResponse struct {
	Return DetailResponseContent
}

// DetailResponseContent holds the actual transaction-level fields.
// NOTE: The API has a typo - it sends 'reqeustId' (not 'requestId').
type DetailResponseContent struct {
	ErrorCode string `xml:"errorCode"` // "0" = Success, "22" = Async, "11" = Timeout, etc.
	Message   string `xml:"message"`
	RequestID string `xml:"reqeustId"` // Intentional API typo: reqeustId
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
