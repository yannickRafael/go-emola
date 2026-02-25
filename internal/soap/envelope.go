package soap

import "encoding/xml"

// Envelope represents the root <soapenv:Envelope> tag.
type Envelope struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	Xmlns   string   `xml:"xmlns:soapenv,attr"`
	Web     string   `xml:"xmlns:web,attr"`
	Header  *Header  `xml:"soapenv:Header,omitempty"`
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
	Username string `xml:"username"`
	Password string `xml:"password"`
	Wscode   string `xml:"wscode"`
	Param    string `xml:"param"`
	RawData  string `xml:"rawData"`
}

// ResponseEnvelope represents the envelope returned by Movitel.
type ResponseEnvelope struct {
	XMLName xml.Name     `xml:"Envelope"`
	Body    ResponseBody `xml:"Body"`
}

type ResponseBody struct {
	GwOperationResp GwOperationResponse `xml:"gwOperationResponse"`
}

// GwOperationResponse contains the Result tag from Movitel.
type GwOperationResponse struct {
	Result string `xml:"Result"`
}

// DetailResponse represents the parsed content of the <return> tag.
type DetailResponse struct {
	ErrorCode       string `xml:"errorCode"`
	Message         string `xml:"message"`
	TransID         string `xml:"transId"`
	Balance         string `xml:"balance"`
	OrgResponseCode string `xml:"orgResponseCode"`
}

// NewEnvelope creates a properly formatted SOAP Envelope.
func NewEnvelope(username, password, wscode, paramXML string) *Envelope {
	return &Envelope{
		Xmlns: "http://schemas.xmlsoap.org/soap/envelope/",
		Web:   "http://webservice.bccsgw.viettel.com/", // Correct namespace from WSDL
		Body: Body{
			GwOperation: GwOperation{
				Username: username,
				Password: password,
				Wscode:   wscode,
				Param:    paramXML,
				RawData:  "",
			},
		},
	}
}
